// Package ble implements the vehicle.Connector interface using BLE.

package ble

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"
	"sync"
	"time"

	// "github.com/go-ble/ble"
	"github.com/rigado/ble"
	"github.com/teslamotors/vehicle-command/internal/log"
	"github.com/teslamotors/vehicle-command/pkg/connector"
	"github.com/teslamotors/vehicle-command/pkg/protocol"
)

const maxBLEMessageSize = 1024

var ErrMaxConnectionsExceeded = protocol.NewError("the vehicle is already connected to the maximum number of BLE devices", false, false)

var (
	rxTimeout  = time.Second     // Timeout interval between receiving chunks of a mesasge
	maxLatency = 4 * time.Second // Max allowed error when syncing vehicle clock
)

var (
	vehicleServiceUUID = ble.MustParse("00000211-b2d1-43f0-9b88-960cebf8b91e")
	toVehicleUUID      = ble.MustParse("00000212-b2d1-43f0-9b88-960cebf8b91e")
	fromVehicleUUID    = ble.MustParse("00000213-b2d1-43f0-9b88-960cebf8b91e")
)

var (
	device ble.Device
	mu     sync.Mutex
)

type Connection struct {
	vin         string
	inbox       chan []byte
	txChar      *ble.Characteristic
	rxChar      *ble.Characteristic
	inputBuffer []byte
	client      ble.Client
	lastRx      time.Time
	lock        sync.Mutex
	device      ble.Device
	rssi        int
}

type ScanList struct {
	scanEntries []ScanResult
	device      ble.Device
}

type ScanResult struct {
	localName   string
	addr        ble.Addr
	manufacData []byte
	services    []ble.UUID
	rssi        int
}

func (scanList *ScanList) ScanEntries() []ScanResult {
	return scanList.scanEntries
}

func (scanEntry *ScanResult) ScanEntryString() string {
	return fmt.Sprintf("name: %s, addr: [%s], MD: %X, Services: %v, RSSI: %3d", scanEntry.localName, scanEntry.addr, scanEntry.manufacData, scanEntry.services, scanEntry.rssi)
}

func (scanEntry *ScanResult) LocalName() string {
	return scanEntry.localName
}

func (scanEntry *ScanResult) RSSI() int {
	return scanEntry.rssi
}

func SetLogLevelTrace() {
	ble.SetLogLevelTrace()
}

func SetLogLevelDebug() {
	ble.SetLogLevelDebug()
}

func SetLogLevelInfo() {
	ble.SetLogLevelInfo()
}

func SetLogLevelWarn() {
	ble.SetLogLevelWarn()
}

func SetLogLevelError() {
	ble.SetLogLevelError()
}

func SetLogLevelFatal() {
	ble.SetLogLevelFatal()
}

func SetLogLevelPanic() {
	ble.SetLogLevelPanic()
}

func (c *Connection) PreferredAuthMethod() connector.AuthMethod {
	return connector.AuthMethodGCM
}

func (c *Connection) RetryInterval() time.Duration {
	return time.Second
}

func (c *Connection) Receive() <-chan []byte {
	return c.inbox
}

func (c *Connection) RSSI() int {
	return c.rssi
}

func (c *Connection) flush() bool {
	if len(c.inputBuffer) >= 2 {
		msgLength := 256*int(c.inputBuffer[0]) + int(c.inputBuffer[1])
		if msgLength > maxBLEMessageSize {
			c.inputBuffer = []byte{}
			return false
		}
		if len(c.inputBuffer) >= 2+msgLength {
			buffer := c.inputBuffer[2 : 2+msgLength]
			log.Debug("RX: %02x", buffer)
			c.inputBuffer = c.inputBuffer[2+msgLength:]
			select {
			case c.inbox <- buffer:
			default:
				return false
			}
			return true
		}
	}
	return false
}

func (c *Connection) Close() {
	c.client.ClearSubscriptions()
	c.client.CancelConnection()
}

func (c *Connection) AllowedLatency() time.Duration {
	return maxLatency
}

func (c *Connection) rx(id uint, p []byte) {
	if time.Since(c.lastRx) > rxTimeout {
		c.inputBuffer = []byte{}
	}
	c.lastRx = time.Now()
	c.inputBuffer = append(c.inputBuffer, p...)
	for c.flush() {
	}
}

func (c *Connection) Send(ctx context.Context, buffer []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	var out []byte
	log.Debug("TX: %02x", buffer)
	out = append(out, uint8(len(buffer)>>8), uint8(len(buffer)))
	out = append(out, buffer...)
	blockLength := 20
	for len(out) > 0 {
		if blockLength > len(out) {
			blockLength = len(out)
		}
		if err := c.client.WriteCharacteristic(c.txChar, out[:blockLength], false); err != nil {
			return err
		}
		out = out[blockLength:]
	}
	return nil
}

func (c *Connection) VIN() string {
	return c.vin
}

func NewConnection(ctx context.Context, vin string) (*Connection, error) {
	var lastError error
	for {
		conn, err := tryToConnect(ctx, vin)
		if err == nil {
			return conn, nil
		}
		if strings.Contains(err.Error(), "operation not permitted") {
			return nil, err
		}
		log.Warning("BLE connection attempt failed: %s", err)
		if err := ctx.Err(); err != nil {
			if lastError != nil {
				return nil, lastError
			}
			return nil, err
		}
		lastError = err
	}
}

func tryToConnect(ctx context.Context, vin string) (*Connection, error) {
	var err, err2 error
	var localName string
	// We don't want concurrent calls to NewConnection that would defeat
	// the point of reusing the existing BLE device. Note that this is not
	// an issue on MacOS, but multiple calls to newDevice() on Linux leads to failures.
	mu.Lock()
	defer mu.Unlock()

	if device != nil {
		log.Debug("Reusing existing BLE device")
	} else {
		log.Debug("Creating new BLE device")
		device, err = newDevice()
		if err != nil {
			return nil, fmt.Errorf("failed to find a BLE device: %s", err)
		}
		ble.SetDefaultDevice(device)
	}
	// check, if localName or VIN was provided
	if strings.HasPrefix(vin, "S") && strings.HasSuffix(vin, "C") {
		localName = vin
	} else {
		// calculate localName from SHA1 checksum of VIN
		vinBytes := []byte(vin)
		digest := sha1.Sum(vinBytes)
		localName = fmt.Sprintf("S%02xC", digest[:8])
	}
	log.Debug("Searching for BLE beacon %s...", localName)
	canConnect := false
	filter := func(adv ble.Advertisement) bool {
		ln := adv.LocalName()
		if len(ln) > 0 {
			log.Debug("Advertisement from Name: %s [%s] RSSI: %3d:", ln, adv.Addr(), adv.RSSI())
		}
		if ln != localName {
			return false
		}
		canConnect = adv.Connectable()
		if canConnect {
			log.Debug("Connectable! Services: %v, MD: %X.", adv.Services(), adv.ManufacturerData())
		}

		return true
	}

	client, err := ble.Connect(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find BLE beacon for %s (%s): %s", vin, localName, err)
	}

	if !canConnect {
		return nil, ErrMaxConnectionsExceeded
	}

	log.Debug("Connecting to BLE beacon %s...", client.Addr())
	services, err := client.DiscoverServices([]ble.UUID{vehicleServiceUUID})
	if err != nil {
		return nil, fmt.Errorf("ble: failed to enumerate device services: %s", err)
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("ble: failed to discover service")
	}

	characteristics, err := client.DiscoverCharacteristics([]ble.UUID{toVehicleUUID, fromVehicleUUID}, services[0])
	if err != nil {
		return nil, fmt.Errorf("ble: failed to discover service characteristics: %s", err)
	}

	conn := Connection{
		vin:    vin,
		client: client,
		inbox:  make(chan []byte, 5),
		device: device,
	}
	for _, characteristic := range characteristics {
		if characteristic.UUID.Equal(toVehicleUUID) {
			conn.txChar = characteristic
		} else if characteristic.UUID.Equal(fromVehicleUUID) {
			conn.rxChar = characteristic
		}
		if _, err := client.DiscoverDescriptors(nil, characteristic); err != nil {
			return nil, fmt.Errorf("ble: couldn't fetch descriptors: %s", err)
		}
	}
	if conn.txChar == nil || conn.rxChar == nil {
		return nil, fmt.Errorf("ble: failed to find required characteristics")
	}
	if err := client.Subscribe(conn.rxChar, true, conn.rx); err != nil {
		return nil, fmt.Errorf("ble: failed to subscribe to RX: %s", err)
	}
	log.Info("Connected to vehicle BLE")
	rssi, err2 := client.ReadRSSI()
	conn.rssi = int(rssi)
	if err2 == nil {
		log.Info("RSSI %ddBm", conn.rssi)
	}
	return &conn, nil
}

func NewScan(ctx context.Context) (*ScanList, error) {
	var lastError error
	for {
		scanList, err := tryToScan(ctx)
		if err == nil {
			return scanList, nil
		}
		if strings.Contains(err.Error(), "operation not permitted") {
			return nil, err
		}
		log.Info("Finished scanning for BLE beacons: %s", err)
		if err := ctx.Err(); err != nil {
			if lastError != nil {
				return scanList, lastError
			}
			return scanList, err
		}
		lastError = err
	}
}

func tryToScan(ctx context.Context) (*ScanList, error) {
	var err error
	// We don't want concurrent calls to NewConnection that would defeat
	// the point of reusing the existing BLE device. Note that this is not
	// an issue on MacOS, but multiple calls to newDevice() on Linux leads to failures.
	mu.Lock()
	defer mu.Unlock()

	if device != nil {
		log.Debug("Reusing existing BLE device")
	} else {
		log.Debug("Creating new BLE device")
		device, err = newDevice()
		if err != nil {
			return nil, fmt.Errorf("failed to find a BLE device: %s", err)
		}
		ble.SetDefaultDevice(device)
	}

	log.Debug("Searching for BLE beacons...")
	scanList := ScanList{
		device: device,
	}
	canConnect := false
	filter := func(adv ble.Advertisement) bool {
		ln := adv.LocalName()
		if len(ln) > 0 {
			log.Debug("Advertisement from Name: %s [%s] RSSI: %3d:", ln, adv.Addr(), adv.RSSI())
		}
		if len(ln) != 18 {
			return false
		}
		if strings.HasPrefix(ln, "S") || strings.HasSuffix(ln, "C") {
			scanResult := ScanResult{
				localName: ln,
				addr:      adv.Addr(),
				rssi:      adv.RSSI(),
			}
			canConnect = adv.Connectable()
			if canConnect {
				log.Debug("Tesla vehicle found! Services: %v, MD: %X.", adv.Services(), adv.ManufacturerData())
				scanResult.manufacData = adv.ManufacturerData()
				scanResult.services = adv.Services()
			}
			scanList.scanEntries = append(scanList.scanEntries, scanResult)
		}
		return false
	}

	_, err = ble.Connect(ctx, filter)

	return &scanList, err
}

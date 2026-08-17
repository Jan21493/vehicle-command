package ble

import (
	"context"
	"crypto/sha1"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/teslamotors/vehicle-command/internal/log"
	"github.com/teslamotors/vehicle-command/pkg/connector"
	"github.com/teslamotors/vehicle-command/pkg/protocol"
)

var ErrMaxConnectionsExceeded = protocol.NewError("the vehicle is already connected to the maximum number of BLE devices", false, false)

const (
	defaultMTU        = 23
	maxBLEMTUSize     = 512 + 3
	maxBLEMessageSize = 1024

	rxTimeout  = time.Second     // Timeout interval between receiving chunks of a message
	maxLatency = 4 * time.Second // Max allowed error when syncing vehicle clock
)

const (
	vehicleServiceUUID = "00000211-b2d1-43f0-9b88-960cebf8b91e"
	toVehicleUUID      = "00000212-b2d1-43f0-9b88-960cebf8b91e"
	fromVehicleUUID    = "00000213-b2d1-43f0-9b88-960cebf8b91e"
)

var vehicleLocalNamePattern = regexp.MustCompile(`^S[0-9A-Fa-f]{16}C$`)

// Connection implements connector.Connector over BLE.
type Connection struct {
	vin    string
	inbox  chan []byte
	device Device
	writer Writer
	rssi   int

	blockLength int
	inputBuffer []byte
	lastRx      time.Time
	lock        sync.Mutex
}

// ScanResult holds the details of a single BLE scan entry from tesla-blescan.
type ScanResult struct {
	localName   string
	addr        string
	rssi        int
	connectable bool
}

// ScanList holds the results of a full BLE scan.
type ScanList struct {
	scanEntries []ScanResult
}

func (scanList *ScanList) ScanEntries() []ScanResult {
	return scanList.scanEntries
}

func (scanEntry *ScanResult) ScanEntryString() string {
	return fmt.Sprintf("name: %s, addr: [%s], RSSI: %3d", scanEntry.localName, scanEntry.addr, scanEntry.rssi)
}

func (scanEntry *ScanResult) LocalName() string {
	return scanEntry.localName
}

func (scanEntry *ScanResult) RSSI() int {
	return scanEntry.rssi
}

func (scanEntry *ScanResult) VehicleScanResult() *VehicleScanResult {
	return &VehicleScanResult{
		Address:     scanEntry.addr,
		LocalName:   scanEntry.localName,
		RSSI:        int16(scanEntry.rssi),
		Connectable: scanEntry.connectable,
	}
}

func SetLogLevelTrace() {
	log.SetLevel(log.LevelDebug)
}

func SetLogLevelDebug() {
	log.SetLevel(log.LevelDebug)
}

func SetLogLevelInfo() {
	log.SetLevel(log.LevelInfo)
}

func SetLogLevelWarn() {
	log.SetLevel(log.LevelWarning)
}

func SetLogLevelError() {
	log.SetLevel(log.LevelError)
}

func SetLogLevelFatal() {
	log.SetLevel(log.LevelError)
}

func SetLogLevelPanic() {
	log.SetLevel(log.LevelError)
}

func VehicleLocalName(vin string) string {
	vinBytes := []byte(vin)
	digest := sha1.Sum(vinBytes)
	return fmt.Sprintf("S%02xC", digest[:8])
}

// IsVehicleLocalName reports whether s matches Tesla BLE local name format.
func IsVehicleLocalName(s string) bool {
	return vehicleLocalNamePattern.MatchString(s)
}

func ScanVehicleBeacon(ctx context.Context, vin string, adapter Adapter) (*Beacon, error) {
	return adapter.ScanBeacon(ctx, VehicleLocalName(vin))
}

func NewConnection(ctx context.Context, vin string, adapter Adapter) (*Connection, error) {
	beacon, err := adapter.ScanBeacon(ctx, VehicleLocalName(vin))
	if err != nil {
		return nil, err
	}
	return NewConnectionFromBeacon(ctx, vin, beacon, adapter)
}

// NewConnectionFromBeacon creates a new BLE connection to the given beacon.
func NewConnectionFromBeacon(ctx context.Context, vin string, beacon *Beacon, adapter Adapter) (*Connection, error) {
	var lastError error

	localName := VehicleLocalName(vin)
	if IsVehicleLocalName(vin) {
		localName = vin
	}

	if beacon.LocalName != localName {
		return nil, fmt.Errorf("ble: beacon with unexpected local name: '%s'", beacon.LocalName)
	}

	if !beacon.Connectable {
		return nil, ErrMaxConnectionsExceeded
	}

	for {
		conn, err := tryToConnect(ctx, vin, beacon, adapter)
		if err == nil {
			return conn, nil
		}

		log.Warning("BLE connection attempt failed: %+v", err)
		if err := ctx.Err(); err != nil {
			if lastError != nil {
				return nil, lastError
			}
			return nil, err
		}
		lastError = err
	}
}

func tryToConnect(ctx context.Context, vin string, beacon *Beacon, adapter Adapter) (*Connection, error) {
	dev, err := adapter.Connect(ctx, beacon)
	if err != nil {
		return nil, err
	}

	svc, err := dev.Service(ctx, vehicleServiceUUID)
	if err != nil {
		return nil, err
	}

	w, err := svc.Tx(toVehicleUUID)
	if err != nil {
		return nil, err
	}

	txMtu, err := w.MTU(maxBLEMTUSize)
	if err != nil {
		txMtu = defaultMTU - 3 // Fallback to default MTU size
	} else {
		txMtu = min(txMtu, maxBLEMessageSize) - 3 // 3 bytes for header
	}

	conn := &Connection{
		vin:         vin,
		inbox:       make(chan []byte, 5),
		device:      dev,
		writer:      w,
		blockLength: txMtu,
		rssi:        int(beacon.RSSI),
	}

	err = svc.Rx(fromVehicleUUID, conn.rx)
	if err != nil {
		return nil, err
	}

	log.Info("Connected to vehicle BLE")
	return conn, nil
}

// NewScan performs a full BLE scan and returns all Tesla vehicle beacons found.
func NewScan(ctx context.Context, adapter Adapter) (*ScanList, error) {
	scanList := &ScanList{}

	err := adapter.ScanAll(ctx, func(beacon *Beacon) {
		localName := beacon.LocalName
		if len(localName) > 0 {
			log.Debug("Advertisement from Name: %s [%s] RSSI: %3d:", localName, beacon.Address, beacon.RSSI)
		}
		if !IsVehicleLocalName(localName) {
			return
		}
		log.Debug("Tesla vehicle found! Name: %s, RSSI: %3d, Connectable: %v", localName, beacon.RSSI, beacon.Connectable)
		scanList.scanEntries = append(scanList.scanEntries, ScanResult{
			localName:   localName,
			addr:        beacon.Address,
			rssi:        int(beacon.RSSI),
			connectable: beacon.Connectable,
		})
	})

	return scanList, err
}

func (c *Connection) Receive() <-chan []byte {
	return c.inbox
}

func (c *Connection) Send(ctx context.Context, buffer []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	var out []byte
	log.Debug("TX: %02x", buffer)
	out = append(out, uint8(len(buffer)>>8), uint8(len(buffer)))
	out = append(out, buffer...)
	blockLength := c.blockLength
	for len(out) > 0 {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if blockLength > len(out) {
			blockLength = len(out)
		}

		n, err := c.writer.Write(out[:blockLength])
		if err != nil {
			return err
		} else if n != blockLength {
			return fmt.Errorf("ble: failed to write %d bytes", blockLength)
		}

		out = out[blockLength:]
	}
	return nil
}

func (c *Connection) VIN() string {
	return c.vin
}

func (c *Connection) RSSI() int {
	return c.rssi
}

func (c *Connection) Close() {
	if err := c.device.Close(); err != nil {
		log.Warning("ble: failed to close device: %s", err)
	}
}

func (c *Connection) PreferredAuthMethod() connector.AuthMethod {
	return connector.AuthMethodGCM
}

func (c *Connection) RetryInterval() time.Duration {
	return time.Second
}

func (c *Connection) AllowedLatency() time.Duration {
	return maxLatency
}

func (c *Connection) rx(p []byte) {
	if time.Since(c.lastRx) > rxTimeout {
		c.inputBuffer = []byte{}
	}
	c.lastRx = time.Now()
	c.inputBuffer = append(c.inputBuffer, p...)
	for c.flush() {
	}
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

// InitAdapterWithID is kept for backward compatibility.
// In the new design, adapters are created via goble.NewAdapter or tinygo.NewAdapter.
func InitAdapterWithID(_ string) error {
	return nil
}

// CloseAdapter is kept for backward compatibility.
// In the new design, adapters are closed via adapter.Close().
func CloseAdapter() error {
	return nil
}


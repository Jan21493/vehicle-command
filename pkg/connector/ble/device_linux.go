package ble

import (
	"os"
	"strconv"
	"strings"
	"time"

	// "github.com/go-ble/ble"
	// "github.com/go-ble/ble/linux"
	// "github.com/go-ble/ble/linux/hci/cmd"
	"github.com/rigado/ble"
	"github.com/rigado/ble/linux"
	"github.com/rigado/ble/linux/hci/cmd"
)

func IsAdapterError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "operation not permitted") ||
		strings.Contains(errMsg, "no valid transport found") ||
		strings.Contains(errMsg, "can't init hci") ||
		strings.Contains(errMsg, "device or resource busy")
}

func AdapterErrorHelpMessage(err error) string {
	// The underlying BLE package calls HCIDEVDOWN on the BLE device, presumably as a
	// heavy-handed way of dealing with devices that are in a bad state.
	if err != nil && strings.Contains(err.Error(), "device or resource busy") {
		return "Failed to initialize BLE adapter: \n\t" + err.Error() + "\n" +
			"The Bluetooth adapter is currently in use by another process (for example bluetoothd).\n" +
			"Try freeing the adapter and then run this command again:\n\n" +
			"\tsudo systemctl stop bluetooth\n" +
			"\tsudo hciconfig hci0 down\n\n" +
			"If needed, grant CAP_NET_ADMIN to this binary:\n\n" +
			"\tsudo setcap 'cap_net_admin=eip' \"$(which " + os.Args[0] + ")\""
	}

	return "Failed to initialize BLE adapter: \n\t" + err.Error() + "\n" +
		"Try again after granting this application CAP_NET_ADMIN or running with root:\n\n" +
		"\tsudo setcap 'cap_net_admin=eip' \"$(which " + os.Args[0] + ")\""
}

const bleTimeout = 20 * time.Second

// TODO: Depending on the model and state, BLE advertisements come every 20ms or every 150ms.

var scanParams = cmd.LESetScanParameters{
	LEScanType:           1,    // Active scanning
	LEScanInterval:       0x10, // 10ms
	LEScanWindow:         0x10, // 10ms
	OwnAddressType:       0,    // Static
	ScanningFilterPolicy: 0,    // Basic unfiltered
}

func newDevice() (ble.Device, error) {
	// Use auto adapter selection to avoid fixed-ID retry issues in the underlying BLE transport.
	device, err := linux.NewDevice(ble.OptListenerTimeout(bleTimeout), ble.OptDialerTimeout(bleTimeout), ble.OptTransportHCISocket(-1), ble.OptScanParams(scanParams))
	if err != nil {
		return nil, err
	}
	return device, nil
}

func newAdapter(id *string) (ble.Device, error) {
	hciID := -1
	opts := []ble.Option{
		ble.OptDialerTimeout(bleTimeout),
		ble.OptListenerTimeout(bleTimeout),
		ble.OptScanParams(scanParams),
		ble.OptTransportHCISocket(hciID),
	}
	if id != nil && *id != "" {
		if !strings.HasPrefix(*id, "hci") {
			return nil, ErrAdapterInvalidID
		}
		hciStr := strings.TrimPrefix(*id, "hci")
		parsedID, err := strconv.Atoi(hciStr)
		if err != nil || parsedID < 0 || parsedID > 15 {
			return nil, ErrAdapterInvalidID
		}
		hciID = parsedID
		opts[len(opts)-1] = ble.OptTransportHCISocket(hciID)
	}

	device, err := linux.NewDeviceWithName("vehicle-command", opts...)
	if err != nil {
		return nil, err
	}
	return device, nil
}

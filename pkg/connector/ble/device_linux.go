package ble

import (
	// "github.com/go-ble/ble"
	// "github.com/go-ble/ble/linux"
	// "github.com/go-ble/ble/linux/hci/cmd"
	"github.com/rigado/ble"
	"github.com/rigado/ble/linux"
	"github.com/rigado/ble/linux/hci/cmd"

	"time"
)

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
	device, err := linux.NewDevice(ble.OptListenerTimeout(bleTimeout), ble.OptDialerTimeout(bleTimeout), ble.OptTransportHCISocket(0), ble.OptScanParams(scanParams))
	if err != nil {
		return nil, err
	}
	return device, nil
}

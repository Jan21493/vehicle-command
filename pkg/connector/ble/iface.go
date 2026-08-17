package ble

import (
	"context"
	"io"
)

// Beacon holds information about a BLE advertisement from a vehicle.
type Beacon struct {
	Address     string
	LocalName   string
	RSSI        int16
	Connectable bool
}

// VehicleScanResult is an alias for Beacon for backward compatibility.
type VehicleScanResult = Beacon

// Adapter abstracts the BLE hardware and provides scanning and connecting capabilities.
type Adapter interface {
	// ScanBeacon scans for the first BLE advertisement matching the given local name.
	ScanBeacon(ctx context.Context, name string) (*Beacon, error)
	// ScanAll scans for all BLE advertisements, invoking handler for each one received.
	// Returns nil when the context is canceled (normal termination).
	ScanAll(ctx context.Context, handler func(*Beacon)) error
	// Connect establishes a BLE connection to the given beacon.
	Connect(ctx context.Context, beacon *Beacon) (Device, error)
	// Close releases resources held by the adapter.
	Close() error
}

// Device represents an active BLE connection to a peripheral.
type Device interface {
	Service(ctx context.Context, uuid string) (Service, error)
	Close() error
}

// Service represents a BLE GATT service.
type Service interface {
	Rx(uuid string, callback func(buf []byte)) error
	Tx(uuid string) (Writer, error)
}

// Writer provides write access to a BLE characteristic and MTU negotiation.
type Writer interface {
	io.Writer
	MTU(rxMTU int) (txMTU int, err error)
}

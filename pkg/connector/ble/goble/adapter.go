package goble

import (
	"context"
	"errors"

	"github.com/teslamotors/vehicle-command/pkg/connector/ble"
	"github.com/teslamotors/vehicle-command/pkg/protocol"
	rigado "github.com/rigado/ble"
)

var ErrAdapterInvalidID = protocol.NewError("the bluetooth adapter ID is invalid", false, false)

// NewAdapter creates a new BLE adapter using the rigado/ble library.
// id is the Bluetooth adapter identifier (e.g., "hci0" on Linux; empty string selects default).
func NewAdapter(id string) (ble.Adapter, error) {
	device, err := newAdapter(id)
	if err != nil {
		return nil, err
	}
	return &adapter{device: device}, nil
}

type adapter struct {
	device rigado.Device
}

func (s *adapter) ScanBeacon(ctx context.Context, name string) (*ble.Beacon, error) {
	scanCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var result *ble.Beacon

	fn := func(a rigado.Advertisement) {
		if name != a.LocalName() {
			return
		}
		result = advertisementToBeacon(a)
		cancel()
	}

	err := s.device.Scan(scanCtx, false, fn)
	if errors.Is(err, context.Canceled) {
		if result != nil {
			return result, nil
		}
		return nil, ctx.Err()
	}
	return result, err
}

func (s *adapter) ScanAll(ctx context.Context, handler func(*ble.Beacon)) error {
	scanCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	fn := func(a rigado.Advertisement) {
		handler(advertisementToBeacon(a))
	}

	err := s.device.Scan(scanCtx, false, fn)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (s *adapter) Connect(ctx context.Context, beacon *ble.Beacon) (ble.Device, error) {
	client, err := s.device.Dial(ctx, rigado.NewAddr(beacon.Address))
	if err != nil {
		return nil, err
	}
	return &device{client: client}, nil
}

func (s *adapter) Close() error {
	if s.device == nil {
		return nil
	}
	d := s.device
	s.device = nil
	return d.Stop()
}

func advertisementToBeacon(a rigado.Advertisement) *ble.Beacon {
	return &ble.Beacon{
		Address:     a.Addr().String(),
		LocalName:   a.LocalName(),
		RSSI:        int16(a.RSSI()),
		Connectable: a.Connectable(),
	}
}

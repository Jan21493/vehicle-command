package goble

import (
	"fmt"

	"github.com/teslamotors/vehicle-command/pkg/connector/ble"
	rigado "github.com/rigado/ble"
)

type service struct {
	client rigado.Client
	svc    *rigado.Service
}

func (s *service) Rx(uuid string, callback func(buf []byte)) error {
	characteristic, err := s.discover(uuid)
	if err != nil {
		return err
	}
	// rigado's Subscribe callback has signature func(id uint, p []byte); wrap it.
	if err := s.client.Subscribe(characteristic, true, func(_ uint, p []byte) { callback(p) }); err != nil {
		return fmt.Errorf("ble: failed to subscribe to RX: %s", err)
	}
	return nil
}

func (s *service) Tx(uuid string) (ble.Writer, error) {
	characteristic, err := s.discover(uuid)
	if err != nil {
		return nil, err
	}
	return &writer{characteristic: characteristic, client: s.client}, nil
}

func (s *service) discover(uuidStr string) (*rigado.Characteristic, error) {
	uuid := rigado.MustParse(uuidStr)
	characteristics, err := s.client.DiscoverCharacteristics([]rigado.UUID{uuid}, s.svc)
	if err != nil {
		return nil, fmt.Errorf("ble: failed to discover service characteristics: %s", err)
	}
	for _, char := range characteristics {
		if char.UUID.Equal(uuid) {
			if _, err := s.client.DiscoverDescriptors(nil, char); err != nil {
				return nil, fmt.Errorf("ble: couldn't fetch descriptors: %s", err)
			}
			return char, nil
		}
	}
	return nil, fmt.Errorf("ble: characteristic %s not found", uuidStr)
}

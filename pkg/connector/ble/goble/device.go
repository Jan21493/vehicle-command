package goble

import (
	"context"
	"errors"
	"fmt"

	"github.com/teslamotors/vehicle-command/pkg/connector/ble"
	rigado "github.com/rigado/ble"
)

type device struct {
	client rigado.Client
}

func (c *device) Service(_ context.Context, uuid string) (ble.Service, error) {
	services, err := c.client.DiscoverServices([]rigado.UUID{rigado.MustParse(uuid)})
	if err != nil {
		return nil, fmt.Errorf("ble: failed to enumerate device services: %s", err)
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("ble: failed to discover service")
	}
	return &service{client: c.client, svc: services[0]}, nil
}

func (c *device) Close() error {
	if c.client == nil {
		return nil
	}
	client := c.client
	c.client = nil
	err1 := client.ClearSubscriptions()
	err2 := client.CancelConnection()
	return errors.Join(err1, err2)
}

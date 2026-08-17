package goble

import rigado "github.com/rigado/ble"

type writer struct {
	characteristic *rigado.Characteristic
	client         rigado.Client
}

func (w *writer) Write(bytes []byte) (int, error) {
	err := w.client.WriteCharacteristic(w.characteristic, bytes, false)
	if err != nil {
		return 0, err
	}
	return len(bytes), nil
}

func (w *writer) MTU(rxMTU int) (txMTU int, err error) {
	return w.client.ExchangeMTU(rxMTU)
}

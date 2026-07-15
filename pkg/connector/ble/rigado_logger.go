package ble

import (
	rble "github.com/rigado/ble"
	"github.com/teslamotors/vehicle-command/internal/log"
)

// quietRigadoLogger suppresses verbose rigado BLE output that interferes with JSON consumers.
// Warnings and errors are still forwarded to the project logger.
type quietRigadoLogger struct{}

func (quietRigadoLogger) Info(...interface{}) {}

func (quietRigadoLogger) Debug(...interface{}) {}

func (quietRigadoLogger) Error(args ...interface{}) {
	log.Error("rigado/ble: %v", args)
}

func (quietRigadoLogger) Warn(args ...interface{}) {
	log.Warning("rigado/ble: %v", args)
}

func (quietRigadoLogger) Infof(string, ...interface{}) {}

func (quietRigadoLogger) Debugf(string, ...interface{}) {}

func (quietRigadoLogger) Errorf(format string, args ...interface{}) {
	log.Error("rigado/ble: "+format, args...)
}

func (quietRigadoLogger) Warnf(format string, args ...interface{}) {
	log.Warning("rigado/ble: "+format, args...)
}

func (quietRigadoLogger) ChildLogger(map[string]interface{}) rble.Logger {
	return quietRigadoLogger{}
}

func init() {
	rble.SetLogger(quietRigadoLogger{})
}

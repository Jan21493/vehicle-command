package ble

import (
	"fmt"
	"strings"

	rble "github.com/rigado/ble"
	"github.com/teslamotors/vehicle-command/internal/log"
)

func shouldIgnoreRigadoLine(msg string) bool {
	msg = strings.ToLower(msg)
	if strings.Contains(msg, "socket accept") && strings.Contains(msg, "timed out") {
		return true
	}
	if strings.Contains(msg, "event handler for 62 failed: <nil>") {
		return true
	}
	return strings.Contains(msg, "<nil>, bytes [")
}

// quietRigadoLogger suppresses verbose rigado BLE output that interferes with JSON consumers.
// Warnings and errors are still forwarded to the project logger.
type quietRigadoLogger struct{}

func (quietRigadoLogger) Info(...interface{}) {}

func (quietRigadoLogger) Debug(...interface{}) {}

func (quietRigadoLogger) Error(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if shouldIgnoreRigadoLine(msg) {
		return
	}
	log.Error("rigado/ble: %s", msg)
}

func (quietRigadoLogger) Warn(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if shouldIgnoreRigadoLine(msg) {
		return
	}
	log.Warning("rigado/ble: %s", msg)
}

func (quietRigadoLogger) Infof(string, ...interface{}) {}

func (quietRigadoLogger) Debugf(string, ...interface{}) {}

func (quietRigadoLogger) Errorf(format string, args ...interface{}) {
	if shouldIgnoreRigadoLine(format) || shouldIgnoreRigadoLine(fmt.Sprintf(format, args...)) {
		return
	}
	log.Error("rigado/ble: "+format, args...)
}

func (quietRigadoLogger) Warnf(format string, args ...interface{}) {
	if shouldIgnoreRigadoLine(format) || shouldIgnoreRigadoLine(fmt.Sprintf(format, args...)) {
		return
	}
	log.Warning("rigado/ble: "+format, args...)
}

func (quietRigadoLogger) ChildLogger(map[string]interface{}) rble.Logger {
	return quietRigadoLogger{}
}

func init() {
	rble.SetLogger(quietRigadoLogger{})
}

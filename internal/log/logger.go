// Package log provides a global logger with configurable logging level. The intended use is for
// development builds.

package log

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Level int

const (
	LevelNone    Level = iota // Disables logging.
	LevelError                // Logs anamolies that are not expected to occur during normal use.
	LevelWarning              // Logs anamolies that are expected to occur occasionally during normal use.
	LevelInfo                 // Logs major events.
	LevelDebug                // Logs detailed IO
)

var globalLogLevel Level
var logMutex sync.Mutex

var labels = map[Level]string{
	LevelDebug:   "[DEBUG]",
	LevelInfo:    "[INFO ]",
	LevelWarning: "[WARN ]",
	LevelError:   "[ERROR]",
}

func SetLevel(level Level) {
	logMutex.Lock()
	defer logMutex.Unlock()
	globalLogLevel = level
}

func logLevel() Level {
	logMutex.Lock()
	defer logMutex.Unlock()
	return globalLogLevel
}

func log(level Level, format string, a ...interface{}) {
	if level <= logLevel() {
		// msg := fmt.Sprintf("%s %s ", time.Now().Format(time.RFC3339), labels[level])
		msg := fmt.Sprintf("%s %s ", time.Now().Format("2006-01-02 15:04:05.000"), labels[level])
		msg += fmt.Sprintf(format, a...)
		fmt.Fprintln(os.Stderr, msg)
	}
}

func Debug(format string, a ...interface{}) {
	log(LevelDebug, format, a...)
}
func Info(format string, a ...interface{}) {
	log(LevelInfo, format, a...)
}
func Warning(format string, a ...interface{}) {
	log(LevelWarning, format, a...)
}
func Error(format string, a ...interface{}) {
	log(LevelError, format, a...)
}

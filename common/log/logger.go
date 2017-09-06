package logger

import "os"

// LogLevel is used to determine what to log.
type LogLevel int

func (l LogLevel) String() string {
	switch l {
	case 31:
		return "FATAL"
	case 15:
		return "ERROR"
	case 7:
		return "WARN"
	case 3:
		return "INFO"
	case 1:
		return "DEBUG"
	default:
		return "unknown"
	}
}

// The available log levels
const (
	FATAL LogLevel = 31
	ERROR LogLevel = 15
	WARN  LogLevel = 7
	INFO  LogLevel = 3
	DEBUG LogLevel = 1
)

var (
	level = ERROR
)

// SetLogLevel should be used to set the loglevel of the application
func SetLogLevel(logLevel LogLevel) {
	level = logLevel
}

// Fatal should be used to log a critical incident and exit the application
func Fatal(msg string) {
	defer os.Exit(1)

}

// Error should be used for application errors that should be resolved
func Error(msg string) {
	if shouldLog(ERROR) {

	}
}

// Warn should be used for events that can be dangerous
func Warn(msg string) {
	if shouldLog(WARN) {

	}
}

// Info should be used to share data.
func Info(msg string) {
	if shouldLog(INFO) {

	}
}

// Debug should be used for debugging purposes only.
func Debug(msg string) {
	if shouldLog(DEBUG) {

	}
}

func shouldLog(lvl LogLevel) bool {
	if level&lvl == level {
		return true
	}

	return false
}

package logger

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/djavorszky/rlog"
	"google.golang.org/grpc"
)

// LogLevel is used to determine what to log.
type LogLevel int

func (l LogLevel) String() string {
	switch l {
	case 31:
		return "fatal"
	case 15:
		return "error"
	case 7:
		return "warn"
	case 3:
		return "info"
	case 1:
		return "debug"
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
	level  = INFO
	remote = false

	conn   *grpc.ClientConn
	client rlog.LogClient
	id     int32
)

// UseRemoteLogger will try to register the app to the remote
// logger at the addr endpoint.
func UseRemoteLogger(addr, app, service string) error {
	var err error

	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("dialing rlog: %v", err)
	}

	client = rlog.NewLogClient(conn)

	resp, err := client.Register(context.Background(), &rlog.RegisterRequest{App: app, Service: service})
	if err != nil {
		return fmt.Errorf("register: %v", err)
	}

	id = resp.Id
	remote = true

	return nil
}

// Close closes the client connection to the remote logger.
func Close() {
	conn.Close()
}

// SetLogLevel should be used to set the loglevel of the application
func SetLogLevel(logLevel LogLevel) {
	level = logLevel
}

// Fatal should be used to log a critical incident and exit the application
func Fatal(msg string, args ...interface{}) {
	doLog(FATAL, client.Fatal, msg, args...)
	os.Exit(1)
}

// Error should be used for application errors that should be resolved
func Error(msg string, args ...interface{}) {
	doLog(ERROR, client.Error, msg, args...)
}

// Warn should be used for events that can be dangerous
func Warn(msg string, args ...interface{}) {
	doLog(WARN, client.Warn, msg, args...)
}

// Info should be used to share data.
func Info(msg string, args ...interface{}) {
	doLog(INFO, client.Info, msg, args...)
}

// Debug should be used for debugging purposes only.
func Debug(msg string, args ...interface{}) {
	doLog(DEBUG, client.Debug, msg, args...)
}

func shouldLog(lvl LogLevel) bool {
	if level&lvl == level {
		return true
	}

	return false
}

func doLog(level LogLevel, remotely func(context.Context, *rlog.LogMessage, ...grpc.CallOption) (*rlog.LogResponse, error), msg string, args ...interface{}) {
	if shouldLog(level) {
		if remote {
			_, err := remotely(context.Background(), &rlog.LogMessage{Id: id, Message: fmt.Sprintf(msg, args...)})
			if err != nil {
				remote = false
				Error("Remote logger disappeared: %v", err)
			}
		}

		log.Printf("[%s] %s", level, fmt.Sprintf(msg, args...))
	}
}

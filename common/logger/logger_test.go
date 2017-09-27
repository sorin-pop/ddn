package logger

import (
	"testing"
)

func TestShouldLog(t *testing.T) {
	tests := []struct {
		Level     LogLevel
		CheckName string
		Check     map[LogLevel]bool
	}{
		{CheckName: "FATAL", Level: FATAL, Check: map[LogLevel]bool{FATAL: true, ERROR: false, WARN: false, INFO: false, DEBUG: false}},
		{CheckName: "ERROR", Level: ERROR, Check: map[LogLevel]bool{FATAL: true, ERROR: true, WARN: false, INFO: false, DEBUG: false}},
		{CheckName: "WARN", Level: WARN, Check: map[LogLevel]bool{FATAL: true, ERROR: true, WARN: true, INFO: false, DEBUG: false}},
		{CheckName: "INFO", Level: INFO, Check: map[LogLevel]bool{FATAL: true, ERROR: true, WARN: true, INFO: true, DEBUG: false}},
		{CheckName: "DEBUG", Level: DEBUG, Check: map[LogLevel]bool{FATAL: true, ERROR: true, WARN: true, INFO: true, DEBUG: true}},
	}

	for _, test := range tests {
		Level = test.Level
		for lvl, expected := range test.Check {
			if shouldLog(lvl) != expected {
				t.Errorf("shouldLog(%s) <-> %s is expected to be %t", lvl, test.CheckName, expected)
			}
		}
	}
}

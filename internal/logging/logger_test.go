package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.level.String() != tt.expected {
				t.Errorf("LogLevel(%d).String() = %s, want %s", tt.level, tt.level.String(), tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(INFO, buf, "test")

	if logger.level != INFO {
		t.Errorf("Expected level INFO, got %v", logger.level)
	}
	if logger.prefix != "test" {
		t.Errorf("Expected prefix 'test', got %s", logger.prefix)
	}
	if logger.logger == nil {
		t.Errorf("Logger is nil")
	}
}

func TestDefaultLogger(t *testing.T) {
	// Save original env and restore after test
	origEnv := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", origEnv)

	tests := []struct {
		envValue string
		expected LogLevel
	}{
		{"", INFO},
		{"DEBUG", DEBUG},
		{"INFO", INFO},
		{"WARN", WARN},
		{"ERROR", ERROR},
		{"FATAL", FATAL},
		{"INVALID", INFO}, // Should default to INFO for invalid values
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", tt.envValue)
			logger := DefaultLogger("test")
			if logger.level != tt.expected {
				t.Errorf("DefaultLogger with LOG_LEVEL=%s: expected level %v, got %v",
					tt.envValue, tt.expected, logger.level)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	logger := New(INFO, &bytes.Buffer{}, "test")
	logger.SetLevel(DEBUG)
	if logger.level != DEBUG {
		t.Errorf("Expected level DEBUG after SetLevel, got %v", logger.level)
	}
}

func TestLogMethods(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  LogLevel
		logFunc   func(l *Logger, msg string)
		message   string
		shouldLog bool
		levelStr  string
	}{
		{"Debug at DEBUG level", DEBUG, func(l *Logger, msg string) { l.Debug("%s", msg) }, "test message", true, "DEBUG"},
		{"Debug at INFO level", INFO, func(l *Logger, msg string) { l.Debug("%s", msg) }, "test message", false, "DEBUG"},
		{"Info at INFO level", INFO, func(l *Logger, msg string) { l.Info("%s", msg) }, "test message", true, "INFO"},
		{"Info at WARN level", WARN, func(l *Logger, msg string) { l.Info("%s", msg) }, "test message", false, "INFO"},
		{"Warn at WARN level", WARN, func(l *Logger, msg string) { l.Warn("%s", msg) }, "test message", true, "WARN"},
		{"Warn at ERROR level", ERROR, func(l *Logger, msg string) { l.Warn("%s", msg) }, "test message", false, "WARN"},
		{"Error at ERROR level", ERROR, func(l *Logger, msg string) { l.Error("%s", msg) }, "test message", true, "ERROR"},
		{"Error at FATAL level", FATAL, func(l *Logger, msg string) { l.Error("%s", msg) }, "test message", false, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := New(tt.logLevel, buf, "test")

			tt.logFunc(logger, tt.message)

			output := buf.String()
			if tt.shouldLog {
				if !strings.Contains(output, tt.message) {
					t.Errorf("Expected log to contain message '%s', got '%s'", tt.message, output)
				}
				if !strings.Contains(output, "["+tt.levelStr+"]") {
					t.Errorf("Expected log to contain level '[%s]', got '%s'", tt.levelStr, output)
				}
				if !strings.Contains(output, "[test]") {
					t.Errorf("Expected log to contain prefix '[test]', got '%s'", output)
				}
			} else {
				if output != "" {
					t.Errorf("Expected no log output, got '%s'", output)
				}
			}
		})
	}
}

// TestFatal can't easily test os.Exit behavior, so we just test the logging part
func TestFatal(t *testing.T) {
	// Replace os.Exit with a no-op for testing
	origExit := osExit
	defer func() { osExit = origExit }()

	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		if code != 1 {
			t.Errorf("Expected exit code 1, got %d", code)
		}
	}

	buf := &bytes.Buffer{}
	logger := New(FATAL, buf, "test")

	logger.Fatal("%s", "fatal message")

	if !exitCalled {
		t.Error("os.Exit was not called")
	}

	output := buf.String()
	if !strings.Contains(output, "fatal message") {
		t.Errorf("Expected log to contain 'fatal message', got '%s'", output)
	}
	if !strings.Contains(output, "[FATAL]") {
		t.Errorf("Expected log to contain level '[FATAL]', got '%s'", output)
	}
}

package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// For testing purposes, we can replace this function
var osExit = os.Exit

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DEBUG level for detailed troubleshooting information
	DEBUG LogLevel = iota
	// INFO level for general operational information
	INFO
	// WARN level for warning conditions
	WARN
	// ERROR level for error conditions
	ERROR
	// FATAL level for critical errors that cause the program to exit
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a logger instance
type Logger struct {
	level  LogLevel
	prefix string
	logger *log.Logger
}

// New creates a new logger with the specified level and output
func New(level LogLevel, output io.Writer, prefix string) *Logger {
	return &Logger{
		level:  level,
		prefix: prefix,
		logger: log.New(output, "", log.LstdFlags),
	}
}

// DefaultLogger creates a new logger with default settings
func DefaultLogger(prefix string) *Logger {
	level := INFO
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		switch strings.ToUpper(levelStr) {
		case "DEBUG":
			level = DEBUG
		case "INFO":
			level = INFO
		case "WARN":
			level = WARN
		case "ERROR":
			level = ERROR
		case "FATAL":
			level = FATAL
		}
	}
	return New(level, os.Stderr, prefix)
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.log(DEBUG, format, v...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.log(INFO, format, v...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= WARN {
		l.log(WARN, format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.log(ERROR, format, v...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.level <= FATAL {
		l.log(FATAL, format, v...)
		osExit(1)
	}
}

// log logs a message with the specified level
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	prefix := fmt.Sprintf("[%s] ", level)
	if l.prefix != "" {
		prefix = fmt.Sprintf("[%s] [%s] ", level, l.prefix)
	}
	if err := l.logger.Output(2, prefix+fmt.Sprintf(format, v...)); err != nil {
		fmt.Fprintf(os.Stderr, "Error logging message: %v\n", err)
	}
}

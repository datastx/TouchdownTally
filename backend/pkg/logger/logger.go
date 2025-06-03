package logger

import (
	"fmt"
	"log"
	"os"
)

// Logger provides a simple logging interface
type Logger struct {
	*log.Logger
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info logs an info message with key-value pairs
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("INFO", msg, keysAndValues...)
}

// Error logs an error message with key-value pairs
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("ERROR", msg, keysAndValues...)
}

// Warn logs a warning message with key-value pairs
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("WARN", msg, keysAndValues...)
}

// Debug logs a debug message with key-value pairs
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("DEBUG", msg, keysAndValues...)
}

// Fatal logs at error level and exits the program
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Error(msg, args...)
	os.Exit(1)
}

// HTTP logs HTTP request information
func (l *Logger) HTTP(method, path string, statusCode int, duration string, args ...interface{}) {
	kvs := append(args, "method", method, "path", path, "status", statusCode, "duration", duration)
	l.Info("HTTP request", kvs...)
}

func (l *Logger) logWithLevel(level, msg string, keysAndValues ...interface{}) {
	var kvPairs []string
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			kvPairs = append(kvPairs, fmt.Sprintf("%v=%v", keysAndValues[i], keysAndValues[i+1]))
		}
	}
	
	if len(kvPairs) > 0 {
		l.Printf("[%s] %s (%s)", level, msg, fmt.Sprintf("%v", kvPairs))
	} else {
		l.Printf("[%s] %s", level, msg)
	}
}

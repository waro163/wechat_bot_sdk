package common

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

// Field represents a logging field
type Field struct {
	Key   string
	Value interface{}
}

// defaultLogger implements Logger using standard log package
type defaultLogger struct {
	accountID string
	logger    *log.Logger
}

// NewDefaultLogger creates a default logger that writes to stderr
func NewDefaultLogger(accountID string) Logger {
	return &defaultLogger{
		accountID: accountID,
		logger:    log.New(os.Stderr, "", log.LstdFlags),
	}
}

// Debug logs a debug message
func (l *defaultLogger) Debug(msg string, fields ...Field) {
	l.log("DEBUG", msg, fields)
}

// Info logs an info message
func (l *defaultLogger) Info(msg string, fields ...Field) {
	l.log("INFO", msg, fields)
}

// Warn logs a warning message
func (l *defaultLogger) Warn(msg string, fields ...Field) {
	l.log("WARN", msg, fields)
}

// Error logs an error message
func (l *defaultLogger) Error(msg string, fields ...Field) {
	l.log("ERROR", msg, fields)
}

// log is the internal logging method
func (l *defaultLogger) log(level, msg string, fields []Field) {
	prefix := fmt.Sprintf("[%s]", level)
	if l.accountID != "" {
		prefix = fmt.Sprintf("[%s] [account=%s]", level, l.accountID)
	}

	if len(fields) == 0 {
		l.logger.Printf("%s %s", prefix, msg)
		return
	}

	// Build field string
	fieldStr := ""
	for i, field := range fields {
		if i > 0 {
			fieldStr += " "
		}
		fieldStr += fmt.Sprintf("%s=%v", field.Key, field.Value)
	}

	l.logger.Printf("%s %s | %s", prefix, msg, fieldStr)
}

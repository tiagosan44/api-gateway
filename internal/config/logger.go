package config

import (
	"encoding/json"
	"os"
	"time"
)

// Logger provides structured JSON logging
type Logger struct {
	level string
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// NewLogger creates a new logger instance
func NewLogger(level string) *Logger {
	return &Logger{level: level}
}

// shouldLog checks if a log level should be logged
func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
	}

	currentLevel, ok := levels[l.level]
	if !ok {
		currentLevel = 1 // default to info
	}

	logLevel, ok := levels[level]
	if !ok {
		return false
	}

	return logLevel >= currentLevel
}

// log writes a log entry
func (l *Logger) log(level, message string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		os.Stderr.WriteString(level + ": " + message + "\n")
		return
	}

	os.Stderr.Write(jsonData)
	os.Stderr.WriteString("\n")
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.log("debug", message, fields)
}

// Info logs an info message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.log("info", message, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.log("warn", message, fields)
}

// Error logs an error message
func (l *Logger) Error(message string, fields map[string]interface{}) {
	l.log("error", message, fields)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, fields map[string]interface{}) {
	l.log("fatal", message, fields)
	os.Exit(1)
}

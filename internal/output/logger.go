package output

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

// LogLevel represents the level of logging
type LogLevel string

const (
	LevelInfo    LogLevel = "INFO"
	LevelWarning LogLevel = "WARNING"
	LevelError   LogLevel = "ERROR"
	LevelDebug   LogLevel = "DEBUG"
)

// Logger handles all output operations
type Logger struct {
	red    func(a ...interface{}) string
	yellow func(a ...interface{}) string
	cyan   func(a ...interface{}) string
	green  func(a ...interface{}) string
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		red:    color.New(color.FgRed).SprintFunc(),
		yellow: color.New(color.FgYellow).SprintFunc(),
		cyan:   color.New(color.FgCyan).SprintFunc(),
		green:  color.New(color.FgGreen).SprintFunc(),
	}
}

// Close is a no-op now that we don't use log files
func (l *Logger) Close() error {
	return nil
}

// logWithTimestamp outputs a message with timestamp and level
func (l *Logger) logWithTimestamp(level LogLevel, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("%s [%s] %s", timestamp, level, message)

	switch level {
	case LevelError:
		fmt.Println(l.red(logLine))
	case LevelWarning:
		fmt.Println(l.yellow(logLine))
	case LevelInfo:
		fmt.Println(l.green(logLine))
	case LevelDebug:
		fmt.Println(l.cyan(logLine))
	default:
		fmt.Println(logLine)
	}
}

// Info outputs an informational message
func (l *Logger) Info(message string) {
	l.logWithTimestamp(LevelInfo, message)
}

// Warning outputs a warning message
func (l *Logger) Warning(message string) {
	l.logWithTimestamp(LevelWarning, message)
}

// Error outputs an error message
func (l *Logger) Error(message string) {
	l.logWithTimestamp(LevelError, message)
}

// Fatal outputs an error message and exits
func (l *Logger) Fatal(message string) {
	l.Error(message)
	os.Exit(1)
}

// Debug outputs a debug message to console
func (l *Logger) Debug(message string) {
	l.logWithTimestamp(LevelDebug, message)
}

// LogConfiguration logs configuration key-value pairs
func (l *Logger) LogConfiguration(key, value string) {
	message := fmt.Sprintf("Configuration: %s = %s", key, value)
	l.logWithTimestamp(LevelInfo, message)
}

// LogStatistic logs a statistic key-value pair
func (l *Logger) LogStatistic(key, value string) {
	message := fmt.Sprintf("Result: %s = %s", key, value)
	l.logWithTimestamp(LevelInfo, message)
}

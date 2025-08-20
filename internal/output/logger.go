package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Logger handles all output operations
type Logger struct {
	red    func(a ...interface{}) string
	yellow func(a ...interface{}) string
	cyan   func(a ...interface{}) string
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		red:    color.New(color.FgRed).SprintFunc(),
		yellow: color.New(color.FgYellow).SprintFunc(),
		cyan:   color.New(color.FgCyan).SprintFunc(),
	}
}

// Close is a no-op now that we don't use log files
func (l *Logger) Close() error {
	return nil
}

// Info outputs an informational message
func (l *Logger) Info(message string) {
	fmt.Println(message)
}

// Warning outputs a warning message
func (l *Logger) Warning(message string) {
	msg := fmt.Sprintf("[WARNING] %s", message)
	fmt.Println(l.yellow(msg))
}

// Error outputs an error message
func (l *Logger) Error(message string) {
	msg := fmt.Sprintf("[ERROR] %s", message)
	fmt.Println(l.red(msg))
}

// Fatal outputs an error message and exits
func (l *Logger) Fatal(message string) {
	l.Error(message)
	os.Exit(1)
}

// Flag outputs a flag with its value
func (l *Logger) Flag(key, value string) {
	sep := ": "
	msg := fmt.Sprintf("%s%s%s", l.cyan(key), sep, value)
	fmt.Println(msg)
}

// Debug outputs a debug message to console
func (l *Logger) Debug(message string) {
	fmt.Println(message)
}

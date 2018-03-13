package utils

import (
	"os"

	"github.com/fatih/color"
	"github.com/palantir/stacktrace"
)

// Info .
func Info(msg string, args ...interface{}) {
	color.New(color.FgBlue).Printf(msg, args...)
}

// Info2 .
func Info2(msg string, args ...interface{}) {
	color.New(color.FgCyan).Printf(msg, args...)
}

// Warn .
func Warn(msg string, args ...interface{}) {
	color.New(color.FgYellow).Printf(msg, args...)
}

// Success .
func Success(msg string, args ...interface{}) {
	color.New(color.FgGreen).Printf(msg, args...)
}

// Error .
func Error(err error) {
	debug := os.Getenv("DEBUG")
	if debug == "true" || debug == "1" {
		color.New(color.FgRed).Fprintln(os.Stderr, err)
	} else {
		color.New(color.FgRed).Fprintln(os.Stderr, stacktrace.RootCause(err))
	}
}

// Fatal .
func Fatal(err error) {
	Error(err)
	os.Exit(1)
}

package utils

import (
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/palantir/stacktrace"
)

// Info .
func Info(msg string, args ...interface{}) {
	color.New(color.FgBlue).Printf(appendNewLine(msg), args...)
}

// Info2 .
func Info2(msg string, args ...interface{}) {
	color.New(color.FgCyan).Printf(appendNewLine(msg), args...)
}

// Warn .
func Warn(msg string, args ...interface{}) {
	color.New(color.FgYellow).Printf(appendNewLine(msg), args...)
}

// Success .
func Success(msg string, args ...interface{}) {
	color.New(color.FgGreen).Printf(appendNewLine(msg), args...)
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

func appendNewLine(msg string) string {
	if strings.HasSuffix(msg, "\n") {
		return msg
	}
	return msg + "\n"
}

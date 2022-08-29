package utils

import (
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/palantir/stacktrace"
)

// Infof .
func Infof(out io.Writer, msg string, args ...interface{}) {
	color.New(color.FgBlue).Fprintf(out, appendNewLine(msg), args...)
}

// Info .
func Info(msg string, args ...interface{}) {
	Infof(os.Stdout, msg, args...)
}

// Infof2 .
func Infof2(out io.Writer, msg string, args ...interface{}) {
	color.New(color.FgCyan).Fprintf(out, appendNewLine(msg), args...)
}

// Info2 .
func Info2(msg string, args ...interface{}) {
	Infof2(os.Stdout, msg, args...)
}

// Warnf .
func Warnf(out io.Writer, msg string, args ...interface{}) {
	color.New(color.FgYellow).Fprintf(out, appendNewLine(msg), args...)
}

// Warn .
func Warn(msg string, args ...interface{}) {
	Warnf(os.Stdout, msg, args...)
}

// Successf .
func Successf(out io.Writer, msg string, args ...interface{}) {
	color.New(color.FgGreen).Fprintf(out, appendNewLine(msg), args...)
}

// Success .
func Success(msg string, args ...interface{}) {
	Successf(os.Stdout, msg, args...)
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

package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/palantir/stacktrace"
)

// Command only Unix is supported
type Command struct {
	name          string
	args          []string
	execCmd       *exec.Cmd
	defaultStdout *bytes.Buffer
	defaultStderr *bytes.Buffer
}

// CommandStatus .
type CommandStatus struct {
	ExitCode    int
	Stdout      io.Reader
	Stderr      io.Reader
	ElaspedTime time.Duration
}

// NewCommand .
func NewCommand(name string, args ...string) *Command {
	cmd := &Command{
		name:          name,
		args:          args,
		execCmd:       exec.Command(name, args...),
		defaultStdout: &bytes.Buffer{},
		defaultStderr: &bytes.Buffer{},
	}
	cmd.execCmd.Stdout = cmd.defaultStdout
	cmd.execCmd.Stderr = cmd.defaultStderr
	return cmd
}

// SetStdout .
func (cmd *Command) SetStdout(w io.Writer) {
	cmd.execCmd.Stdout = w
	cmd.defaultStdout = nil
}

// SetStderr .
func (cmd *Command) SetStderr(w io.Writer) {
	cmd.execCmd.Stderr = w
	cmd.defaultStderr = nil
}

// CombineOutput to command stdout buffer
func (cmd *Command) CombineOutput() *Command {
	cmd.execCmd.Stdout = cmd.defaultStdout
	cmd.execCmd.Stderr = cmd.defaultStdout
	cmd.defaultStderr = nil
	return cmd
}

// RedirectToStandard os.Stdout and os.Stderr
func (cmd *Command) RedirectToStandard() {
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)
}

// SilenceOutput .
func (cmd *Command) SilenceOutput() {
	w := &nullWriter{}
	cmd.SetStdout(w)
	cmd.SetStderr(w)
}

// Run the command.
// If the command exit non-zero, the returned error is still nil.
func (cmd *Command) Run() (*CommandStatus, error) {
	startTime := time.Now()
	err := cmd.execCmd.Run()
	endTime := time.Now()
	elaspedTime := endTime.Sub(startTime)
	if err == nil {
		return &CommandStatus{
			ExitCode:    0,
			Stdout:      cmd.defaultStdout,
			Stderr:      cmd.defaultStderr,
			ElaspedTime: elaspedTime,
		}, nil
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return nil, stacktrace.Propagate(err, "execute command error")
	}
	exitCode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
	return &CommandStatus{
		ExitCode:    exitCode,
		Stdout:      cmd.defaultStdout,
		Stderr:      cmd.defaultStderr,
		ElaspedTime: elaspedTime,
	}, nil
}

// ExecuteCommand and redirect output to standard output/standard error
func ExecuteCommand(command string, args ...string) (*CommandStatus, error) {
	silentFlag := os.Getenv("SILENCE_OUTPUT")
	cmd := NewCommand(command, args...)
	if silentFlag == "true" || silentFlag == "1" {
		cmd.SilenceOutput()
	} else {
		cmd.RedirectToStandard()
	}
	return cmd.Run()
}

// ExecuteCommandSilently .
func ExecuteCommandSilently(command string, args ...string) (*CommandStatus, error) {
	cmd := NewCommand(command, args...)
	w := &nullWriter{}
	cmd.SetStdout(w)
	cmd.SetStderr(w)
	return cmd.Run()
}

type nullWriter struct {
}

func (w *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

package kubernetes

import (
	"fmt"
	"io/ioutil"

	"github.com/anduintransaction/rivendell/utils"
)

// ErrMissingCommand .
type ErrMissingCommand struct {
	Command string
}

func (err ErrMissingCommand) Error() string {
	return fmt.Sprintf("missing %q", err.Command)
}

// ErrCommandExecute .
type ErrCommandExecute struct {
	ExitCode int
	Output   string
}

func commandErrorFromStatus(status *utils.CommandStatus) ErrCommandExecute {
	output, _ := ioutil.ReadAll(status.Stdout)
	return ErrCommandExecute{status.ExitCode, string(output)}
}

func (err ErrCommandExecute) Error() string {
	return fmt.Sprintf("execute command error, exit code: %d, output: %q", err.ExitCode, err.Output)
}

// ErrCommandExitCode .
type ErrCommandExitCode struct {
	ExitCode int
}

func (err ErrCommandExitCode) Error() string {
	return fmt.Sprintf("execute command error, exit code: %d", err.ExitCode)
}

// ErrUnknownStatus .
type ErrUnknownStatus struct {
}

func (err ErrUnknownStatus) Error() string {
	return fmt.Sprintf("unknown status")
}

// ErrTimeout .
type ErrTimeout struct {
}

func (err ErrTimeout) Error() string {
	return fmt.Sprintf("timeout reached")
}

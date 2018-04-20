package kubernetes

import (
	"fmt"
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

// ErrInvalidResponse .
type ErrInvalidResponse struct {
	Underlying error
	Response   string
}

func (err ErrInvalidResponse) Error() string {
	return fmt.Sprintf("invalid response, error is %s, response is: %q", err.Underlying, err.Response)
}

// ErrUnknownStatus .
type ErrUnknownStatus struct {
	Name   string
	Kind   string
	status RsStatus
}

func (err ErrUnknownStatus) Error() string {
	return fmt.Sprintf("unknown status for %s %q (%d)", err.Kind, err.Name, err.status)
}

// ErrTimeout .
type ErrTimeout struct {
}

func (err ErrTimeout) Error() string {
	return fmt.Sprintf("timeout reached")
}

// ErrUnsupportedKind .
type ErrUnsupportedKind struct {
	Kind string
}

func (err ErrUnsupportedKind) Error() string {
	return fmt.Sprintf("unsupported kind: %s", err.Kind)
}

// ErrNotExist .
type ErrNotExist struct {
	Name string
	Kind string
}

func (err ErrNotExist) Error() string {
	return fmt.Sprintf("not exist: %s %q", err.Kind, err.Name)
}

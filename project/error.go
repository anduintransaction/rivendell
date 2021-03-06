package project

import (
	"fmt"
)

// ErrMissingDependency .
type ErrMissingDependency struct {
	Child  string
	Parent string
}

func (err ErrMissingDependency) Error() string {
	return fmt.Sprintf("missing dependency %q for %q", err.Parent, err.Child)
}

// ErrCyclicDependency .
type ErrCyclicDependency struct {
	Node string
}

func (err ErrCyclicDependency) Error() string {
	return fmt.Sprintf("cyclic dependency found for %q", err.Node)
}

// ErrWaitTimeout .
type ErrWaitTimeout struct {
	Name string
	Kind string
}

func (err ErrWaitTimeout) Error() string {
	return fmt.Sprintf("wait timeout for %s %q", err.Kind, err.Name)
}

// ErrWaitFailed .
type ErrWaitFailed struct {
	Name string
	Kind string
}

func (err ErrWaitFailed) Error() string {
	return fmt.Sprintf("wait failed for %s %q", err.Kind, err.Name)
}

package kubernetes

import (
	"bufio"
	"regexp"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
)

// Namespace operations
type Namespace struct {
	context *Context
}

// Create namespace in context
func (n *Namespace) Create() (exists bool, err error) {
	exists, err = n.Exists()
	if err != nil {
		return
	}
	if exists {
		return
	}
	args := n.context.completeArgsWithoutNamespace([]string{"create", "ns", n.context.namespace})
	status, err := utils.ExecuteCommand("kubectl", args...)
	if err != nil {
		return
	}
	if status.ExitCode != 0 {
		return false, stacktrace.Propagate(ErrCommandExitCode{status.ExitCode}, "command execute error")
	}
	return
}

// Exists .
func (n *Namespace) Exists() (bool, error) {
	args := n.context.completeArgsWithoutNamespace([]string{"get", "ns"})
	status, err := utils.NewCommand("kubectl", args...).CombineOutput().Run()
	if err != nil {
		return false, err
	}
	if status.ExitCode != 0 {
		return false, stacktrace.Propagate(commandErrorFromStatus(status), "command execute error")
	}
	splitRegex := regexp.MustCompile("\\s+")
	scanner := bufio.NewScanner(status.Stdout)
	firstLine := false
	found := false
	for scanner.Scan() {
		if firstLine {
			firstLine = true
			continue
		}
		parts := splitRegex.Split(scanner.Text(), -1)
		if len(parts) != 3 {
			continue
		}
		if parts[0] == n.context.namespace && parts[1] == "Active" {
			found = true
			break
		}
	}
	err = scanner.Err()
	if err != nil {
		return false, stacktrace.Propagate(err, "cannot read from command output")
	}
	return found, nil
}

// Delete .
func (n *Namespace) Delete() (exists bool, err error) {
	exists, err = n.Exists()
	if err != nil {
		return
	}
	if !exists {
		return
	}
	args := n.context.completeArgsWithoutNamespace([]string{"delete", "ns", n.context.namespace})
	status, err := utils.ExecuteCommand("kubectl", args...)
	if err != nil {
		return
	}
	if status.ExitCode != 0 {
		return false, stacktrace.Propagate(ErrCommandExitCode{status.ExitCode}, "command execute error")
	}
	return
}

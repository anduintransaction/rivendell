package kubernetes

import (
	"bufio"
	"regexp"
	"time"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
)

type nsStatus int

const (
	nsStatusNotExist nsStatus = iota
	nsStatusTerminating
	nsStatusActive
	nsStatusUnknown
)

// Namespace operations
type Namespace struct {
	context *Context
}

// Create namespace in context
func (n *Namespace) Create() (exists bool, err error) {
	namespaceStatus, err := n.getStatus()
	if err != nil {
		return false, err
	}
	if namespaceStatus == nsStatusActive {
		exists = true
		return
	}
	if namespaceStatus == nsStatusTerminating {
		err = n.waitForTerminate()
		if err != nil {
			return false, err
		}
	}
	exists = false
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
	status, err := n.getStatus()
	if err != nil {
		return false, err
	}
	switch status {
	case nsStatusActive:
		return true, nil
	case nsStatusUnknown:
		return false, stacktrace.Propagate(ErrUnknownStatus{}, "unknown status")
	default:
		return false, nil
	}
}

// Delete .
func (n *Namespace) Delete() (exists bool, err error) {
	namespaceStatus, err := n.getStatus()
	if err != nil {
		return false, err
	}
	if namespaceStatus == nsStatusNotExist {
		exists = false
		return
	}
	if namespaceStatus == nsStatusTerminating {
		err = n.waitForTerminate()
		if err != nil {
			return false, err
		}
		exists = false
		return
	}
	exists = true
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

func (n *Namespace) getStatus() (nsStatus, error) {
	args := n.context.completeArgsWithoutNamespace([]string{"get", "ns"})
	status, err := utils.NewCommand("kubectl", args...).CombineOutput().Run()
	if err != nil {
		return nsStatusUnknown, err
	}
	if status.ExitCode != 0 {
		return nsStatusUnknown, stacktrace.Propagate(commandErrorFromStatus(status), "command execute error")
	}
	splitRegex := regexp.MustCompile("\\s+")
	scanner := bufio.NewScanner(status.Stdout)
	firstLine := false
	namespaceStatus := nsStatusNotExist
	for scanner.Scan() {
		if firstLine {
			firstLine = true
			continue
		}
		parts := splitRegex.Split(scanner.Text(), -1)
		if len(parts) != 3 {
			continue
		}
		if parts[0] == n.context.namespace {
			switch parts[1] {
			case "Active":
				namespaceStatus = nsStatusActive
			case "Terminating":
				namespaceStatus = nsStatusTerminating
			default:
				namespaceStatus = nsStatusUnknown
			}
			break
		}
	}
	err = scanner.Err()
	if err != nil {
		return nsStatusUnknown, stacktrace.Propagate(err, "cannot read from command output")
	}
	return namespaceStatus, nil
}

func (n *Namespace) waitForTerminate() error {
	check := 0
	for {
		namespaceStatus, err := n.getStatus()
		if err != nil {
			return err
		}
		switch namespaceStatus {
		case nsStatusTerminating:
			time.Sleep(defaultTerminateInterval)
			check++
			if check > defaultTerminateCheckLimit {
				return stacktrace.Propagate(ErrTimeout{}, "timeout waiting for terminating namespace %q", n.context.namespace)
			}
		case nsStatusUnknown:
			return stacktrace.Propagate(ErrUnknownStatus{}, "unknown status")
		default:
			return nil
		}
	}
}

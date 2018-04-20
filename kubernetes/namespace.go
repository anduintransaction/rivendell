package kubernetes

import (
	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
)

// Namespace operations
type Namespace struct {
	context *Context
}

// Create namespace in context
func (n *Namespace) Create() (exists bool, err error) {
	status, err := n.getStatus()
	if err != nil {
		return false, err
	}
	if status == RsStatusUnknown {
		return false, stacktrace.Propagate(ErrUnknownStatus{n.context.namespace, "namespace", status}, "unknown status")
	}
	if status == RsStatusActive || status == RsStatusPending {
		exists = true
		return
	}
	if status == RsStatusTerminating {
		err = n.context.waitForNonPodTerminate(n.context.namespace, "namespace")
		if err != nil {
			return false, err
		}
	}
	exists = false
	args := n.context.completeArgsWithoutNamespace([]string{"create", "ns", n.context.namespace})
	cmdResult, err := utils.ExecuteCommand("kubectl", args...)
	if err != nil {
		return
	}
	if cmdResult.ExitCode != 0 {
		return false, stacktrace.Propagate(ErrCommandExitCode{cmdResult.ExitCode}, "command execute error")
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
	case RsStatusActive:
		return true, nil
	case RsStatusUnknown:
		return false, stacktrace.Propagate(ErrUnknownStatus{n.context.namespace, "namespace", status}, "unknown status")
	default:
		return false, nil
	}
}

// Delete .
func (n *Namespace) Delete() (exists bool, err error) {
	status, err := n.getStatus()
	if err != nil {
		return false, err
	}
	if status == RsStatusNotExist {
		exists = false
		return
	}
	if status == RsStatusTerminating {
		err = n.context.waitForNonPodTerminate(n.context.namespace, "namespace")
		if err != nil {
			return false, err
		}
		exists = false
		return
	}
	exists = true
	args := n.context.completeArgsWithoutNamespace([]string{"delete", "ns", n.context.namespace})
	cmdResult, err := utils.ExecuteCommand("kubectl", args...)
	if err != nil {
		return
	}
	if cmdResult.ExitCode != 0 {
		return false, stacktrace.Propagate(ErrCommandExitCode{cmdResult.ExitCode}, "command execute error")
	}
	return
}

func (n *Namespace) getStatus() (RsStatus, error) {
	return n.context.getNonPodStatus(n.context.namespace, "namespace")
}

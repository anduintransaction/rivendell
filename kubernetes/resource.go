package kubernetes

import (
	"strings"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
)

// Resource operations
type Resource struct {
	context *Context
}

// Create resource
func (r *Resource) Create(name, kind, rawContent string) (exists bool, err error) {
	kind = strings.ToLower(kind)
	status, err := r.getStatus(name, kind)
	if err != nil {
		return false, err
	}
	if status == rsStatusUnknown {
		return false, stacktrace.Propagate(ErrUnknownStatus{}, "unknown status")
	}
	if status == rsStatusActive || status == rsStatusPending {
		exists = true
		return
	}
	if status == rsStatusTerminating {
		err = r.waitForTerminating(name, kind)
		if err != nil {
			return false, err
		}
	}
	exists = false
	args := r.context.completeArgs([]string{"apply", "-f", "-"})
	cmd := utils.NewCommand("kubectl", args...)
	cmd.RedirectToStandard()
	cmd.SetStdin([]byte(rawContent))
	cmdResult, err := cmd.Run()
	if err != nil {
		return
	}
	if cmdResult.ExitCode != 0 {
		err = ErrCommandExitCode{cmdResult.ExitCode}
		return
	}
	return
}

// Exists check
func (r *Resource) Exists(name, kind string) (exists bool, err error) {
	status, err := r.getStatus(name, kind)
	if err != nil {
		return false, err
	}
	switch status {
	case rsStatusActive, rsStatusPending:
		return true, nil
	case rsStatusNotExist, rsStatusTerminating:
		return false, nil
	default:
		return false, stacktrace.Propagate(ErrUnknownStatus{}, "unknown status")
	}
}

// Delete .
func (r *Resource) Delete(name, kind string) (exists bool, err error) {
	kind = strings.ToLower(kind)
	status, err := r.getStatus(name, kind)
	if err != nil {
		return false, err
	}
	if status == rsStatusUnknown {
		return false, stacktrace.Propagate(ErrUnknownStatus{}, "unknown status")
	}
	if status == rsStatusNotExist || status == rsStatusTerminating {
		exists = false
		return
	}
	if status == rsStatusPending {
		err = r.waitForPending(name, kind)
		if err != nil {
			return false, err
		}
	}
	exists = true
	args := r.context.completeArgs([]string{"delete", kind, name})
	cmdResult, err := utils.ExecuteCommand("kubectl", args...)
	if err != nil {
		return
	}
	if cmdResult.ExitCode != 0 {
		err = ErrCommandExitCode{cmdResult.ExitCode}
		return
	}
	return
}

func (r *Resource) getStatus(name, kind string) (rsStatus, error) {
	if kind != "pod" {
		return r.context.getNonPodStatus(name, kind)
	}
	return rsStatusUnknown, nil
}

func (r *Resource) waitForPending(name, kind string) error {
	return nil
}

func (r *Resource) waitForTerminating(name, kind string) error {
	return nil
}

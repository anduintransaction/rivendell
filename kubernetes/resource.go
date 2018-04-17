package kubernetes

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
	yaml "gopkg.in/yaml.v2"
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
		return false, stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
	}
	if status == rsStatusActive || status == rsStatusPending || status == rsStatusSucceeded || status == rsStatusFailed {
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
	kind = strings.ToLower(kind)
	status, err := r.getStatus(name, kind)
	if err != nil {
		return false, err
	}
	switch status {
	case rsStatusActive, rsStatusPending, rsStatusSucceeded, rsStatusFailed:
		return true, nil
	case rsStatusNotExist, rsStatusTerminating:
		return false, nil
	default:
		return false, stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
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
		return false, stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
	}
	if status == rsStatusNotExist || status == rsStatusTerminating {
		exists = false
		return
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

// Wait .
func (r *Resource) Wait(name, kind string) (success bool, err error) {
	kind = strings.ToLower(kind)
	if kind != "pod" && kind != "job" {
		return false, stacktrace.Propagate(ErrUnsupportedKind{kind}, "unsupported kind")
	}
	waitDelay := 5 * time.Second
	for {
		status, err := r.getStatus(name, kind)
		if err != nil {
			return false, err
		}
		switch status {
		case rsStatusNotExist:
			return false, stacktrace.Propagate(ErrNotExist{name, kind}, "not exist")
		case rsStatusActive, rsStatusPending, rsStatusTerminating:
			time.Sleep(waitDelay)
		case rsStatusSucceeded:
			return true, nil
		case rsStatusFailed:
			return false, nil
		default:
			return false, stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
		}
	}
}

func (r *Resource) getStatus(name, kind string) (rsStatus, error) {
	switch kind {
	case "pod":
		return r.getPodStatus(name)
	case "job":
		return r.getJobStatus(name)
	default:
		return r.context.getNonPodStatus(name, kind)
	}
}

func (r *Resource) getPodStatus(name string) (rsStatus, error) {
	args := r.context.completeArgs([]string{"get", "pod", name, "-o", "yaml"})
	cmdResult, err := utils.NewCommand("kubectl", args...).Run()
	if err != nil {
		return rsStatusUnknown, err
	}
	if cmdResult.ExitCode != 0 {
		output, err := ioutil.ReadAll(cmdResult.Stderr)
		if err != nil {
			return rsStatusUnknown, stacktrace.Propagate(err, "cannot read stderr")
		}
		errOutput := string(output)
		if strings.Contains(errOutput, "(NotFound)") {
			return rsStatusNotExist, nil
		}
		return rsStatusUnknown, stacktrace.Propagate(ErrCommandExecute{cmdResult.ExitCode, errOutput}, "error execute command")
	}
	output, err := ioutil.ReadAll(cmdResult.Stdout)
	if err != nil {
		return rsStatusUnknown, stacktrace.Propagate(err, "cannot read stdout")
	}
	podInfo := &podResourceInfo{}
	err = yaml.Unmarshal(output, podInfo)
	if err != nil {
		return rsStatusUnknown, stacktrace.Propagate(ErrInvalidResponse{err, string(output)}, "invalid response")
	}
	if podInfo.Status == nil || podInfo.Metadata == nil {
		return rsStatusUnknown, nil
	}
	switch podInfo.Status.Phase {
	case "Pending":
		if podInfo.Metadata.DeletionTimestamp == "" {
			return rsStatusPending, nil
		}
		return rsStatusTerminating, nil
	case "Running":
		if podInfo.Metadata.DeletionTimestamp == "" {
			return rsStatusActive, nil
		}
		return rsStatusTerminating, nil
	case "Succeeded":
		return rsStatusSucceeded, nil
	case "Failed":
		return rsStatusFailed, nil
	default:
		return rsStatusUnknown, nil
	}
}

func (r *Resource) getJobStatus(name string) (rsStatus, error) {
	args := r.context.completeArgs([]string{"get", "job", name, "-o", "yaml"})
	cmdResult, err := utils.NewCommand("kubectl", args...).Run()
	if err != nil {
		return rsStatusUnknown, err
	}
	if cmdResult.ExitCode != 0 {
		output, _ := ioutil.ReadAll(cmdResult.Stderr)
		errOutput := string(output)
		if strings.Contains(errOutput, "(NotFound)") {
			return rsStatusNotExist, nil
		}
		return rsStatusUnknown, stacktrace.Propagate(ErrCommandExecute{cmdResult.ExitCode, errOutput}, "error execute command")
	}
	output, err := ioutil.ReadAll(cmdResult.Stdout)
	if err != nil {
		return rsStatusUnknown, stacktrace.Propagate(err, "cannot read stdout")
	}
	jobInfo := &jobResourceInfo{}
	err = yaml.Unmarshal(output, jobInfo)
	if err != nil {
		return rsStatusUnknown, stacktrace.Propagate(ErrInvalidResponse{err, string(output)}, "invalid response")
	}
	if jobInfo.Status == nil {
		return rsStatusUnknown, nil
	}
	if len(jobInfo.Status.Conditions) == 0 {
		return rsStatusActive, nil
	}
	condition := jobInfo.Status.Conditions[0]
	switch condition.Type {
	case "Complete":
		return rsStatusSucceeded, nil
	case "Failed":
		return rsStatusFailed, nil
	}
	return rsStatusUnknown, nil
}

func (r *Resource) waitForPending(name, kind string) error {
	check := 0
	for {
		status, err := r.getStatus(name, kind)
		if err != nil {
			return err
		}
		switch status {
		case rsStatusPending:
			time.Sleep(defaultPendingInterval)
			check++
			if check > defaultPendingCheckLimit {
				return ErrTimeout{}
			}
		case rsStatusUnknown:
			return stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
		default:
			return nil
		}
	}
}

func (r *Resource) waitForTerminating(name, kind string) error {
	check := 0
	for {
		status, err := r.getStatus(name, kind)
		if err != nil {
			return err
		}
		switch status {
		case rsStatusTerminating:
			time.Sleep(defaultTerminateInterval)
			check++
			if check > defaultTerminateCheckLimit {
				return stacktrace.Propagate(ErrTimeout{}, "timeout waiting for terminating %s %q", kind, name)
			}
		case rsStatusUnknown:
			return stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
		default:
			return nil
		}
	}
}

func (r *Resource) waitForPod(name string) (success bool, err error) {
	waitDelay := 5 * time.Second
	for {
		status, err := r.getStatus(name, "pod")
		if err != nil {
			return false, err
		}
		switch status {
		case rsStatusNotExist:
			return false, stacktrace.Propagate(ErrNotExist{name, "pod"}, "not exist")
		case rsStatusActive, rsStatusPending, rsStatusTerminating:
			time.Sleep(waitDelay)
		case rsStatusSucceeded:
			return true, nil
		case rsStatusFailed:
			return false, nil
		default:
			return false, stacktrace.Propagate(ErrUnknownStatus{name, "pod", status}, "unknown status")
		}
	}
}

func (r *Resource) waitForJob(name string) (success bool, err error) {
	return false, nil
}

type podResourceInfo struct {
	Metadata *podMetadata `yaml:"metadata"`
	Status   *podStatus   `yaml:"status"`
}

type podMetadata struct {
	DeletionTimestamp string `yaml:"deletionTimestamp"`
}

type podStatus struct {
	Phase string `yaml:"phase"`
}

type jobResourceInfo struct {
	Status *jobStatus `yaml:"status"`
}

type jobStatus struct {
	Conditions []*jobCondition `yaml:"conditions"`
}

type jobCondition struct {
	Type string `yaml:"type"`
}

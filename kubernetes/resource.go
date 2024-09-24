package kubernetes

import (
	"io"
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
	status, err := r.GetStatus(name, kind)
	if err != nil {
		return false, err
	}
	if status == RsStatusUnknown {
		return false, stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
	}
	if status == RsStatusActive || status == RsStatusPending || status == RsStatusSucceeded || status == RsStatusFailed {
		exists = true
		return
	}
	if status == RsStatusTerminating {
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
	status, err := r.GetStatus(name, kind)
	if err != nil {
		return false, err
	}
	switch status {
	case RsStatusActive, RsStatusPending, RsStatusSucceeded, RsStatusFailed:
		return true, nil
	case RsStatusNotExist, RsStatusTerminating:
		return false, nil
	default:
		return false, stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
	}
}

// Delete .
func (r *Resource) Delete(name, kind string) (exists bool, err error) {
	kind = strings.ToLower(kind)
	status, err := r.GetStatus(name, kind)
	if err != nil {
		return false, err
	}
	if status == RsStatusUnknown {
		return false, stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
	}
	if status == RsStatusNotExist || status == RsStatusTerminating {
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

// UpdateStatus .
type UpdateStatus int

// UpdateStatus values
const (
	UpdateStatusNotExist UpdateStatus = iota
	UpdateStatusExisted
	UpdateStatusSkipped
)

// Update .
func (r *Resource) Update(name, kind, rawContent string) (updateStatus UpdateStatus, err error) {
	kind = strings.ToLower(kind)
	if kind == "pod" || kind == "job" {
		return UpdateStatusSkipped, nil
	}
	status, err := r.GetStatus(name, kind)
	if err != nil {
		return UpdateStatusNotExist, err
	}
	if status == RsStatusUnknown {
		return UpdateStatusNotExist, stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
	}
	if status == RsStatusNotExist || status == RsStatusTerminating {
		return UpdateStatusNotExist, nil
	}
	updateStatus = UpdateStatusExisted
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

// Upgrade .
func (r *Resource) Upgrade(name, kind, rawContent string) (updateStatus UpdateStatus, err error) {
	kind = strings.ToLower(kind)
	status, err := r.GetStatus(name, kind)
	if err != nil {
		return UpdateStatusNotExist, err
	}
	if status == RsStatusUnknown {
		return UpdateStatusNotExist, stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
	}
	if (kind == "pod" || kind == "job") && (status == RsStatusActive || status == RsStatusPending) {
		return UpdateStatusSkipped, nil
	}
	if status == RsStatusNotExist || status == RsStatusTerminating {
		updateStatus = UpdateStatusNotExist
	} else {
		updateStatus = UpdateStatusExisted
	}
	if kind == "pod" || kind == "job" {
		_, err = r.Delete(name, kind)
		if err != nil {
			return UpdateStatusNotExist, err
		}
	}
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

// Wait .
func (r *Resource) Wait(name, kind string) (success bool, err error) {
	kind = strings.ToLower(kind)
	switch kind {
	case "pod":
		fallthrough
	case "job":
		return r.waitByObjStatus(name, kind)
	case "deploy":
		fallthrough
	case "deployment":
		return r.waitByRolloutStatus(name, kind)
	default:
		return false, stacktrace.Propagate(ErrUnsupportedKind{kind}, "unsupported kind")
	}
}

func (r *Resource) waitByRolloutStatus(name, kind string) (bool, error) {
	args := r.context.completeArgs([]string{"rollout", "status", kind, name})
	cmd := utils.NewCommand("kubectl", args...)
	cmd.RedirectToStandard()
	cmdResult, err := cmd.Run()
	if err != nil {
		return false, err
	}
	if cmdResult.ExitCode != 0 {
		return false, ErrCommandExitCode{cmdResult.ExitCode}
	}
	return true, nil
}

func (r *Resource) waitByObjStatus(name, kind string) (bool, error) {
	waitDelay := 5 * time.Second
	for {
		status, err := r.GetStatus(name, kind)
		if err != nil {
			return false, err
		}
		switch status {
		case RsStatusNotExist:
			return false, stacktrace.Propagate(ErrNotExist{name, kind}, "not exist")
		case RsStatusActive, RsStatusPending, RsStatusTerminating:
			time.Sleep(waitDelay)
		case RsStatusSucceeded:
			return true, nil
		case RsStatusFailed:
			return false, nil
		default:
			return false, stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
		}
	}
}

// Logs .
func (r *Resource) Logs(name, containerName string, stdout io.Writer, stderr io.Writer) error {
	kind := "pod"
	status, err := r.GetStatus(name, kind)
	if err != nil {
		return err
	}
	switch status {
	case RsStatusNotExist, RsStatusTerminating:
		return stacktrace.Propagate(ErrNotExist{name, kind}, "not exist")
	case RsStatusUnknown:
		return stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
	case RsStatusPending:
		err = r.waitForPending(name, kind)
		if err != nil {
			return err
		}
	}
	if containerName == "" {
		containerName, err = r.getFirstContainerName(kind, name)
		if err != nil {
			return err
		}
	}
	firstRun := true
	for {
		var args []string
		if firstRun {
			args = r.context.completeArgs([]string{"logs", "-f", "-c", containerName, name})
			firstRun = false
		} else {
			args = r.context.completeArgs([]string{"logs", "-f", "-c", containerName, "--tail", "10", name})
		}
		cmd := utils.NewCommand("kubectl", args...)
		cmd.SetStdout(stdout)
		cmd.SetStderr(stderr)
		cmdResult, err := cmd.Run()
		if err != nil {
			return err
		}
		if cmdResult.ExitCode == 0 {
			break
		}
	}
	return nil
}

// GetStatus .
func (r *Resource) GetStatus(name, kind string) (RsStatus, error) {
	switch kind {
	case "pod":
		return r.getPodStatus(name)
	case "job":
		return r.getJobStatus(name)
	default:
		return r.context.getNonPodStatus(name, kind)
	}
}

func (r *Resource) getPodStatus(name string) (RsStatus, error) {
	args := r.context.completeArgs([]string{"get", "pod", name, "-o", "yaml"})
	cmdResult, err := utils.NewCommand("kubectl", args...).Run()
	if err != nil {
		return RsStatusUnknown, err
	}
	if cmdResult.ExitCode != 0 {
		output, err := ioutil.ReadAll(cmdResult.Stderr)
		if err != nil {
			return RsStatusUnknown, stacktrace.Propagate(err, "cannot read stderr")
		}
		errOutput := string(output)
		if strings.Contains(errOutput, "(NotFound)") {
			return RsStatusNotExist, nil
		}
		return RsStatusUnknown, stacktrace.Propagate(ErrCommandExecute{cmdResult.ExitCode, errOutput}, "error execute command")
	}
	output, err := ioutil.ReadAll(cmdResult.Stdout)
	if err != nil {
		return RsStatusUnknown, stacktrace.Propagate(err, "cannot read stdout")
	}
	podInfo := &podResourceInfo{}
	err = yaml.Unmarshal(output, podInfo)
	if err != nil {
		return RsStatusUnknown, stacktrace.Propagate(ErrInvalidResponse{err, string(output)}, "invalid response")
	}
	if podInfo.Status == nil || podInfo.Metadata == nil {
		return RsStatusUnknown, nil
	}
	switch podInfo.Status.Phase {
	case "Pending":
		if podInfo.Metadata.DeletionTimestamp == "" {
			return RsStatusPending, nil
		}
		return RsStatusTerminating, nil
	case "Running":
		if podInfo.Metadata.DeletionTimestamp == "" {
			return RsStatusActive, nil
		}
		return RsStatusTerminating, nil
	case "Succeeded":
		return RsStatusSucceeded, nil
	case "Failed":
		return RsStatusFailed, nil
	default:
		return RsStatusUnknown, nil
	}
}

func (r *Resource) getJobStatus(name string) (RsStatus, error) {
	args := r.context.completeArgs([]string{"get", "job", name, "-o", "yaml"})
	cmdResult, err := utils.NewCommand("kubectl", args...).Run()
	if err != nil {
		return RsStatusUnknown, err
	}
	if cmdResult.ExitCode != 0 {
		output, _ := ioutil.ReadAll(cmdResult.Stderr)
		errOutput := string(output)
		if strings.Contains(errOutput, "(NotFound)") {
			return RsStatusNotExist, nil
		}
		return RsStatusUnknown, stacktrace.Propagate(ErrCommandExecute{cmdResult.ExitCode, errOutput}, "error execute command")
	}
	output, err := ioutil.ReadAll(cmdResult.Stdout)
	if err != nil {
		return RsStatusUnknown, stacktrace.Propagate(err, "cannot read stdout")
	}
	jobInfo := &jobResourceInfo{}
	err = yaml.Unmarshal(output, jobInfo)
	if err != nil {
		return RsStatusUnknown, stacktrace.Propagate(ErrInvalidResponse{err, string(output)}, "invalid response")
	}
	if jobInfo.Status == nil {
		return RsStatusUnknown, nil
	}
	if len(jobInfo.Status.Conditions) == 0 {
		return RsStatusActive, nil
	}
	condition := jobInfo.Status.Conditions[0]
	switch condition.Type {
	case "Complete":
		return RsStatusSucceeded, nil
	case "SuccessCriteriaMet":
		return RsStatusSucceeded, nil
	case "Failed":
		return RsStatusFailed, nil
	}
	return RsStatusUnknown, nil
}

func (r *Resource) waitForPending(name, kind string) error {
	check := 0
	for {
		status, err := r.GetStatus(name, kind)
		if err != nil {
			return err
		}
		switch status {
		case RsStatusPending:
			time.Sleep(defaultPendingInterval)
			check++
			if check > defaultPendingCheckLimit {
				return ErrTimeout{}
			}
		case RsStatusUnknown:
			return stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
		default:
			return nil
		}
	}
}

func (r *Resource) waitForTerminating(name, kind string) error {
	check := 0
	for {
		status, err := r.GetStatus(name, kind)
		if err != nil {
			return err
		}
		switch status {
		case RsStatusTerminating:
			time.Sleep(defaultTerminateInterval)
			check++
			if check > defaultTerminateCheckLimit {
				return stacktrace.Propagate(ErrTimeout{}, "timeout waiting for terminating %s %q", kind, name)
			}
		case RsStatusUnknown:
			return stacktrace.Propagate(ErrUnknownStatus{name, kind, status}, "unknown status")
		default:
			return nil
		}
	}
}

func (r *Resource) getFirstContainerName(kind, name string) (containerName string, err error) {
	kind = strings.ToLower(kind)
	switch kind {
	case "pod":
		return r.getFirstContainerNameFromPod(name)
	default:
		return "", stacktrace.Propagate(ErrUnsupportedKind{kind}, "unsupported kind")
	}
}

func (r *Resource) getFirstContainerNameFromPod(name string) (containerName string, err error) {
	args := r.context.completeArgs([]string{"get", "pod", name, "-o", "jsonpath={.spec.containers[0].name}"})
	cmdResult, err := utils.NewCommand("kubectl", args...).Run()
	if err != nil {
		return "", err
	}
	if cmdResult.ExitCode != 0 {
		errOutput, _ := ioutil.ReadAll(cmdResult.Stderr)
		return "", stacktrace.Propagate(ErrCommandExecute{cmdResult.ExitCode, string(errOutput)}, "error execute command")
	}
	output, err := ioutil.ReadAll(cmdResult.Stdout)
	if err != nil {
		return "", stacktrace.Propagate(err, "cannot read stdout")
	}
	return strings.TrimSpace(string(output)), nil
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

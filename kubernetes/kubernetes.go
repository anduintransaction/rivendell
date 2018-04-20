package kubernetes

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
	yaml "gopkg.in/yaml.v2"
)

// Context .
type Context struct {
	namespace  string
	context    string
	kubeConfig string
}

// NewContext .
func NewContext(namespace, context, kubeConfig string) (*Context, error) {
	c := &Context{namespace, context, kubeConfig}
	err := c.checkDeps()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Namespace .
func (c *Context) Namespace() *Namespace {
	return &Namespace{c}
}

// Resource .
func (c *Context) Resource() *Resource {
	return &Resource{c}
}

func (c *Context) checkDeps() error {
	status, err := utils.ExecuteCommandSilently("which", "kubectl")
	if err != nil {
		return nil
	}
	if status.ExitCode != 0 {
		return stacktrace.Propagate(ErrMissingCommand{"kubectl"}, "missing command %q", "kubectl")
	}
	return nil
}

func (c *Context) completeArgsWithoutNamespace(args []string) []string {
	if c.context != "" {
		args = append(args, "--context", c.context)
	}
	if c.kubeConfig != "" {
		args = append(args, "--kubeconfig", c.kubeConfig)
	}
	return args
}

func (c *Context) completeArgs(args []string) []string {
	if c.namespace != "" {
		args = append(args, "-n", c.namespace)
	}
	return c.completeArgsWithoutNamespace(args)
}

func (c *Context) getNonPodStatus(name, kind string) (RsStatus, error) {
	var args []string
	if kind == "namespace" {
		args = c.completeArgsWithoutNamespace([]string{"get", kind, name, "-o", "yaml"})
	} else {
		args = c.completeArgs([]string{"get", kind, name, "-o", "yaml"})
	}
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
	output, _ := ioutil.ReadAll(cmdResult.Stdout)
	rsInfo := &kubernetesResourceInfo{}
	err = yaml.Unmarshal(output, rsInfo)
	if err != nil {
		return RsStatusUnknown, stacktrace.Propagate(ErrInvalidResponse{err, string(output)}, "invalid response")
	}
	if rsInfo.Status == nil {
		// static resources like configmaps
		return RsStatusActive, nil
	}
	switch rsInfo.Status.Phase {
	case "Active", "":
		return RsStatusActive, nil
	case "Terminating":
		return RsStatusTerminating, nil
	default:
		return RsStatusUnknown, nil
	}
}

func (c *Context) waitForNonPodTerminate(name, kind string) error {
	check := 0
	for {
		status, err := c.getNonPodStatus(name, kind)
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

type kubernetesResourceInfo struct {
	Status *kubernetesResourceStatus `yaml:"status"`
}

type kubernetesResourceStatus struct {
	Phase string `yaml:"phase"`
}

// RsStatus .
type RsStatus int

func (status RsStatus) String() string {
	switch status {
	case RsStatusUnknown:
		return "Unknown"
	case RsStatusNotExist:
		return "NotExist"
	case RsStatusPending:
		return "Pending"
	case RsStatusActive:
		return "Active"
	case RsStatusTerminating:
		return "Terminating"
	case RsStatusSucceeded:
		return "Succeeded"
	case RsStatusFailed:
		return "Failed"
	default:
		return ""
	}
}

// Status code
const (
	RsStatusUnknown RsStatus = iota
	RsStatusNotExist
	RsStatusPending
	RsStatusActive
	RsStatusTerminating
	RsStatusSucceeded
	RsStatusFailed
)

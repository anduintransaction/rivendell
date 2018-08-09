package kubernetes

import (
	"bufio"
	"io/ioutil"
	"strings"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
	yaml "gopkg.in/yaml.v2"
)

// Service operations
type Service struct {
	context *Context
}

func (s *Service) ListPods(name string) ([]string, error) {
	status, err := s.context.getNonPodStatus(name, "service")
	if err != nil {
		return nil, err
	}
	if status != RsStatusActive {
		// We ignore errors here.
		return nil, nil
	}
	cmdResult, err := utils.NewCommand("kubectl", s.context.completeArgs([]string{"get", "services", name, "-o", "yaml"})...).Run()
	if err != nil {
		return nil, err
	}
	if cmdResult.ExitCode != 0 {
		output, _ := ioutil.ReadAll(cmdResult.Stderr)
		errOutput := string(output)
		return nil, stacktrace.Propagate(ErrCommandExecute{cmdResult.ExitCode, errOutput}, "error execute command")
	}
	output, _ := ioutil.ReadAll(cmdResult.Stdout)
	serviceInfo := &serviceResourceInfo{}
	err = yaml.Unmarshal(output, serviceInfo)
	if err != nil {
		return nil, stacktrace.Propagate(ErrInvalidResponse{err, string(output)}, "invalid response")
	}
	if serviceInfo.Spec == nil || len(serviceInfo.Spec.Selector) == 0 {
		return nil, nil
	}
	selectors := []string{}
	for k, v := range serviceInfo.Spec.Selector {
		selectors = append(selectors, k+"="+v)
	}
	args := []string{"get", "pod", "-l", strings.Join(selectors, ","), "-o", "name"}
	cmdResult, err = utils.NewCommand("kubectl", s.context.completeArgs(args)...).Run()
	if err != nil {
		return nil, err
	}
	if cmdResult.ExitCode != 0 {
		output, _ := ioutil.ReadAll(cmdResult.Stderr)
		errOutput := string(output)
		return nil, stacktrace.Propagate(ErrCommandExecute{cmdResult.ExitCode, errOutput}, "error execute command")
	}
	pods := []string{}
	scanner := bufio.NewScanner(cmdResult.Stdout)
	for scanner.Scan() {
		pods = append(pods, strings.TrimPrefix(scanner.Text(), "pod/"))
	}
	return pods, nil
}

type serviceResourceInfo struct {
	Spec *serviceSpec `yaml:"spec"`
}

type serviceSpec struct {
	Selector map[string]string `yaml:"selector"`
}

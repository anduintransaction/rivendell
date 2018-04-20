package project

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/anduintransaction/rivendell/kubernetes"
	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CommandTestSuite struct {
	suite.Suite
	resourceRoot  string
	testNamespace string
}

func (s *CommandTestSuite) SetupSuite() {
	s.resourceRoot = "../test-resources"
	s.testNamespace = fmt.Sprintf("test-ns-%d", time.Now().Unix())
}

func (s *CommandTestSuite) TestUpAndDown() {
	if !utils.TestEnable() {
		fmt.Println("Skipping up and down test")
		return
	}
	projectFile := filepath.Join(s.resourceRoot, "command-test", "up-down", "project.yml")
	namespace := s.testNamespace
	context := ""
	kubeConfig := ""
	variables := make(map[string]string)
	variableFiles := []string{}
	project, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = project.Up()
	require.Nil(s.T(), err)
	kubeContext, err := kubernetes.NewContext(project.namespace, project.context, project.kubeConfig)
	require.Nil(s.T(), err)
	project.resourceGraph.WalkForward(func(g *ResourceGroup) error {
		for _, r := range g.allResources() {
			exists, err := kubeContext.Resource().Exists(r.Name, r.Kind)
			require.Nil(s.T(), err)
			require.True(s.T(), exists)
		}
		return nil
	})
	err = project.Down()
	require.Nil(s.T(), err)
}

func (s *CommandTestSuite) TestUpdate() {
	if !utils.TestEnable() {
		fmt.Println("Skipping update test")
		return
	}
	projectFile := filepath.Join(s.resourceRoot, "command-test", "update", "project.yml")
	namespace := s.testNamespace
	context := ""
	kubeConfig := ""
	variables := map[string]string{
		"tag": "1.13.12",
	}
	variableFiles := []string{}
	project, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = project.Up()
	require.Nil(s.T(), err)
	variables["tag"] = "1.13"
	updatedProject, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = updatedProject.Update()
	require.Nil(s.T(), err)
	err = project.Down()
	require.Nil(s.T(), err)
}

func (s *CommandTestSuite) TestUpgrade() {
	if !utils.TestEnable() {
		fmt.Println("Skipping upgrade test")
		return
	}
	projectFile := filepath.Join(s.resourceRoot, "command-test", "upgrade", "project.yml")
	namespace := s.testNamespace
	context := ""
	kubeConfig := ""
	variables := map[string]string{
		"nginxTag":  "1.13.12",
		"ubuntuTag": "16.04",
	}
	variableFiles := []string{}
	project, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = project.Up()
	require.Nil(s.T(), err)
	err = Wait(namespace, context, kubeConfig, "job", "success", 60)
	require.Nil(s.T(), err)
	variables["nginxTag"] = "1.13"
	variables["ubuntuTag"] = "16.10"
	updatedProject, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = updatedProject.Upgrade()
	require.Nil(s.T(), err)
	err = Wait(namespace, context, kubeConfig, "job", "success", 60)
	require.Nil(s.T(), err)
	err = project.Down()
	require.Nil(s.T(), err)
}

func (s *CommandTestSuite) TestWaitPod() {
	if !utils.TestEnable() {
		fmt.Println("Skipping wait pod")
		return
	}
	projectFile := filepath.Join(s.resourceRoot, "command-test", "wait-pod", "project.yml")
	namespace := s.testNamespace
	context := ""
	kubeConfig := ""
	variables := make(map[string]string)
	variableFiles := []string{}
	p, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = p.Up()
	require.Nil(s.T(), err)
	kubeContext, err := kubernetes.NewContext(p.namespace, p.context, p.kubeConfig)
	require.Nil(s.T(), err)
	p.resourceGraph.WalkForward(func(g *ResourceGroup) error {
		for _, r := range g.allResources() {
			exists, err := kubeContext.Resource().Exists(r.Name, r.Kind)
			require.Nil(s.T(), err)
			require.True(s.T(), exists)
		}
		return nil
	})
	err = Wait(namespace, context, kubeConfig, "pod", "pod2", 300)
	require.Nil(s.T(), err)
	err = p.Down()
	require.Nil(s.T(), err)
}

func (s *CommandTestSuite) TestWaitJob() {
	if !utils.TestEnable() {
		fmt.Println("Skipping wait")
		return
	}
	projectFile := filepath.Join(s.resourceRoot, "command-test", "wait-job", "project.yml")
	namespace := s.testNamespace
	context := ""
	kubeConfig := ""
	variables := make(map[string]string)
	variableFiles := []string{}
	p, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = p.Up()
	require.Nil(s.T(), err)
	kubeContext, err := kubernetes.NewContext(p.namespace, p.context, p.kubeConfig)
	require.Nil(s.T(), err)
	p.resourceGraph.WalkForward(func(g *ResourceGroup) error {
		for _, r := range g.allResources() {
			exists, err := kubeContext.Resource().Exists(r.Name, r.Kind)
			require.Nil(s.T(), err)
			require.True(s.T(), exists)
		}
		return nil
	})
	err = Wait(namespace, context, kubeConfig, "job", "job2", 300)
	require.Nil(s.T(), err)
	err = p.Down()
	require.Nil(s.T(), err)
}

func (s *CommandTestSuite) TestWaitNotExists() {
	if !utils.TestEnable() {
		fmt.Println("Skipping test wait not exists")
		return
	}
	err := Wait("", "", "", "job", "not-exists", 0)
	require.NotNil(s.T(), err)
	_, ok := stacktrace.RootCause(err).(kubernetes.ErrNotExist)
	require.True(s.T(), ok)
}

func (s *CommandTestSuite) TestJobWaitTimeoutInProject() {
	if !utils.TestEnable() {
		fmt.Println("Skipping test wait timeout in project")
		return
	}
	projectFile := filepath.Join(s.resourceRoot, "command-test", "job-wait-timeout-in-project", "project.yml")
	namespace := s.testNamespace
	context := ""
	kubeConfig := ""
	variables := make(map[string]string)
	variableFiles := []string{}
	p, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = p.Up()
	require.NotNil(s.T(), err)
	_, ok := stacktrace.RootCause(err).(ErrWaitTimeout)
	require.True(s.T(), ok)
	err = p.Down()
	require.Nil(s.T(), err)
}

func (s *CommandTestSuite) TestJobWaitFailedInProject() {
	if !utils.TestEnable() {
		fmt.Println("Skipping test wait failed in project")
		return
	}
	projectFile := filepath.Join(s.resourceRoot, "command-test", "job-wait-failed-in-project", "project.yml")
	namespace := s.testNamespace
	context := ""
	kubeConfig := ""
	variables := make(map[string]string)
	variableFiles := []string{}
	p, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = p.Up()
	require.NotNil(s.T(), err)
	_, ok := stacktrace.RootCause(err).(ErrWaitFailed)
	require.True(s.T(), ok)
	err = p.Down()
	require.Nil(s.T(), err)
}

func (s *CommandTestSuite) TestPodWaitTimeoutInProject() {
	if !utils.TestEnable() {
		fmt.Println("Skipping test pod wait timeout in project")
		return
	}
	projectFile := filepath.Join(s.resourceRoot, "command-test", "pod-wait-timeout-in-project", "project.yml")
	namespace := s.testNamespace
	context := ""
	kubeConfig := ""
	variables := make(map[string]string)
	variableFiles := []string{}
	p, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = p.Up()
	require.NotNil(s.T(), err)
	_, ok := stacktrace.RootCause(err).(ErrWaitTimeout)
	require.True(s.T(), ok)
	err = p.Down()
	require.Nil(s.T(), err)
}

func (s *CommandTestSuite) TestPodWaitFailedInProject() {
	if !utils.TestEnable() {
		fmt.Println("Skipping test pod wait failed in project")
		return
	}
	projectFile := filepath.Join(s.resourceRoot, "command-test", "pod-wait-failed-in-project", "project.yml")
	namespace := s.testNamespace
	context := ""
	kubeConfig := ""
	variables := make(map[string]string)
	variableFiles := []string{}
	p, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err)
	err = p.Up()
	require.NotNil(s.T(), err)
	_, ok := stacktrace.RootCause(err).(ErrWaitFailed)
	require.True(s.T(), ok)
	err = p.Down()
	require.Nil(s.T(), err)
}

func TestCommand(t *testing.T) {
	suite.Run(t, new(CommandTestSuite))
}

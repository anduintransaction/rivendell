package project

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
)

// Project holds configuration for a rivendell task
type Project struct {
	rootDir       string
	namespace     string
	context       string
	kubeConfig    string
	prepullImage  bool
	variables     map[string]string
	credentials   []*DockerCredential
	resourceGraph *ResourceGraph
}

// ReadProject reads a project from file
func ReadProject(projectFile, namespace, context, kubeConfig string, variables map[string]string) (*Project, error) {
	projectConfig, err := ReadProjectConfig(projectFile, variables)
	if err != nil {
		return nil, err
	}
	project := &Project{
		context:     context,
		kubeConfig:  kubeConfig,
		credentials: []*DockerCredential{},
	}
	project.resolveProjectRoot(projectFile, projectConfig.RootDir)
	project.resolveNamespace(namespace, projectConfig.Namespace)
	project.resolveVariables(variables, projectConfig.Variables)
	err = project.resolveCredentials(projectConfig.Credentials)
	if err != nil {
		return nil, err
	}
	err = project.resolveResourceGraph(projectConfig.ResourceGroups)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (p *Project) resolveProjectRoot(projectFile, configRoot string) {
	projectFileDirname := filepath.Dir(projectFile)
	p.rootDir = filepath.Join(projectFileDirname, configRoot)
}

func (p *Project) resolveNamespace(namespaceFromCommand, namespaceFromConfig string) {
	if namespaceFromCommand != "" {
		p.namespace = namespaceFromCommand
	} else {
		p.namespace = namespaceFromConfig
	}
}

func (p *Project) resolveVariables(variablesFromCommand, variablesFromConfig map[string]string) {
	rivendellVariables := map[string]string{
		"rivendellVarNamespace":  p.namespace,
		"rivendellVarContext":    p.context,
		"rivendellVarKubeConfig": p.kubeConfig,
	}
	p.variables = utils.MergeMaps(variablesFromConfig, variablesFromCommand, rivendellVariables)
}

func (p *Project) resolveCredentials(credentials []*CredentialConfig) error {
	for _, credential := range credentials {
		dockerCredential := &DockerCredential{
			Username: credential.Username,
			Host:     credential.Host,
		}
		if credential.PasswordFile == "" {
			dockerCredential.Password = credential.Password
		} else {
			passwordFile := filepath.Join(p.rootDir, credential.PasswordFile)
			password, err := ioutil.ReadFile(passwordFile)
			if err != nil {
				return stacktrace.Propagate(err, "cannot read password file %q", passwordFile)
			}
			dockerCredential.Password = strings.TrimSpace(string(password))
		}
		p.credentials = append(p.credentials, dockerCredential)
	}
	return nil
}

func (p *Project) resolveResourceGraph(resourceGroupConfigs []*ResourceGroupConfig) error {
	resourceGraph, err := ReadResourceGraph(p.rootDir, resourceGroupConfigs, p.variables)
	if err != nil {
		return err
	}
	p.resourceGraph = resourceGraph
	return nil
}

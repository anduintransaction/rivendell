package project

import (
	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
	yaml "gopkg.in/yaml.v2"
)

// Config holds configuration data parsed from yaml file
type Config struct {
	RootDir        string                 `yaml:"root_dir"`
	Namespace      string                 `yaml:"namespace"`
	PrepullImages  bool                   `yaml:"prepull_images"`
	Variables      map[string]string      `yaml:"variables"`
	Credentials    []*CredentialConfig    `yaml:"credentials"`
	ResourceGroups []*ResourceGroupConfig `yaml:"resource_groups"`
}

// CredentialConfig holds configuration for docker registry credential
type CredentialConfig struct {
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	PasswordFile string `yaml:"password_file"`
	Host         string `yaml:"host"`
}

// ResourceGroupConfig holds configuration for resource group
type ResourceGroupConfig struct {
	Name      string   `yaml:"name"`
	Resources []string `yaml:"resources"`
	Excludes  []string `yaml:"excludes"`
	Depend    []string `yaml:"depend"`
	Wait      []string `yaml:"wait"`
}

// ReadProjectConfig .
func ReadProjectConfig(projectFile string, variables map[string]string) (*Config, error) {
	content, err := utils.ExecuteTemplate(projectFile, variables)
	if err != nil {
		return nil, err
	}
	projectConfig := &Config{}
	err = yaml.Unmarshal(content, projectConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "cannot parse yaml configuration")
	}
	return projectConfig, nil
}

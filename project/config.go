package project

import (
	"io"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
	yaml "gopkg.in/yaml.v2"
)

// Config holds configuration data parsed from yaml file
type Config struct {
	RootDir         string                 `yaml:"root_dir"`
	Namespace       string                 `yaml:"namespace"`
	Variables       map[string]string      `yaml:"variables"`
	ResourceGroups  []*ResourceGroupConfig `yaml:"resource_groups"`
	DeleteNamespace bool                   `yaml:"delete_namespace"`
}

// ResourceGroupConfig holds configuration for resource group
type ResourceGroupConfig struct {
	Name      string        `yaml:"name"`
	Resources []string      `yaml:"resources"`
	Excludes  []string      `yaml:"excludes"`
	Depend    []string      `yaml:"depend"`
	Wait      []*WaitConfig `yaml:"wait"`
}

// WaitConfig .
type WaitConfig struct {
	Name    string `yaml:"name"`
	Kind    string `yaml:"kind"`
	Timeout int    `yaml:"timeout"`
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

func (c *Config) Write(w io.Writer) error {
	out, err := yaml.Marshal(c)
	if err != nil {
		return stacktrace.Propagate(err, "cannot encode config")
	}
	_, err = w.Write(out)
	return stacktrace.Propagate(err, "cannot write config")
}

package project

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	resourceRoot string
}

func (s *ConfigTestSuite) SetupSuite() {
	s.resourceRoot = "../test-resources"
}

func (s *ConfigTestSuite) TestReadProjectConfig() {
	projectFile := filepath.Join(s.resourceRoot, "config-test", "happy-path", "project.yml")
	variables := map[string]string{
		"postgresImageTag": "9.6",
		"appTag":           "1.1.4",
	}
	projectConfig, err := ReadProjectConfig(projectFile, variables)
	require.Nil(s.T(), err, "should read project config successfully")
	expected := &Config{
		RootDir:   ".",
		Namespace: "coruscant",
		Variables: map[string]string{
			"postgresTag":          "9.6",
			"redisTag":             "4-alpine",
			"postgresSidecarImage": "postgres-sidecar:1.1.4",
			"redisSidecarImage":    "redis-sidecar:1.1.4",
		},
		ResourceGroups: []*ResourceGroupConfig{
			{
				Name:      "configs",
				Resources: []string{"./configs/*.yml"},
				Excludes:  []string{"./configs/*ignore*"},
			},
			{
				Name:      "secrets",
				Resources: []string{"./secrets/*.yml"},
			},
			{
				Name:      "databases",
				Resources: []string{"./databases/*.yml"},
				Depend:    []string{"configs", "secrets"},
			},
			{
				Name:      "init-jobs",
				Resources: []string{"./jobs/*.yml"},
				Depend:    []string{"databases"},
			},
			{
				Name:      "services",
				Resources: []string{"./services/*.yml"},
				Depend:    []string{"init-jobs"},
				Wait: []*WaitConfig{
					{
						Name: "init-postgres",
						Kind: "job",
					},
					{
						Name: "init-redis",
						Kind: "job",
					},
				},
			},
			{
				Name:      "nginx",
				Resources: []string{"https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/controllers/nginx-deployment.yaml"},
			},
		},
	}
	require.Equal(s.T(), *expected, *projectConfig)
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

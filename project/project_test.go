package project

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	suite.Suite
	resourceRoot string
}

func (s *ProjectTestSuite) SetupTest() {
	s.resourceRoot = "../test-resources"
}

func (s *ProjectTestSuite) TestReadProject() {
	projectDir := filepath.Join(s.resourceRoot, "config-test", "happy-path")
	projectFile := filepath.Join(projectDir, "project.yml")
	namespace := "alderaan"
	context := "test"
	kubeConfig := "/etc/kubernetes/default.yml"
	variables := map[string]string{
		"postgresImageTag": "9.6",
		"appTag":           "1.1.4",
	}
	os.Setenv("DOCKERHUB_USERNAME", "luke-skywalker")
	os.Setenv("DOCKERHUB_PASSWORD", "tatooine")
	project, err := ReadProject(projectFile, namespace, context, kubeConfig, variables)
	require.Nil(s.T(), err, "should read project file successfully")
	require.Equal(s.T(), projectDir, project.rootDir)
	require.Equal(s.T(), namespace, project.namespace)
	require.Equal(s.T(), context, project.context)
	require.Equal(s.T(), kubeConfig, project.kubeConfig)
	expectedVariables := map[string]string{
		"postgresTag":            "9.6",
		"postgresImageTag":       "9.6",
		"redisTag":               "4-alpine",
		"appTag":                 "1.1.4",
		"postgresSidecarImage":   "postgres-sidecar:1.1.4",
		"redisSidecarImage":      "redis-sidecar:1.1.4",
		"rivendellVarNamespace":  namespace,
		"rivendellVarContext":    context,
		"rivendellVarKubeConfig": kubeConfig,
	}
	require.Equal(s.T(), expectedVariables, project.variables)
	expectedCredentials := []*DockerCredential{
		&DockerCredential{
			Username: "luke-skywalker",
			Password: "tatooine",
		},
		&DockerCredential{
			Username: "_json_key",
			Password: "Order 66",
			Host:     "https://gcr.io",
		},
	}
	require.Equal(s.T(), expectedCredentials, project.credentials)
	expectedResourceGraph := &ResourceGraph{
		ResourceGroups: map[string]*ResourceGroup{
			"configs": &ResourceGroup{
				Name:   "configs",
				Depend: []string{},
				Wait:   []string{},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "configs", "postgres-configs.yml"),
					},
				},
				Children: []string{"databases"},
			},
			"secrets": &ResourceGroup{
				Name:   "secrets",
				Depend: []string{},
				Wait:   []string{},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "secrets", "postgres-secrets.yml"),
					},
				},
				Children: []string{"databases"},
			},
			"databases": &ResourceGroup{
				Name:   "databases",
				Depend: []string{"configs", "secrets"},
				Wait:   []string{},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "databases", "postgres.yml"),
					},
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "databases", "redis.yml"),
					},
				},
				Children: []string{"init-jobs"},
			},
			"init-jobs": &ResourceGroup{
				Name:   "init-jobs",
				Depend: []string{"databases"},
				Wait:   []string{},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "jobs", "init-postgres.yml"),
					},
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "jobs", "init-redis.yml"),
					},
				},
				Children: []string{"services"},
			},
			"services": &ResourceGroup{
				Name:   "services",
				Depend: []string{"init-jobs"},
				Wait:   []string{"init-postgres", "init-redis"},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "services", "app.yml"),
					},
				},
				Children: []string{},
			},
		},
		RootNodes: []string{"configs", "secrets"},
	}
	require.Equal(s.T(), expectedResourceGraph, s.stripResourceContent(project.resourceGraph))

	s.verifyVariableValue(
		project.resourceGraph.ResourceGroups["databases"].ResourceFiles[0].RawContent,
		"image: ",
		"image: postgres:9.6",
	)
	s.verifyVariableValue(
		project.resourceGraph.ResourceGroups["databases"].ResourceFiles[1].RawContent,
		"image: ",
		"image: redis:4-alpine",
	)
	s.verifyVariableValue(
		project.resourceGraph.ResourceGroups["init-jobs"].ResourceFiles[0].RawContent,
		"image: ",
		"image: postgres-sidecar:1.1.4",
	)
	s.verifyVariableValue(
		project.resourceGraph.ResourceGroups["init-jobs"].ResourceFiles[1].RawContent,
		"image: ",
		"image: redis-sidecar:1.1.4",
	)
	s.verifyVariableValue(
		project.resourceGraph.ResourceGroups["services"].ResourceFiles[0].RawContent,
		"image: ",
		"image: app:1.1.4",
	)
}

func (s *ProjectTestSuite) stripResourceContent(resourceGraph *ResourceGraph) *ResourceGraph {
	// Deep copy to a new resource by encode - decode json
	b, err := json.Marshal(resourceGraph)
	require.Nil(s.T(), err)
	strippedGraph := &ResourceGraph{}
	err = json.Unmarshal(b, strippedGraph)
	require.Nil(s.T(), err)
	for _, rg := range strippedGraph.ResourceGroups {
		for _, rf := range rg.ResourceFiles {
			rf.RawContent = nil
			for _, r := range rf.Resources {
				r.RawContent = nil
			}
		}
	}
	return strippedGraph
}

func (s *ProjectTestSuite) verifyVariableValue(rawContent []byte, search, expectedValue string) {
	scanner := bufio.NewScanner(bytes.NewReader(rawContent))
	found := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, search) {
			require.Equal(s.T(), expectedValue, line)
			found = true
			break
		}
	}
	if !found {
		require.FailNow(s.T(), "cannot find pattern %q", search)
	}
}

func TestProject(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}

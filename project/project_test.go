package project

import (
	"bufio"
	"bytes"
	"encoding/json"
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

func (s *ProjectTestSuite) SetupSuite() {
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
	expectedResourceGraph := &ResourceGraph{
		ResourceGroups: map[string]*ResourceGroup{
			"configs": &ResourceGroup{
				Name:   "configs",
				Depend: []string{},
				Wait:   []*WaitConfig{},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "configs", "postgres-configs.yml"),
						Resources: []*Resource{
							&Resource{
								Name:     "postgres",
								Kind:     "ConfigMap",
								Filepath: filepath.Join(projectDir, "configs", "postgres-configs.yml"),
							},
						},
					},
				},
				Children: []string{"databases"},
			},
			"secrets": &ResourceGroup{
				Name:   "secrets",
				Depend: []string{},
				Wait:   []*WaitConfig{},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "secrets", "postgres-secrets.yml"),
						Resources: []*Resource{
							&Resource{
								Name:     "postgres",
								Kind:     "Secret",
								Filepath: filepath.Join(projectDir, "secrets", "postgres-secrets.yml"),
							},
						},
					},
				},
				Children: []string{"databases"},
			},
			"databases": &ResourceGroup{
				Name:   "databases",
				Depend: []string{"configs", "secrets"},
				Wait:   []*WaitConfig{},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "databases", "postgres.yml"),
						Resources: []*Resource{
							&Resource{
								Name:     "postgres",
								Kind:     "Deployment",
								Filepath: filepath.Join(projectDir, "databases", "postgres.yml"),
							},
							&Resource{
								Name:     "postgres",
								Kind:     "Service",
								Filepath: filepath.Join(projectDir, "databases", "postgres.yml"),
							},
						},
					},
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "databases", "redis.yml"),
						Resources: []*Resource{
							&Resource{
								Name:     "redis",
								Kind:     "Deployment",
								Filepath: filepath.Join(projectDir, "databases", "redis.yml"),
							},
							&Resource{
								Name:     "redis",
								Kind:     "Service",
								Filepath: filepath.Join(projectDir, "databases", "redis.yml"),
							},
						},
					},
				},
				Children: []string{"init-jobs"},
			},
			"init-jobs": &ResourceGroup{
				Name:   "init-jobs",
				Depend: []string{"databases"},
				Wait:   []*WaitConfig{},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "jobs", "init-postgres.yml"),
						Resources: []*Resource{
							&Resource{
								Name:     "init-postgres",
								Kind:     "Job",
								Filepath: filepath.Join(projectDir, "jobs", "init-postgres.yml"),
							},
						},
					},
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "jobs", "init-redis.yml"),
						Resources: []*Resource{
							&Resource{
								Name:     "init-redis",
								Kind:     "Job",
								Filepath: filepath.Join(projectDir, "jobs", "init-redis.yml"),
							},
						},
					},
				},
				Children: []string{"services"},
			},
			"services": &ResourceGroup{
				Name:   "services",
				Depend: []string{"init-jobs"},
				Wait: []*WaitConfig{
					&WaitConfig{
						Name: "init-postgres",
						Kind: "job",
					},
					&WaitConfig{
						Name: "init-redis",
						Kind: "job",
					},
				},
				ResourceFiles: []*ResourceFile{
					&ResourceFile{
						FilePath: filepath.Join(projectDir, "services", "app.yml"),
						Resources: []*Resource{
							&Resource{
								Name:     "app",
								Kind:     "Deployment",
								Filepath: filepath.Join(projectDir, "services", "app.yml"),
							},
							&Resource{
								Name:     "app",
								Kind:     "Service",
								Filepath: filepath.Join(projectDir, "services", "app.yml"),
							},
						},
					},
				},
				Children: []string{},
			},
		},
		RootNodes: []string{"configs", "secrets"},
		LeafNodes: []string{"services"},
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

	expectedConfigContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres
data:
  pgdata: /data/postgres
`
	require.Equal(s.T(), expectedConfigContent, project.resourceGraph.ResourceGroups["configs"].ResourceFiles[0].RawContent)
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
			rf.RawContent = ""
			for _, r := range rf.Resources {
				r.RawContent = ""
			}
		}
	}
	return strippedGraph
}

func (s *ProjectTestSuite) verifyVariableValue(rawContent string, search, expectedValue string) {
	scanner := bufio.NewScanner(bytes.NewReader([]byte(rawContent)))
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

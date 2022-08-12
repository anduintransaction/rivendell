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
	variableFiles := []string{
		filepath.Join(projectDir, "vars", "vars"),
	}
	project, err := ReadProject(projectFile, namespace, context, kubeConfig, variables, variableFiles, nil, nil)
	require.Nil(s.T(), err, "should read project file successfully")
	require.Equal(s.T(), projectDir, project.rootDir)
	require.Equal(s.T(), namespace, project.namespace)
	require.Equal(s.T(), context, project.context)
	require.Equal(s.T(), kubeConfig, project.kubeConfig)
	expectedVariables := map[string]string{
		"key1":                   "value1",
		"key2":                   "value2",
		"postgresTag":            "9.6",
		"postgresImageTag":       "9.6",
		"redisTag":               "4-alpine",
		"appTag":                 "1.1.4",
		"postgresSidecarImage":   "postgres-sidecar:1.1.4",
		"redisSidecarImage":      "redis-sidecar:1.1.4",
		"rivendellVarNamespace":  namespace,
		"rivendellVarContext":    context,
		"rivendellVarKubeConfig": kubeConfig,
		"rivendellVarRootDir":    project.rootDir,
	}
	require.Equal(s.T(), expectedVariables, project.variables)
	expectedResourceGraph := &ResourceGraph{
		ResourceGroups: map[string]*ResourceGroup{
			"configs": {
				Name:   "configs",
				Depend: []string{},
				Wait:   []*WaitConfig{},
				ResourceFiles: []*ResourceFile{
					{
						Source: filepath.Join(projectDir, "configs", "postgres-configs.yml"),
						Resources: []*Resource{
							{
								Name:     "postgres",
								Kind:     "ConfigMap",
								Filepath: filepath.Join(projectDir, "configs", "postgres-configs.yml"),
							},
						},
					},
				},
				Children: []string{"databases"},
			},
			"secrets": {
				Name:   "secrets",
				Depend: []string{},
				Wait:   []*WaitConfig{},
				ResourceFiles: []*ResourceFile{
					{
						Source: filepath.Join(projectDir, "secrets", "postgres-secrets.yml"),
						Resources: []*Resource{
							{
								Name:     "postgres",
								Kind:     "Secret",
								Filepath: filepath.Join(projectDir, "secrets", "postgres-secrets.yml"),
							},
						},
					},
				},
				Children: []string{"databases"},
			},
			"databases": {
				Name:   "databases",
				Depend: []string{"configs", "secrets"},
				Wait:   []*WaitConfig{},
				ResourceFiles: []*ResourceFile{
					{
						Source: filepath.Join(projectDir, "databases", "postgres.yml"),
						Resources: []*Resource{
							{
								Name:     "postgres",
								Kind:     "Deployment",
								Filepath: filepath.Join(projectDir, "databases", "postgres.yml"),
							},
							{
								Name:     "postgres",
								Kind:     "Service",
								Filepath: filepath.Join(projectDir, "databases", "postgres.yml"),
							},
						},
					},
					{
						Source: filepath.Join(projectDir, "databases", "redis.yml"),
						Resources: []*Resource{
							{
								Name:     "redis",
								Kind:     "Deployment",
								Filepath: filepath.Join(projectDir, "databases", "redis.yml"),
							},
							{
								Name:     "redis",
								Kind:     "Service",
								Filepath: filepath.Join(projectDir, "databases", "redis.yml"),
							},
						},
					},
				},
				Children: []string{"init-jobs"},
			},
			"init-jobs": {
				Name:   "init-jobs",
				Depend: []string{"databases"},
				Wait:   []*WaitConfig{},
				ResourceFiles: []*ResourceFile{
					{
						Source: filepath.Join(projectDir, "jobs", "init-postgres.yml"),
						Resources: []*Resource{
							{
								Name:     "init-postgres",
								Kind:     "Job",
								Filepath: filepath.Join(projectDir, "jobs", "init-postgres.yml"),
							},
						},
					},
					{
						Source: filepath.Join(projectDir, "jobs", "init-redis.yml"),
						Resources: []*Resource{
							{
								Name:     "init-redis",
								Kind:     "Job",
								Filepath: filepath.Join(projectDir, "jobs", "init-redis.yml"),
							},
						},
					},
				},
				Children: []string{"services"},
			},
			"services": {
				Name:   "services",
				Depend: []string{"init-jobs"},
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
				ResourceFiles: []*ResourceFile{
					{
						Source: filepath.Join(projectDir, "services", "app.yml"),
						Resources: []*Resource{
							{
								Name:     "app",
								Kind:     "Deployment",
								Filepath: filepath.Join(projectDir, "services", "app.yml"),
							},
							{
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

func (s *ProjectTestSuite) TestGlob() {
	projectDir := filepath.Join(s.resourceRoot, "config-test", "glob")
	projectFile := filepath.Join(projectDir, "project.yml")
	includes := []string{"**/*.yml"}
	excludes := []string{"**/mango.yml"}
	variableFiles := []string{}
	project, err := ReadProject(projectFile, "dota", "", "", nil, variableFiles, includes, excludes)
	require.Nil(s.T(), err)
	actualFiles := []string{}
	project.resourceGraph.WalkForward(func(g *ResourceGroup) error {
		for _, f := range g.ResourceFiles {
			actualFiles = append(actualFiles, strings.TrimPrefix(f.Source, projectDir))
		}
		return nil
	})
	expected := []string{
		"/items/consumables/bottle.yml",
		"/items/consumables/tango.yml",
		"/items/defence/shiva.yml",
	}
	require.Equal(s.T(), expected, actualFiles)
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

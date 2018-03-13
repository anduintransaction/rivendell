package project

import (
	"bufio"
	"bytes"
	"sort"
	"strings"

	"github.com/anduintransaction/rivendell/kubernetes"
	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
	yaml "gopkg.in/yaml.v2"
)

// ResourceGraph .
type ResourceGraph struct {
	ResourceGroups map[string]*ResourceGroup
	RootNodes      []string
}

// ResourceGroup holds configuration for a resource group
type ResourceGroup struct {
	Name          string
	ResourceFiles []*ResourceFile
	Depend        []string
	Wait          []string
	Children      []string
}

// ResourceFile holds configuration for a single resource file.
// There are maybe multiple resources in a resource file
type ResourceFile struct {
	FilePath   string
	Resources  []*Resource
	RawContent string
}

// Resource holds configuration for a single resource
type Resource struct {
	Filepath   string
	Name       string
	Kind       string
	RawContent string
}

type resourceYAML struct {
	Kind     string                `yaml:"kind"`
	Metadata *resourceMetadataYAML `yaml:"metadata"`
}

type resourceMetadataYAML struct {
	Name string `yaml:"name"`
}

// ReadResourceGraph .
func ReadResourceGraph(rootDir string, resourceGroupConfigs []*ResourceGroupConfig, variables map[string]string) (*ResourceGraph, error) {
	resourceGraph := &ResourceGraph{
		ResourceGroups: make(map[string]*ResourceGroup),
		RootNodes:      []string{},
	}
	for _, resourceGroupConfig := range resourceGroupConfigs {
		resourceGroup := &ResourceGroup{
			Name:     resourceGroupConfig.Name,
			Depend:   utils.NilArrayToEmpty(resourceGroupConfig.Depend),
			Wait:     utils.NilArrayToEmpty(resourceGroupConfig.Wait),
			Children: []string{},
		}
		resourceGraph.ResourceGroups[resourceGroup.Name] = resourceGroup
		if len(resourceGroup.Depend) == 0 {
			resourceGraph.RootNodes = append(resourceGraph.RootNodes, resourceGroup.Name)
		}
		includePatterns := utils.PrependPaths(rootDir, resourceGroupConfig.Resources)
		excludePatterns := utils.PrependPaths(rootDir, resourceGroupConfig.Excludes)
		resourceFiles, err := utils.GlobFiles(includePatterns, excludePatterns)
		if err != nil {
			return nil, err
		}
		for _, resourceFile := range resourceFiles {
			fileContent, err := utils.ExecuteTemplate(resourceFile, variables)
			if err != nil {
				return nil, err
			}
			rf := &ResourceFile{
				FilePath:   resourceFile,
				RawContent: resourceGraph.removeNamespace(string(fileContent)),
			}
			err = resourceGraph.splitResourceFile(rf)
			if err != nil {
				return nil, err
			}
			resourceGroup.ResourceFiles = append(resourceGroup.ResourceFiles, rf)
		}
	}
	sort.Strings(resourceGraph.RootNodes)
	err := resourceGraph.resolveChildren()
	if err != nil {
		return nil, err
	}
	err = resourceGraph.cyclicCheck()
	if err != nil {
		return nil, err
	}
	return resourceGraph, nil
}

// Walk through the graph, BFS style
func (rg *ResourceGraph) Walk(f func(g *ResourceGroup) error) error {
	candidates := []string{}
	for _, candidate := range rg.RootNodes {
		candidates = append(candidates, candidate)
	}
	visited := utils.NewStringSet()
	for len(candidates) > 0 {
		current := candidates[0]
		if !visited.Exists(current) {
			depVisited := true
			for _, dep := range rg.ResourceGroups[current].Depend {
				if !visited.Exists(dep) {
					depVisited = false
					break
				}
			}
			if depVisited {
				err := f(rg.ResourceGroups[current])
				if err != nil {
					return err
				}
				visited.Add(current)
			}
		}
		candidates = append(candidates[1:], rg.ResourceGroups[current].Children...)
	}
	return nil
}

// WalkResource .
func (rg *ResourceGraph) WalkResource(f func(r *Resource, g *ResourceGroup) error) error {
	return rg.Walk(func(g *ResourceGroup) error {
		for _, rf := range g.ResourceFiles {
			for _, r := range rf.Resources {
				err := f(r, g)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// Exists .
func (r *Resource) Exists(kubeContext *kubernetes.Context) error {
	return nil
}

// Create .
func (r *Resource) Create(kubeContext *kubernetes.Context) error {
	return nil
}

func (rg *ResourceGraph) removeNamespace(content string) string {
	scanner := bufio.NewScanner(bytes.NewBuffer([]byte(content)))
	strippedContent := ""
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(strings.TrimSpace(line), "namespace:") {
			strippedContent += line + "\n"
		}
	}
	return strippedContent
}

func (rg *ResourceGraph) splitResourceFile(resourceFile *ResourceFile) error {
	parts := strings.Split(string(resourceFile.RawContent), "---\n")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}
		parsedResource := &resourceYAML{}
		err := yaml.Unmarshal([]byte(part), parsedResource)
		if err != nil {
			return err
		}
		resource := &Resource{
			Name:       parsedResource.Metadata.Name,
			Kind:       parsedResource.Kind,
			Filepath:   resourceFile.FilePath,
			RawContent: part,
		}
		resourceFile.Resources = append(resourceFile.Resources, resource)
	}
	return nil
}

func (rg *ResourceGraph) resolveChildren() error {
	for _, resourceGroup := range rg.ResourceGroups {
		sort.Strings(resourceGroup.Depend)
		for _, depend := range resourceGroup.Depend {
			parent, ok := rg.ResourceGroups[depend]
			if !ok {
				return stacktrace.Propagate(ErrMissingDependency{resourceGroup.Name, depend}, "missing dependency")
			}
			parent.Children = append(parent.Children, resourceGroup.Name)
		}
	}
	for _, resourceGroup := range rg.ResourceGroups {
		sort.Strings(resourceGroup.Children)
	}
	return nil
}

func (rg *ResourceGraph) cyclicCheck() error {
	if len(rg.ResourceGroups) == 0 {
		return nil
	}
	white := make(utils.StringSet)
	gray := make(utils.StringSet)
	black := make(utils.StringSet)
	for name := range rg.ResourceGroups {
		white.Add(name)
	}
	for len(white) > 0 {
		current := white.First()
		err := rg.cyclicDFS(current, white, gray, black)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rg *ResourceGraph) cyclicDFS(current string, white, gray, black utils.StringSet) error {
	white.Remove(current)
	gray.Add(current)
	for _, neighbor := range rg.ResourceGroups[current].Children {
		if black.Exists(neighbor) {
			continue
		}
		if gray.Exists(neighbor) {
			return stacktrace.Propagate(ErrCyclicDependency{neighbor}, "cyclic dependency found for %q", neighbor)
		}
		err := rg.cyclicDFS(neighbor, white, gray, black)
		if err != nil {
			return err
		}
	}
	gray.Remove(current)
	black.Add(current)
	return nil
}

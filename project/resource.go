package project

import (
	"sort"
	"time"

	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
)

// ResourceGraph .
type ResourceGraph struct {
	ResourceGroups map[string]*ResourceGroup
	RootNodes      []string
	LeafNodes      []string
}

// ResourceGroup holds configuration for a resource group
type ResourceGroup struct {
	Name          string
	Templater     string
	ResourceFiles []*ResourceFile
	Depend        []string
	Wait          []*WaitConfig
	Children      []string
}

// ResourceFile holds configuration for a single resource file.
// There are maybe multiple resources in a resource file
type ResourceFile struct {
	Source          string
	ContextDir      string
	Resources       []*Resource
	RawContent      string
	ExpandedContent string
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

const (
	defaultWaitTimeout = 300
)

// ReadResourceGraph .
func ReadResourceGraph(rootDir string, resourceGroupConfigs []*ResourceGroupConfig, variables map[string]string, includeResources []string, excludeResources []string) (*ResourceGraph, error) {
	rg := &ResourceGraph{
		ResourceGroups: make(map[string]*ResourceGroup),
		RootNodes:      []string{},
		LeafNodes:      []string{},
	}

	for _, resourceGroupConfig := range resourceGroupConfigs {
		g := &ResourceGroup{
			Name:     resourceGroupConfig.Name,
			Depend:   utils.NilArrayToEmpty(resourceGroupConfig.Depend),
			Children: []string{},
		}
		g.Wait = resourceGroupConfig.Wait
		if g.Wait == nil {
			g.Wait = []*WaitConfig{}
		}
		rg.ResourceGroups[g.Name] = g
		if len(g.Depend) == 0 {
			rg.RootNodes = append(rg.RootNodes, g.Name)
		}

		// resolve and process resource files
		resourceFiles, err := resolveResourceFile(rootDir, resourceGroupConfig, includeResources, excludeResources)
		if err != nil {
			return nil, err
		}
		processors := []ResourceFileProcessor{
			expandResourceContent(variables),
			stripNamespace(),
			splitResourceContent(),
		}
		for _, resourceFile := range resourceFiles {
			for _, proc := range processors {
				if err := proc.Process(resourceFile); err != nil {
					return nil, err
				}
			}
		}
		g.ResourceFiles = resourceFiles
	}

	sort.Strings(rg.RootNodes)
	err := rg.resolveChildren()
	if err != nil {
		return nil, err
	}
	err = rg.cyclicCheck()
	if err != nil {
		return nil, err
	}
	return rg, nil
}

// WalkForwardWithWait from root nodes
func (rg *ResourceGraph) WalkForwardWithWait(f func(g *ResourceGroup) error, readyFunc func(r *Resource, g *ResourceGroup) error, waitFunc func(name, kind string) error) error {
	readyResourceGroups := make(map[*ResourceGroup]bool)
	readyResources := make(map[*Resource]bool)
	return rg.WalkForward(func(g *ResourceGroup) error {
		for _, depGroupName := range g.Depend {
			depGroup := rg.ResourceGroups[depGroupName]
			if !readyResourceGroups[depGroup] {
				for _, r := range depGroup.allResources() {
					if !readyResources[r] {
						if readyFunc != nil {
							err := readyFunc(r, depGroup)
							if err != nil {
								return err
							}
						}
						readyResources[r] = true
					}
				}
			}
			readyResourceGroups[depGroup] = true
		}
		for _, wait := range g.Wait {
			err := rg.waitFor(wait, waitFunc)
			if err != nil {
				return err
			}
		}
		if f == nil {
			return nil
		}
		return f(g)
	})
}

// WalkBackwardWithWait from leaf nodes
func (rg *ResourceGraph) WalkBackwardWithWait(f func(g *ResourceGroup) error, readyFunc func(r *Resource, g *ResourceGroup) error) error {
	readyResourceGroups := make(map[*ResourceGroup]bool)
	readyResources := make(map[*Resource]bool)
	return rg.WalkBackward(func(g *ResourceGroup) error {
		for _, depGroupName := range g.Children {
			depGroup := rg.ResourceGroups[depGroupName]
			if !readyResourceGroups[depGroup] {
				for _, r := range depGroup.allResources() {
					if !readyResources[r] {
						if readyFunc != nil {
							err := readyFunc(r, depGroup)
							if err != nil {
								return err
							}
						}
						readyResources[r] = true
					}
				}
			}
			readyResourceGroups[depGroup] = true
		}
		if f == nil {
			return nil
		}
		return f(g)
	})
}

// WalkForward through the graph, BFS style
func (rg *ResourceGraph) WalkForward(f func(g *ResourceGroup) error) error {
	candidates := append([]string{}, rg.RootNodes...)
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
				if f != nil {
					err := f(rg.ResourceGroups[current])
					if err != nil {
						return err
					}
				}
				visited.Add(current)
			}
		}
		candidates = append(candidates[1:], rg.ResourceGroups[current].Children...)
	}
	return nil
}

// WalkBackward through the graph, BFS style
func (rg *ResourceGraph) WalkBackward(f func(g *ResourceGroup) error) error {
	candidates := append([]string{}, rg.LeafNodes...)
	visited := utils.NewStringSet()
	for len(candidates) > 0 {
		current := candidates[0]
		if !visited.Exists(current) {
			depVisited := true
			for _, dep := range rg.ResourceGroups[current].Children {
				if !visited.Exists(dep) {
					depVisited = false
					break
				}
			}
			if depVisited {
				if f != nil {
					err := f(rg.ResourceGroups[current])
					if err != nil {
						return err
					}
				}
				visited.Add(current)
			}
		}
		candidates = append(candidates[1:], rg.ResourceGroups[current].Depend...)
	}
	return nil
}

// WalkResourceForward with waiting
func (rg *ResourceGraph) WalkResourceForward(f func(r *Resource, g *ResourceGroup) error, readyFunc func(r *Resource, g *ResourceGroup) error, waitFunc func(name, kind string) error) error {
	return rg.WalkForwardWithWait(func(g *ResourceGroup) error {
		for _, rf := range g.ResourceFiles {
			for _, r := range rf.Resources {
				if f == nil {
					return nil
				}
				err := f(r, g)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}, readyFunc, waitFunc)
}

// WalkResourceBackward with waiting
func (rg *ResourceGraph) WalkResourceBackward(f func(r *Resource, g *ResourceGroup) error, readyFunc func(r *Resource, g *ResourceGroup) error) error {
	return rg.WalkBackwardWithWait(func(g *ResourceGroup) error {
		for _, rf := range g.ResourceFiles {
			for _, r := range rf.Resources {
				if f == nil {
					return nil
				}
				err := f(r, g)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}, readyFunc)
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
		if len(resourceGroup.Children) == 0 {
			rg.LeafNodes = append(rg.LeafNodes, resourceGroup.Name)
		}
	}
	sort.Strings(rg.LeafNodes)
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

func (rg *ResourceGraph) waitFor(wait *WaitConfig, waitFunc func(name, kind string) error) error {
	if waitFunc == nil {
		return nil
	}
	waitChan := make(chan error, 1)
	go func() {
		err := waitFunc(wait.Name, wait.Kind)
		waitChan <- err
	}()
	timeout := wait.Timeout
	if timeout <= 0 {
		timeout = defaultWaitTimeout
	}
	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	select {
	case err := <-waitChan:
		return err
	case <-timer.C:
		return stacktrace.Propagate(ErrWaitTimeout{wait.Name, wait.Kind}, "wait timeout")
	}
}

func (g *ResourceGroup) allResources() []*Resource {
	resources := []*Resource{}
	for _, rf := range g.ResourceFiles {
		resources = append(resources, rf.Resources...)
	}
	return resources
}

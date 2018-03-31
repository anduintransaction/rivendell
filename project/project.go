package project

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/anduintransaction/rivendell/kubernetes"
	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
)

const (
	waitDelay = 3 * time.Second
	waitCount = 10
)

// Project holds configuration for a rivendell task
type Project struct {
	rootDir       string
	namespace     string
	context       string
	kubeConfig    string
	variables     map[string]string
	resourceGraph *ResourceGraph
}

// ReadProject reads a project from file
func ReadProject(projectFile, namespace, context, kubeConfig string, variables map[string]string) (*Project, error) {
	projectConfig, err := ReadProjectConfig(projectFile, variables)
	if err != nil {
		return nil, err
	}
	project := &Project{
		context:    context,
		kubeConfig: kubeConfig,
	}
	project.resolveProjectRoot(projectFile, projectConfig.RootDir)
	project.resolveNamespace(namespace, projectConfig.Namespace)
	project.resolveVariables(variables, projectConfig.Variables)
	if err != nil {
		return nil, err
	}
	err = project.resolveResourceGraph(projectConfig.ResourceGroups)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// Debug .
func (p *Project) Debug() {
	p.printCommonInfo()
	p.resourceGraph.WalkForward(func(g *ResourceGroup) error {
		utils.Info("Resource group %q", g.Name)
		for _, rf := range g.ResourceFiles {
			utils.Info2("Resource file %q", rf.FilePath)
			fmt.Println()
			for _, r := range rf.Resources {
				fmt.Println(r.RawContent)
				fmt.Println()
			}
		}
		fmt.Println()
		return nil
	})
}

// Up .
func (p *Project) Up() error {
	kubeContext, err := kubernetes.NewContext(p.namespace, p.context, p.kubeConfig)
	if err != nil {
		return err
	}
	p.printCommonInfo()
	err = p.createNamespace(kubeContext)
	if err != nil {
		return err
	}
	return p.resourceGraph.WalkResourceForward(func(r *Resource, g *ResourceGroup) error {
		return p.createResource(kubeContext, g, r)
	}, func(r *Resource, g *ResourceGroup) error {
		return p.waitForExists(kubeContext, r)
	})
}

// Down .
func (p *Project) Down() error {
	kubeContext, err := kubernetes.NewContext(p.namespace, p.context, p.kubeConfig)
	if err != nil {
		return err
	}
	p.printCommonInfo()
	err = p.resourceGraph.WalkResourceBackward(func(r *Resource, g *ResourceGroup) error {
		return p.deleteResource(kubeContext, g, r)
	}, func(r *Resource, g *ResourceGroup) error {
		return p.waitForDeleted(kubeContext, r)
	})
	// Delete namespace anyway, this may have the side-effect of deleting all resources.
	errNamespaceDelete := p.deleteNamespace(kubeContext)
	if errNamespaceDelete != nil {
		utils.Error(err)
	}
	return err
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

func (p *Project) resolveResourceGraph(resourceGroupConfigs []*ResourceGroupConfig) error {
	resourceGraph, err := ReadResourceGraph(p.rootDir, resourceGroupConfigs, p.variables)
	if err != nil {
		return err
	}
	p.resourceGraph = resourceGraph
	return nil
}

func (p *Project) printCommonInfo() {
	utils.Info("Using namespace %q", p.namespace)
	utils.Info("Using context %q", p.context)
	utils.Info("Using kubernetes config file %q", p.kubeConfig)
}

func (p *Project) createNamespace(kubeContext *kubernetes.Context) error {
	utils.Info("Creating namespace %q", p.namespace)
	exists, err := kubeContext.Namespace().Create()
	if err != nil {
		return err
	}
	p.printCreateResult(exists)
	return nil
}

func (p *Project) deleteNamespace(kubeContext *kubernetes.Context) error {
	utils.Warn("Deleting namespace %q", p.namespace)
	exists, err := kubeContext.Namespace().Delete()
	if err != nil {
		return err
	}
	p.printDeleteResult(exists)
	return nil
}

func (p *Project) createResource(kubeContext *kubernetes.Context, g *ResourceGroup, r *Resource) error {
	utils.Info("Creating %s %q in group %q", r.Kind, r.Name, g.Name)
	exists, err := kubeContext.Resource().Create(r.Name, r.Kind, r.RawContent)
	if err != nil {
		return err
	}
	p.printCreateResult(exists)
	return nil
}

func (p *Project) deleteResource(kubeContext *kubernetes.Context, g *ResourceGroup, r *Resource) error {
	utils.Warn("Deleting %s %q in group %q", r.Kind, r.Name, g.Name)
	exists, err := kubeContext.Resource().Delete(r.Name, r.Kind)
	if err != nil {
		return err
	}
	p.printDeleteResult(exists)
	return nil
}

func (p *Project) waitForExists(kubeContext *kubernetes.Context, r *Resource) error {
	count := 0
	for {
		exists, err := kubeContext.Resource().Exists(r.Name, r.Kind)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		count++
		if count > waitCount {
			return stacktrace.Propagate(ErrWaitTimeout{r.Name, r.Kind}, "wait timeout")
		}
		time.Sleep(waitDelay)
	}
}

func (p *Project) waitForDeleted(kubeContext *kubernetes.Context, r *Resource) error {
	count := 0
	for {
		exists, err := kubeContext.Resource().Exists(r.Name, r.Kind)
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}
		count++
		if count > waitCount {
			return stacktrace.Propagate(ErrWaitTimeout{r.Name, r.Kind}, "wait timeout")
		}
		time.Sleep(waitDelay)
	}
}

func (p *Project) printCreateResult(exists bool) {
	if exists {
		utils.Warn("====> Existed")
	} else {
		utils.Success("====> Success")
	}
}

func (p *Project) printDeleteResult(exists bool) {
	if exists {
		utils.Success("====> Success")
	} else {
		utils.Warn("====> Not exist")
	}
}

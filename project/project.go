package project

import (
	"fmt"
	"path/filepath"

	"github.com/anduintransaction/rivendell/kubernetes"
	"github.com/anduintransaction/rivendell/utils"
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
	p.resourceGraph.Walk(func(g *ResourceGroup) error {
		utils.Info("Resource group %q\n", g.Name)
		for _, rf := range g.ResourceFiles {
			utils.Info2("Resource file %q\n", rf.FilePath)
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
	err = kubeContext.Namespace().Create()
	if err != nil {
		return err
	}
	return p.resourceGraph.WalkResource(func(r *Resource, g *ResourceGroup) error {
		utils.Info("Creating %s %q in group %q\n", r.Kind, r.Name, g.Name)
		return r.Create(kubeContext)
	})
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
	return nil
}

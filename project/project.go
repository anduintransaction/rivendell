package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anduintransaction/rivendell/kubernetes"
	"github.com/anduintransaction/rivendell/utils"
	"github.com/joho/godotenv"
	"github.com/palantir/stacktrace"
)

const (
	waitDelay = 3 * time.Second
	waitCount = 10
)

// FilterFunc criteria if a resource group should be process by Walk function
type FilterFunc func(*ResourceGroup) bool

type Formatter interface {
	Format(p *Project)
}

// Project holds configuration for a rivendell task
type Project struct {
	rootDir               string
	namespace             string
	context               string
	kubeConfig            string
	variables             map[string]string
	resourceGraph         *ResourceGraph
	filterFn              FilterFunc
	deleteNamespaceConfig bool
	config                *Config
}

// ReadProject reads a project from file
func ReadProject(projectFile, namespace, context, kubeConfig string, variables map[string]string, variableFiles []string, includeResources []string, excludeResources []string) (*Project, error) {
	project := &Project{
		context:    context,
		kubeConfig: kubeConfig,
	}
	err := project.resolveCommandlineVariables(variables, variableFiles)
	if err != nil {
		return nil, err
	}
	projectConfig, err := ReadProjectConfig(projectFile, project.variables)
	if err != nil {
		return nil, err
	}
	project.config = projectConfig
	project.deleteNamespaceConfig = projectConfig.DeleteNamespace
	project.resolveProjectRoot(projectFile, projectConfig.RootDir)
	project.resolveNamespace(namespace, projectConfig.Namespace)
	project.resolveVariables(projectConfig.Variables)
	err = project.resolveResourceGraph(projectConfig.ResourceGroups, includeResources, excludeResources)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// Debug .
func (p *Project) Debug(f Formatter) {
	f.Format(p)
}

// Up .
func (p *Project) Up() error {
	kubeContext, err := kubernetes.NewContext(p.namespace, p.context, p.kubeConfig)
	if err != nil {
		return err
	}
	err = p.createNamespace(kubeContext)
	if err != nil {
		return err
	}
	return p.resourceGraph.WalkResourceForward(func(r *Resource, g *ResourceGroup) error {
		return p.createResource(kubeContext, g, r)
	}, func(r *Resource, g *ResourceGroup) error {
		return p.waitForExists(kubeContext, r)
	}, func(name, kind string) error {
		utils.Info2("Waiting for %s %q", kind, name)
		return p.waitForResource(kubeContext, name, kind)
	})
}

// Down .
func (p *Project) Down(deleteNS, deletePVC bool) error {
	kubeContext, err := kubernetes.NewContext(p.namespace, p.context, p.kubeConfig)
	if err != nil {
		return err
	}
	err = p.resourceGraph.WalkResourceBackward(func(r *Resource, g *ResourceGroup) error {
		kind := strings.ToLower(r.Kind)
		isPVC := kind == "persistentvolumeclaim" || kind == "pvc"
		if !deletePVC && isPVC {
			return nil
		}
		return p.deleteResource(kubeContext, g, r)
	}, func(r *Resource, g *ResourceGroup) error {
		return p.waitForDeleted(kubeContext, r)
	})
	if !deleteNS {
		return nil
	}

	// Delete namespace anyway, this may have the side-effect of deleting all resources.
	errNamespaceDelete := p.deleteNamespace(kubeContext)
	if errNamespaceDelete != nil {
		utils.Error(errNamespaceDelete)
	}
	return err
}

// Update .
func (p *Project) Update() error {
	kubeContext, err := kubernetes.NewContext(p.namespace, p.context, p.kubeConfig)
	if err != nil {
		return err
	}
	return p.resourceGraph.WalkResourceForward(func(r *Resource, g *ResourceGroup) error {
		return p.updateResource(kubeContext, g, r)
	}, nil, func(name, kind string) error {
		utils.Info2("Waiting for %s %q", kind, name)
		return p.waitForResource(kubeContext, name, kind)
	})
}

// Upgrade .
func (p *Project) Upgrade() error {
	kubeContext, err := kubernetes.NewContext(p.namespace, p.context, p.kubeConfig)
	if err != nil {
		return err
	}
	return p.resourceGraph.WalkResourceForward(func(r *Resource, g *ResourceGroup) error {
		return p.upgradeResource(kubeContext, g, r)
	}, nil, func(name, kind string) error {
		utils.Info2("Waiting for %s %q", kind, name)
		return p.waitForResource(kubeContext, name, kind)
	})
}

// GetServicePods
func (p *Project) GetServicePods() ([]string, error) {
	kubeContext, err := kubernetes.NewContext(p.namespace, p.context, p.kubeConfig)
	if err != nil {
		return nil, err
	}
	pods := utils.NewStringSet()
	_ = p.resourceGraph.WalkResourceForward(func(r *Resource, g *ResourceGroup) error {
		if strings.ToLower(r.Kind) == "service" {
			servicePods, err := kubeContext.Service().ListPods(r.Name)
			if err != nil {
				return err
			}
			pods.Add(servicePods...)
		}
		return nil
	}, nil, nil)
	return pods.ToSlice(), nil
}

// Restart .
func (p *Project) Restart(pods []string) error {
	kubeContext, err := kubernetes.NewContext(p.namespace, p.context, p.kubeConfig)
	if err != nil {
		return err
	}
	for _, pod := range pods {
		utils.Warn("Deleting pod %q", pod)
		exists, err := kubeContext.Resource().Delete(pod, "pod")
		if err != nil {
			return err
		}
		p.printDeleteResult(exists)
	}
	return nil
}

// PrintCommonInfo .
func (p *Project) PrintCommonInfo() {
	out := os.Stderr
	utils.Infof(out, "Using namespace %q", p.namespace)
	utils.Infof(out, "Using context %q", p.context)
	utils.Infof(out, "Using kubernetes config file %q", p.kubeConfig)
}

func (p *Project) PrintConfig() {
	p.config.Write(os.Stdout)
}

// PrintUpPlan .
func (p *Project) PrintUpPlan() {
	utils.Info("The following resources will be created:")
	p.resourceGraph.WalkResourceForward(func(r *Resource, g *ResourceGroup) error {
		fmt.Printf(" - %s %q\n", r.Kind, r.Name)
		return nil
	}, nil, nil)
}

// PrintDownPlan .
func (p *Project) PrintDownPlan() {
	utils.Warn("The following resources will be destroyed:")
	p.resourceGraph.WalkResourceBackward(func(r *Resource, g *ResourceGroup) error {
		fmt.Printf(" - %s %q\n", r.Kind, r.Name)
		return nil
	}, nil)
}

// PrintUpdatePlan .
func (p *Project) PrintUpdatePlan() {
	utils.Warn("The following resources will be updated: ")
	p.resourceGraph.WalkResourceForward(func(r *Resource, g *ResourceGroup) error {
		fmt.Printf(" - %s %q\n", r.Kind, r.Name)
		return nil
	}, nil, func(name, kind string) error {
		fmt.Printf("- [wait] %s/%s\n", kind, name)
		return nil
	})
}

// PrintRestartPlan .
func (p *Project) PrintRestartPlan(pods []string) {
	utils.Warn("The following pods will be restarted: ")
	for _, pod := range pods {
		fmt.Printf(" - %s\n", pod)
	}
}

func (p *Project) WalkForward(fn func(g *ResourceGroup) error) error {
	return p.resourceGraph.WalkForward(func(g *ResourceGroup) error {
		if p.filterFn != nil && !p.filterFn(g) {
			return nil
		}
		return fn(g)
	})
}

func (p *Project) SetFilter(fn func(*ResourceGroup) bool) *Project {
	p.filterFn = fn
	return p
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

func (p *Project) resolveCommandlineVariables(variablesFromCommand map[string]string, variableFiles []string) error {
	variablesFromFiles := make(map[string]string)
	var err error
	if len(variableFiles) > 0 {
		variablesFromFiles, err = godotenv.Read(variableFiles...)
		if err != nil {
			return stacktrace.Propagate(err, "cannot read variable files")
		}
	}
	p.variables = utils.MergeMaps(variablesFromFiles, variablesFromCommand)
	return nil
}

func (p *Project) resolveVariables(variablesFromConfig map[string]string) {
	rivendellVariables := map[string]string{
		"rivendellVarNamespace":  p.namespace,
		"rivendellVarContext":    p.context,
		"rivendellVarKubeConfig": p.kubeConfig,
		"rivendellVarRootDir":    p.rootDir,
	}
	p.variables = utils.MergeMaps(variablesFromConfig, p.variables, rivendellVariables)
}

func (p *Project) resolveResourceGraph(resourceGroupConfigs []*ResourceGroupConfig, includeResources []string, excludeResources []string) error {
	resourceGraph, err := ReadResourceGraph(p.rootDir, resourceGroupConfigs, p.variables, includeResources, excludeResources)
	if err != nil {
		return err
	}
	p.resourceGraph = resourceGraph
	return nil
}

func (p *Project) createNamespace(kubeContext *kubernetes.Context) error {
	if p.namespace == "" {
		return nil
	}
	utils.Info("Creating namespace %q", p.namespace)
	exists, err := kubeContext.Namespace().Create()
	if err != nil {
		return err
	}
	p.printCreateResult(exists)
	return nil
}

func (p *Project) deleteNamespace(kubeContext *kubernetes.Context) error {
	if p.namespace == "" || p.namespace == "default" || !p.deleteNamespaceConfig {
		return nil
	}
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

func (p *Project) updateResource(kubeContext *kubernetes.Context, g *ResourceGroup, r *Resource) error {
	utils.Warn("Updating %s %q in group %q", r.Kind, r.Name, g.Name)
	updateStatus, err := kubeContext.Resource().Update(r.Name, r.Kind, r.RawContent)
	if err != nil {
		return err
	}
	p.printUpdateResult(updateStatus)
	return nil
}

func (p *Project) upgradeResource(kubeContext *kubernetes.Context, g *ResourceGroup, r *Resource) error {
	utils.Warn("Upgrading %s %q in group %q", r.Kind, r.Name, g.Name)
	updateStatus, err := kubeContext.Resource().Upgrade(r.Name, r.Kind, r.RawContent)
	if err != nil {
		return err
	}
	p.printUpdateResult(updateStatus)
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

func (p *Project) waitForResource(kubeContext *kubernetes.Context, name, kind string) error {
	success, err := kubeContext.Resource().Wait(name, kind)
	if err != nil {
		return err
	}
	if success {
		return nil
	}
	return stacktrace.Propagate(ErrWaitFailed{name, kind}, "wait failed")
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

func (p *Project) printUpdateResult(updateStatus kubernetes.UpdateStatus) {
	switch updateStatus {
	case kubernetes.UpdateStatusNotExist:
		utils.Warn("====> Not exist")
	case kubernetes.UpdateStatusExisted:
		utils.Success("====> Success")
	case kubernetes.UpdateStatusSkipped:
		utils.Info2("====> Skipped")
	}
}

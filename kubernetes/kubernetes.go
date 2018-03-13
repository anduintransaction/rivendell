package kubernetes

import (
	"github.com/anduintransaction/rivendell/utils"
	"github.com/palantir/stacktrace"
)

// Context .
type Context struct {
	namespace  string
	context    string
	kubeConfig string
}

// NewContext .
func NewContext(namespace, context, kubeConfig string) (*Context, error) {
	c := &Context{namespace, context, kubeConfig}
	err := c.checkDeps()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Namespace .
func (c *Context) Namespace() *Namespace {
	return &Namespace{c}
}

func (c *Context) checkDeps() error {
	status, err := utils.ExecuteCommandSilently("which", "kubectl")
	if err != nil {
		return nil
	}
	if status.ExitCode != 0 {
		return stacktrace.Propagate(ErrMissingCommand{"kubectl"}, "missing command %q", "kubectl")
	}
	return nil
}

func (c *Context) completeArgsWithoutNamespace(args []string) []string {
	if c.context != "" {
		args = append(args, "--context", c.context)
	}
	if c.kubeConfig != "" {
		args = append(args, "--kubeconfig", c.kubeConfig)
	}
	return args
}

func (c *Context) completeArgs(args []string) []string {
	if c.namespace != "" {
		args = append(args, "-n", c.namespace)
	}
	return c.completeArgsWithoutNamespace(args)
}

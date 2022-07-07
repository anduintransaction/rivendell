package formatters

import (
	"fmt"
	"os"

	"github.com/anduintransaction/rivendell/project"
)

type TreeFormatter struct{}

func NewTreeFormatter() *TreeFormatter {
	return &TreeFormatter{}
}

func (f *TreeFormatter) Format(p *project.Project) {
	out := os.Stdout

	p.PrintCommonInfo()
	p.WalkForward(func(g *project.ResourceGroup) error {
		fmt.Fprintf(out, "- Group: %s\n", g.Name)
		for _, rf := range g.ResourceFiles {
			fmt.Fprintf(out, "  - File: %s\n", rf.FilePath)
		}
		for _, rd := range g.Depend {
			fmt.Fprintf(out, "  - Dep: %s\n", rd)
		}
		for _, rw := range g.Wait {
			fmt.Fprintf(out, "  - Wait: %s/%s\n", rw.Kind, rw.Name)
		}
		return nil
	})
}

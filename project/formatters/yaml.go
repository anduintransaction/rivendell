package formatters

import (
	"fmt"
	"os"

	"github.com/anduintransaction/rivendell/project"
)

type YamlFormatter struct{}

func NewYamlFormatter() *YamlFormatter {
	return &YamlFormatter{}
}

func (f *YamlFormatter) Format(p *project.Project) {
	out := os.Stdout
	sep := "---"

	p.WalkForward(func(g *project.ResourceGroup) error {
		for _, rf := range g.ResourceFiles {
			for _, r := range rf.Resources {
				fmt.Fprintf(out, "# Resource group %q - Resource file %q\n", g.Name, rf.FilePath)
				fmt.Fprintln(out)
				fmt.Fprintln(out, r.RawContent)
				fmt.Fprintln(out, sep)
			}
		}
		return nil
	})
}

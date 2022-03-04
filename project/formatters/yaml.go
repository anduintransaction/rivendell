package formatters

import (
	"fmt"
	"os"

	"github.com/anduintransaction/rivendell/project"
	"github.com/anduintransaction/rivendell/utils"
)

type YamlFormatter struct{}

func NewYamlFormatter() *YamlFormatter {
	return &YamlFormatter{}
}

func (f *YamlFormatter) Format(p *project.Project, filterGroups []string) {
	m := utils.StringSliceToMap(filterGroups)
	out := os.Stdout
	sep := "---"

	p.WalkForward(func(g *project.ResourceGroup) error {
		if len(m) > 0 && !m[g.Name] {
			return nil
		}

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

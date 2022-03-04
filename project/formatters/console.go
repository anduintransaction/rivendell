package formatters

import (
	"fmt"

	"github.com/anduintransaction/rivendell/project"
	"github.com/anduintransaction/rivendell/utils"
)

type ConsoleFormatter struct{}

func NewConsoleFormatter() *ConsoleFormatter {
	return &ConsoleFormatter{}
}

func (f *ConsoleFormatter) Format(p *project.Project, filterGroups []string) {
	m := utils.StringSliceToMap(filterGroups)
	p.PrintCommonInfo()
	p.WalkForward(func(g *project.ResourceGroup) error {
		if len(m) > 0 && !m[g.Name] {
			return nil
		}

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

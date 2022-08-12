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

func (f *ConsoleFormatter) Format(p *project.Project) {
	p.PrintCommonInfo()
	p.WalkForward(func(g *project.ResourceGroup) error {
		utils.Info("Resource group %q", g.Name)
		for _, rf := range g.ResourceFiles {
			utils.Info2("Resource file %q", rf.Source)
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

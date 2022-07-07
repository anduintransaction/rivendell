package formatters

import (
	"github.com/anduintransaction/rivendell/project"
)

type ConfigFormatter struct{}

func NewConfigFormatter() *ConfigFormatter {
	return &ConfigFormatter{}
}

func (f *ConfigFormatter) Format(p *project.Project) {
	p.PrintCommonInfo()
	p.PrintConfig()
}

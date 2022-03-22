package filters

import (
	"regexp"

	"github.com/anduintransaction/rivendell/project"
)

func FilterByRegex(p *regexp.Regexp) project.FilterFunc {
	return func(rg *project.ResourceGroup) bool {
		if p == nil {
			return true
		}
		return p.Match([]byte(rg.Name))
	}
}

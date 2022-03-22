package filters

import (
	"github.com/anduintransaction/rivendell/project"
	"github.com/anduintransaction/rivendell/utils"
)

func FilterBySlice(s []string) project.FilterFunc {
	m := utils.StringSliceToMap(s)
	return func(rg *project.ResourceGroup) bool {
		if len(m) == 0 || m[rg.Name] {
			return true
		}
		return false
	}
}

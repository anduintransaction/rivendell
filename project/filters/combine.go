package filters

import (
	"github.com/anduintransaction/rivendell/project"
)

func CombineFilter(fns ...project.FilterFunc) project.FilterFunc {
	return func(rg *project.ResourceGroup) bool {
		if len(fns) == 0 {
			return true
		}

		for _, fn := range fns {
			if fn(rg) {
				return true
			}
		}
		return false
	}
}

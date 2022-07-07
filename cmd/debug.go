// Copyright © 2018 Anduin Transactions Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"regexp"
	"strings"

	"github.com/anduintransaction/rivendell/project"
	"github.com/anduintransaction/rivendell/project/filters"
	"github.com/anduintransaction/rivendell/project/formatters"
	"github.com/anduintransaction/rivendell/utils"
	"github.com/spf13/cobra"
)

var (
	outputFormat      string
	filterGroups      []string
	filterGroupsRegex string
)

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug [project file]",
	Short: "Print all resources description",
	Long:  "Print all resources description",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.ReadProject(args[0], namespace, context, kubeConfig, variableMap, variableFiles, includeResources, excludeResources)
		if err != nil {
			utils.Fatal(err)
		}

		var filterFns []project.FilterFunc

		if len(filterGroups) > 0 {
			filterFns = append(filterFns, filters.FilterBySlice(filterGroups))
		}

		var filterPattern *regexp.Regexp
		if fgr := strings.TrimSpace(filterGroupsRegex); fgr != "" {
			filterPattern, err = regexp.Compile(fgr)
			if err != nil {
				utils.Warn("Invalid filter group regex. Ignore pattern")
			}
			if filterPattern != nil {
				filterFns = append(filterFns, filters.FilterByRegex(filterPattern))
			}
		}

		p.SetFilter(filters.CombineFilter(filterFns...))

		var formatter project.Formatter
		switch strings.ToLower(outputFormat) {
		case "console":
			formatter = formatters.NewConsoleFormatter()
		case "yaml":
			formatter = formatters.NewYamlFormatter()
		case "tree":
			formatter = formatters.NewTreeFormatter()
		case "config":
			formatter = formatters.NewConfigFormatter()
		default:
			utils.Warn("Unknown output formatter. Fallback to console")
			formatter = formatters.NewConsoleFormatter()
		}

		p.Debug(formatter)
	},
}

func init() {
	RootCmd.AddCommand(debugCmd)

	debugCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "print format. One of: console|yaml|tree|config")
	debugCmd.Flags().StringSliceVar(&filterGroups, "filter-groups", []string{}, "Only print resource groups")
	debugCmd.Flags().StringVar(&filterGroupsRegex, "filter-groups-regex", "", "Only print resource groups matching pattern")
}

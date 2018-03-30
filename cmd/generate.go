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
	"os"

	"github.com/anduintransaction/rivendell/project"
	"github.com/anduintransaction/rivendell/utils"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [filename]",
	Short: "Generate a project file",
	Long:  "Generate a project file",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		w, err := os.Create(filename)
		if err != nil {
			utils.Fatal(err)
		}
		defer w.Close()
		config := &project.Config{
			RootDir:   ".",
			Namespace: "your_namespace",
			Variables: map[string]string{
				"key":        "value",
				"envKey":     "${ENV_VAR}",
				"dynamicVar": "{{.dynamicVar}}",
			},
			ResourceGroups: []*project.ResourceGroupConfig{
				&project.ResourceGroupConfig{
					Name:      "group1",
					Resources: []string{"path/to/group1/*.yml"},
					Excludes:  []string{"path/to/group1/exclude*"},
					Depend:    []string{},
					Wait:      []string{},
				},
				&project.ResourceGroupConfig{
					Name:      "group2",
					Resources: []string{"path/to/group2/*.yml"},
					Excludes:  []string{"**/exclude*"},
					Depend:    []string{"group1"},
					Wait:      []string{"job1"},
				},
			},
		}
		err = config.Write(w)
		if err != nil {
			utils.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(generateCmd)
}

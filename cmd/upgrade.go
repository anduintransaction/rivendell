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

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade all declared resources, including pod and job",
	Long: `Upgrade all declated resources, including pod and job

Resources other than pod and job are updated normally (like using update command).
Pod and job must be stopped before upgraded, and will be restarted.
Running pod or job are untouched.
Unlike update command, non-existed resources will be created.
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.ReadProject(args[0], namespace, context, kubeConfig, variableMap, includeResources, excludeResources)
		if err != nil {
			utils.Fatal(err)
		}
		p.PrintCommonInfo()
		p.PrintUpdatePlan()
		if !yes {
			utils.Ask("Upgrade all resource?", "yes", "no")
			ok, err := utils.ExpectAnswer("yes")
			if err != nil {
				utils.Fatal(err)
			}
			if !ok {
				os.Exit(0)
			}
		}
		err = p.Upgrade()
		if err != nil {
			utils.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(upgradeCmd)
}

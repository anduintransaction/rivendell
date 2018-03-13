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
	"github.com/anduintransaction/rivendell/project"
	"github.com/anduintransaction/rivendell/utils"
	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up [project file]",
	Short: "Create all resources defined in a rivendell project file",
	Long:  "Create all resources defined in a rivendell project file",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.ReadProject(args[0], namespace, context, kubeConfig, variableMap)
		if err != nil {
			utils.Fatal(err)
		}
		err = p.Up()
		if err != nil {
			utils.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(upCmd)
}

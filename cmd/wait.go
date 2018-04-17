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

var waitTimeout int

// waitCmd represents the wait command
var waitCmd = &cobra.Command{
	Use:   "wait [pod|job] [name]",
	Short: "Wait for a pod or a job to complete",
	Long:  "Wait for a pod or a job to complete",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := project.Wait(namespace, context, kubeConfig, args[0], args[1], waitTimeout)
		if err != nil {
			utils.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(waitCmd)

	waitCmd.Flags().IntVarP(&waitTimeout, "timeout", "t", 300, "timeout in second (value <= 0 for infinity)")
}

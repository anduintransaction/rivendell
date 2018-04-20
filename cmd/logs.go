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

var logContainer string
var logTimeout int

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [name]",
	Short: "Print the log of a container in a pod",
	Long: `Print the log of a container in a pod

If the pod has multiple containers and no container is specified from the command,
the first container will be chosen.
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := project.Logs(namespace, context, kubeConfig, args[0], logContainer, logTimeout)
		if err != nil {
			utils.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)

	logsCmd.Flags().StringVar(&logContainer, "container", "", "print log from this container")
	logsCmd.Flags().IntVarP(&logTimeout, "timeout", "t", 0, "timeout in second, use value <= 0 for infinity")
}

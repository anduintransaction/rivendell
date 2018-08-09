// Copyright © 2018 Anduin Transaction Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"os"

	"github.com/anduintransaction/rivendell/project"
	"github.com/anduintransaction/rivendell/utils"
	"github.com/spf13/cobra"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart [project file]",
	Short: "Restart all pods associated with a service in the project",
	Long:  "Restart all pods associated with a service in the project",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.ReadProject(args[0], namespace, context, kubeConfig, variableMap, variableFiles, includeResources, excludeResources)
		if err != nil {
			utils.Fatal(err)
		}
		pods, err := p.GetServicePods()
		if err != nil {
			utils.Fatal(err)
		}
		p.PrintCommonInfo()
		p.PrintRestartPlan(pods)
		if !yes {
			utils.Ask("Restart all pods?", "yes", "no")
			ok, err := utils.ExpectAnswer("yes")
			if err != nil {
				utils.Fatal(err)
			}
			if !ok {
				os.Exit(0)
			}
		}
		err = p.Restart(pods)
		if err != nil {
			utils.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(restartCmd)
}

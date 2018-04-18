// Copyright © 2018 Anduin Transactions Inc
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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var namespace string
var context string
var kubeConfig string
var variableArray = []string{}
var variableMap = map[string]string{}
var includeResources []string
var excludeResources []string
var yes = false

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "rivendell",
	Short: "Kubernetes resources and services management tool",
	Long:  `Kubernetes resources and services management tool`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		for _, variable := range variableArray {
			segments := strings.SplitN(variable, "=", 2)
			if len(segments) != 2 {
				fmt.Fprintf(os.Stderr, "Invalid variable: %s\n", variable)
				os.Exit(2)
			}
			variableMap[segments[0]] = segments[1]
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "set kubernetes namespace")
	RootCmd.PersistentFlags().StringVarP(&context, "context", "c", "", "set kubernetes context")
	RootCmd.PersistentFlags().StringVar(&kubeConfig, "kubeconfig", "", "set kubectl config file")
	RootCmd.PersistentFlags().StringArrayVar(&variableArray, "variable", []string{}, "variables to pass to rivendell task file, for example: --variable key1=value1 --variable key2=value2")
	RootCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "Run command immediately")
	RootCmd.PersistentFlags().StringArrayVar(&includeResources, "include", []string{}, "include file patterns, for example --include=**/service.yml --include=**/deployment.yml")
	RootCmd.PersistentFlags().StringArrayVar(&includeResources, "exclude", []string{}, "exclude file patterns, for example --exclude=**/config.yml --exclude=**/secret.yml")
}

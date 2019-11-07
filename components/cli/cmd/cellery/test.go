/*
 * Copyright (c) 2018 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http:www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package main

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/commands"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
)

func newTestCommand(cli cli.Cli) *cobra.Command {
	var name string
	var debug bool
	var verbose bool
	var incell bool
	var startDependencies bool
	var shareAllInstances bool
	var assumeYes bool
	var dependencyLinks []string
	var envVars []string
	cmd := &cobra.Command{
		Use:   "test [<registry>/]<organization>/<cell-image>:<version>",
		Short: "Test a cell image",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.MinimumNArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			err = image.ValidateImageTagWithRegistry(args[0])
			if err != nil {
				return err
			}
			if name != "" {
				isCellValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), name)
				if err != nil || !isCellValid {
					return fmt.Errorf("expects a valid cell name, received %s", args[0])
				}
			}
			for _, dependencyLink := range dependencyLinks {
				isMatch, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.DEPENDENCY_LINK_PATTERN),
					dependencyLink)
				if err != nil || !isMatch {
					return fmt.Errorf("expects dependency links in the format "+
						"[<parent-instance>.]<alias>:<dependency-instance>, received %s", dependencyLink)
				}
			}
			for _, envVar := range envVars {
				isMatch, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CLI_ARG_ENV_VAR_PATTERN),
					envVar)
				if err != nil || !isMatch {
					return fmt.Errorf("expects environment varibles in the format "+
						"[<instance>:]<key>=<value>, received %s", envVar)
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunTest(cli, args[0], name, startDependencies, shareAllInstances, dependencyLinks, envVars,
				assumeYes, debug, verbose, incell)
		},
		Example: "  cellery test cellery-samples/hr:1.0.0 -n hr-inst\n" +
			"  cellery test cellery-samples/hr:1.0.0 -n hr-inst -y\n" +
			"  cellery test cellery-samples/hr:1.0.0 -n hr-inst --assume-yes\n" +
			"  cellery test registry.foo.io/cellery-samples/hr:1.0.0 -n hr-inst -l employee:employee-inst" +
			" -l stock:stock-inst \n" +
			" -v\n" +
			" --debug\n" +
			" --incell",
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the cell instance")
	cmd.Flags().BoolVarP(&assumeYes, "assume-yes", "y", false,
		"Flag to enable/disable prompting for confirmation before starting instance(s)")
	cmd.Flags().BoolVarP(&startDependencies, "start-dependencies", "d", false,
		"Start all the dependencies of this Cell Image in order")
	cmd.Flags().BoolVarP(&shareAllInstances, "share-instances", "s", false,
		"Share all instances among equivalent Cell Instances")
	cmd.Flags().StringArrayVarP(&dependencyLinks, "link", "l", []string{},
		"Link an instance with a dependency alias")
	cmd.Flags().StringArrayVarP(&envVars, "env", "e", []string{},
		"Set an environment variable for the cellery test method in the Cell file")

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Run on verbose mode")
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable test debug mode")
	cmd.Flags().BoolVar(&incell, "incell", false, "Enable in-cell testing")
	return cmd
}

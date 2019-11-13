/*
 * Copyright (c) 2018 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/cellery-io/sdk/components/cli/pkg/commands/instance"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func newLogsCommand(cli cli.Cli) *cobra.Command {
	var component string
	var syslog bool
	cmd := &cobra.Command{
		Use:   "logs <instance-name>",
		Short: "Displays logs for either the cell instance, or a component of a running cell instance.",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			isCellValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), args[0])
			if err != nil || !isCellValid {
				return fmt.Errorf("expects a valid cell name, received %s", args[0])
			}
			if component != "" {
				isComponentValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), component)
				if err != nil || !isComponentValid {
					return fmt.Errorf("expects a valid component name, received %s", args[0])
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := instance.RunLogs(cli, args[0], component, syslog); err != nil {
				util.ExitWithErrorMessage("Cellery logs command failed", err)
			}
		},
		Example: "  cellery logs employee\n" +
			"  cellery logs employee -c salary",
	}
	cmd.Flags().StringVarP(&component, "component", "c", "", "component of the cell")
	cmd.Flags().BoolVarP(&syslog, "syslog", "s", false, "view system logs")
	return cmd
}

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
	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/commands"
)

func newTerminateCommand(cli cli.Cli) *cobra.Command {
	var terminateAll = false
	cmd := &cobra.Command{
		Use:     "terminate <instance1> <instance2> <instance-3>",
		Short:   "Terminate running cell instances",
		Aliases: []string{"term"},
		Args: func(cmd *cobra.Command, args []string) error {
			if !terminateAll {
				err := cobra.MinimumNArgs(1)(cmd, args)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunTerminate(cli, args, terminateAll)
		},
		Example: "  cellery terminate employee\n" +
			"  cellery terminate pet-fe pet-be\n" +
			"  cellery terminate --all",
	}
	cmd.Flags().BoolVar(&terminateAll, "all", false, "Delete all cell instances")
	return cmd
}

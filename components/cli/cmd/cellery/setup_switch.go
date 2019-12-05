/*
 * Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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
	"github.com/spf13/cobra"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/commands/setup"
	"cellery.io/cellery/components/cli/pkg/util"
)

func newSetupSwitchCommand(cli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch <command>",
		Short: "Switch into a k8s cluster",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			return setup.ValidateCluster(cli, args[0])
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := setup.RunSetupSwitch(cli, args[0]); err != nil {
				util.ExitWithErrorMessage("Cellery setup switch command failed", err)
			}
		},
		Example: "  cellery setup switch cellery-admin@cellery",
	}
	return cmd
}

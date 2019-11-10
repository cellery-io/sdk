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
	"github.com/cellery-io/sdk/components/cli/pkg/commands/instance"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func newListInstancesCommand(cli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "instances",
		Short:   "List all running cells",
		Aliases: []string{"instance", "inst"},
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := instance.RunListInstances(cli); err != nil {
				util.ExitWithErrorMessage("Cellery list instances command failed", err)
			}
		},
		Example: "  cellery list instances",
	}
	return cmd
}

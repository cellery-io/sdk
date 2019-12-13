/*
 * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/commands/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
)

func newSetNamespaceCommand(cli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespace <namespace>",
		Short:   "Set the namespace targeted by the CLI",
		Aliases: []string{"namespace", "ns"},
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := runtime.RunSetNamespace(cli, args[0]); err != nil {
				util.ExitWithErrorMessage("Cellery set namespace command failed", err)
			}
		},
		Example: "  cellery set namespace test",
	}
	return cmd
}

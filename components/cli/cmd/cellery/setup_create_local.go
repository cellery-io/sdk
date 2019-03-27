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

	"github.com/cellery-io/sdk/components/cli/pkg/commands"
)

func newSetupCreateLocalCommand() *cobra.Command {
	var isCompleteSetup = false
	cmd := &cobra.Command{
		Use:   "local",
		Short: "Create a local Cellery runtime",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunSetupCreateLocal(isCompleteSetup)
		},
		Example: "  cellery setup create local",
	}
	cmd.Flags().BoolVarP(&isCompleteSetup, "complete", "c", false, "Create complete local setup")
	return cmd
}

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
	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/commands"
)

func newInitCommand() *cobra.Command {
	var projectName = ""
	var testStr = ""
	cmd := &cobra.Command{
		Use:   "init [PROJECT_NAME]",
		Short: "Initialize a cell project",
		Args:  cobra.MaximumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				testStr = args[0]
				projectName = args[1]
				commands.RunInit(projectName, testStr)
			} else if len(args) > 0 {
				testStr = ""
				projectName = args[0]
				commands.RunInit(projectName, testStr)
			}

		},
		Example: "  cellery init [PROJECT_NAME] or cellery init test [PROJECT_PATH]",
	}
	return cmd
}

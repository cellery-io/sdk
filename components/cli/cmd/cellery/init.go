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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/commands/project"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func newInitCommand(cli cli.Cli) *cobra.Command {
	var projectName = ""
	var testStr = ""
	cmd := &cobra.Command{
		Use:   "init [PROJECT_NAME]",
		Short: "Initialize a cell project",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.MaximumNArgs(2)(cmd, args)
			if err != nil {
				return err
			}

			if len(args) > 1 {
				testStr = args[0]
				if testStr != "test" {
					return fmt.Errorf("invalid argument. expects %s, recieved %s", util.Bold("test"), testStr)
				}
				projectName = args[1]
				if !(strings.HasSuffix(projectName, ".bal")) {
					return fmt.Errorf("expects a valid bal file, recieved %v", projectName)
				}
				isExist, err := util.FileExists(filepath.Join(cli.FileSystem().CurrentDir(), projectName))
				if err != nil {
					return err
				}
				if !isExist {
					return fmt.Errorf("expects a valid path, recieved %v", projectName)
				}
			} else if len(args) > 0 {
				testStr = ""
				projectName = args[0]
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := project.RunInit(cli, projectName, testStr); err != nil {
				util.ExitWithErrorMessage("Cellery init command failed", err)
			}

		},
		Example: "  cellery init [PROJECT_NAME]\n" +
			"  cellery init test [PROJECT_PATH]/[PROJECT_NAME].bal",
	}
	return cmd
}

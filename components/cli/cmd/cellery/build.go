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
	"os"
	"path/filepath"

	"cellery.io/cellery/components/cli/pkg/constants"

	"github.com/spf13/cobra"

	"cellery.io/cellery/components/cli/cli"
	image2 "cellery.io/cellery/components/cli/pkg/commands/image"
	"cellery.io/cellery/components/cli/pkg/image"
	"cellery.io/cellery/components/cli/pkg/util"
)

// newBuildCommand creates a cobra command which can be invoked to build a cell image from a cell file
func newBuildCommand(cli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build <cell-file-or-project>",
		Short: "Build an immutable cell image with the required dependencies",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(2)(cmd, args)
			if err != nil {
				return err
			}
			fileinfo, err := os.Stat(args[0])
			if err != nil {
				return err
			}
			if fileinfo.IsDir() {
				isProperProject, err := util.FileExists(filepath.Join(args[0], constants.BallerinaToml))
				if err != nil || !isProperProject {
					return fmt.Errorf("expects a proper ballerina project, received %s", args[0])
				}
			} else {
				isProperFile, err := util.FileExists(args[0])
				if err != nil || !isProperFile {
					return fmt.Errorf("expects a proper file, received %s", args[0])
				}
			}
			err = image.ValidateImageTag(args[1])
			if err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := image2.RunBuild(cli, args[1], args[0]); err != nil {
				util.ExitWithErrorMessage("Cellery build command failed", err)
			}
		},
		Example: "  cellery build employee.bal cellery-samples/employee:1.0.0\n" +
			"  cellery build employee/ cellery-samples/employee:1.0.0",
	}
	return cmd
}

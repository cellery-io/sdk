/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/commands"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// newExtractResourcesCommand creates a command which can be invoked to extract the cell
// image resources to a specific directory.
func newExtractResourcesCommand() *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "extract-resources <organization>/<cell-image>:<version> <output-directory>",
		Short: "Extract the resource files of a pulled image to the provided location",
		Aliases: []string{"exctr"},
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			err = util.ValidateImageTag(args[0])
			if err != nil {
				return fmt.Errorf("expects <organization>/<cell-image>:<version> as cell-image, received %s", args[0])
			}
			if outputPath != "" {
				err = util.CreateDir(args[1])
				if err != nil {
					return fmt.Errorf("expects valid directory as the output-directory, received %s", args[1])
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunExtractResources(args[0], outputPath)
		},
		Example: "  cellery extract-resources cellery-samples/employee:1.0.0 ./resources",
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "",
		"Cell image in the format: <organization>/<cell-image>:<version>")
	return cmd
}

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
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/commands"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func newUpdateCellComponentsCommand() *cobra.Command {
	var containerImage string
	var envVars []string
	cmd := &cobra.Command{
		Use: "update <cell instance> <new cell image> \n" +
			"update <instance name> <component name> --container-image mycellorg/hellocell:1.0.0 \n" +
			"update <instance name> <component name> --container-image mycellorg/hellocell:1.0.0 --env foo=bar --env bob=alice",
		Short: "update components of a cell instance using a newer cell image",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(2)(cmd, args)
			if err != nil {
				return err
			}
			isCellInstValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), args[0])
			if err != nil {
				return err
			}
			if !isCellInstValid {
				return fmt.Errorf("expects a valid cell instance name, received %s", args[0])
			}
			// if the container image is empty we consider the 2nd argument as a image, hence validate
			if containerImage == "" {
				isCellImageValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELL_IMAGE_PATTERN), args[1])
				if err != nil {
					return err
				}
				if !isCellImageValid {
					return fmt.Errorf("expects a valid cell image name, received %s", args[1])
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// if container image is provided, this is an inline update command targeting a particular component of the
			// specified cell instance.
			var err error
			if containerImage != "" {
				err = commands.RunUpdateForSingleComponent(args[0], args[1], containerImage, envVars)
			} else {
				err = commands.RunUpdateComponents(args[0], args[1])
			}
			if err != nil {
				util.ExitWithErrorMessage(fmt.Sprintf("Unable to apply autoscale policies to instance %s", args[0]), err)
			}
		},
		Example: "  cellery update hello1 cellery/sample-hello:1.0.3",
	}
	cmd.Flags().StringVarP(&containerImage, "container-image", "i", "", "container image name")
	cmd.Flags().StringArrayVarP(&envVars, "env", "e", []string{}, "environment variables")
	return cmd
}

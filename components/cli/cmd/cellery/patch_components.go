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

func newPatchComponentsCommand() *cobra.Command {
	var containerImage string
	var containerName string
	var envVars []string
	cmd := &cobra.Command{
		Use: "patch <instance name> <component name> --container-image mycontainerorg/hello:1.0.0 \n" +
			"  	  patch <instance name> <component name> --container-image mycontainerorg/hello:1.0.0 --env foo=bar --env bob=alice",
		Short: "patch a particular component of a cell/composite instance with a new container image",
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
				return fmt.Errorf("expects a valid cell/composite instance name, received %s", args[0])
			}
			// if the container image is empty we consider the 2nd argument as a image, hence validate
			if containerImage == "" {
				return fmt.Errorf("expects a valid container image, received none")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := commands.RunPatchForSingleComponent(args[0], args[1], containerImage, containerName, envVars)
			if err != nil {
				util.ExitWithErrorMessage(fmt.Sprintf("Unable to patch cell component %s in instance %s", args[1], args[0]), err)
			}
		},
		Example: "  cellery patch myhello hellocomponent --container-image mycontainerorg/hello:1.0.1 --env foo=bar",
	}
	cmd.Flags().StringVarP(&containerImage, "container-image", "i", "", "container image")
	cmd.Flags().StringVarP(&containerImage, "container-name", "n", "", "container name")
	cmd.Flags().StringArrayVarP(&envVars, "env", "e", []string{}, "environment variables")
	return cmd
}

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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/commands"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func newPushCommand(cli cli.Cli) *cobra.Command {
	var username string
	var password string
	cmd := &cobra.Command{
		Use:   "push [<registry>/]<organization>/<cell-image>:<version>",
		Short: "Push cell image to the remote repository",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			err = image.ValidateImageTagWithRegistry(args[0])
			if err != nil {
				return err
			}
			if password != "" && username == "" {
				return fmt.Errorf("expects username if the password is provided, username not provided")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := commands.RunPush(cli, args[0], username, password); err != nil {
				util.ExitWithErrorMessage("Cellery push command failed", err)
			}
		},
		Example: "  cellery push cellery-samples/employee:1.0.0\n" +
			"  cellery push registry.foo.io/cellery-samples/employee:1.0.0",
	}
	cmd.Flags().StringVarP(&username, "username", "u", "", "Username for Cellery Registry")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for Cellery Registry")
	return cmd
}

/*
 * Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/commands/hub"
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/util"
)

// newLoginCommand saves the credentials for a particular registry
func newLoginCommand(cli cli.Cli) *cobra.Command {
	var username string
	var password string
	cmd := &cobra.Command{
		Use:   "login [registry-url]",
		Short: "Login to a Cellery Registry",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.MaximumNArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			if password != "" && username == "" {
				return fmt.Errorf("expects username if the password/ token is provided, username not provided")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				if err := hub.RunLogin(cli, args[0], username, password); err != nil {
					util.ExitWithErrorMessage("Cellery login command failed", err)
				}
			} else {
				if err := hub.RunLogin(cli, constants.CentralRegistryHost, username, password); err != nil {
					util.ExitWithErrorMessage("Cellery login command failed", err)
				}
			}
		},
		Example: "  cellery login\n" +
			"  cellery login -u john -p john123" +
			"  cellery login registry.foo.io",
	}
	cmd.Flags().StringVarP(&username, "username", "u", "", "Username for Cellery Registry")
	cmd.Flags().StringVarP(&password, "password", "p", "",
		"Password/ Token for Cellery Registry")
	return cmd
}

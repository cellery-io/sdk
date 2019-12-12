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
	"strings"

	"github.com/spf13/cobra"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/commands/setup"
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/gcp"
	"cellery.io/cellery/components/cli/pkg/util"
)

func newSetupCleanupGcpCommand(cli cli.Cli) *cobra.Command {
	var confirmed = false
	var uniqueNumber string
	cmd := &cobra.Command{
		Use:   "gcp",
		Short: "Cleanup gcp setup",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			valid, err := setup.ValidateGcpCluster(args[0])
			if !valid || err != nil {
				return fmt.Errorf("Gcp cluster " + args[0] + " doesn't exist")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(strings.Split(args[0], constants.GcpClusterName)) > 1 {
				uniqueNumber = strings.Split(args[0], constants.GcpClusterName)[1]
			}
			platform, err := gcp.NewGcp(gcp.SetUuid(uniqueNumber))
			if err != nil {
				util.ExitWithErrorMessage("Cellery setup cleanup gcp command failed", fmt.Errorf(
					"failed to initialize celleryGcp platform, %v", err))
			}
			if err := setup.RunSetupCleanupPlatform(cli, platform, confirmed); err != nil {
				util.ExitWithErrorMessage("Cellery setup cleanup gcp command failed", err)
			}
		},
		Example: "  cellery setup cleanup gcp <cluster_name>",
	}
	cmd.Flags().BoolVarP(&confirmed, "assume-yes", "y", false, "Confirm setup removal")
	return cmd
}

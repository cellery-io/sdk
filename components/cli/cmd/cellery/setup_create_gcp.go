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
	"cellery.io/cellery/components/cli/pkg/commands/setup"
	"cellery.io/cellery/components/cli/pkg/gcp"
	"cellery.io/cellery/components/cli/pkg/util"
)

func newSetupCreateGcpCommand(cli cli.Cli, isComplete *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcp",
		Short: "Create a Cellery runtime in gcp",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			platform, err := gcp.NewGcp()
			if err != nil {
				util.ExitWithErrorMessage("Cellery setup create gcp command failed",
					fmt.Errorf("failed to initialize gcp platform, %v", err))
			}
			_, mysql, nfs, err := setup.RunSetupCreateCelleryPlatform(cli, platform)
			if err != nil {
				cleanupErr := setup.RunSetupCleanupPlatform(cli, platform, true)
				if cleanupErr != nil {
					util.ExitWithErrorMessage("Cellery setup create gcp command failed",
						fmt.Errorf("failed to create gcp platform, %v. Failed to remove partially created gcp "+
							"platform, %v", err, cleanupErr))
				} else {
					util.ExitWithErrorMessage("Cellery setup create gcp command failed",
						fmt.Errorf("failed to create gcp platform, %v", err))
				}
			}
			if err := setup.RunSetupCreateCelleryRuntime(cli, *isComplete, true, true, true, nfs, mysql, ""); err != nil {
				util.ExitWithErrorMessage("Cellery setup create gcp command failed",
					fmt.Errorf("failed to create cellery runtime on gcp cluster, %v", err))
			}
		},
		Example: "  cellery setup create gcp",
	}
	return cmd
}

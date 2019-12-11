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
	"cellery.io/cellery/components/cli/pkg/minikube"
	"cellery.io/cellery/components/cli/pkg/util"
)

func newSetupCreateLocalCommand(cli cli.Cli, isComplete *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local",
		Short: "Create a local Cellery runtime in minikube",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if minikubeExists, err := minikube.ClusterExists(setup.CelleryLocalSetup); err != nil {
				util.ExitWithErrorMessage("Cellery setup create local command failed",
					fmt.Errorf("failed to check if minikube is running, %v", err))
			} else if minikubeExists {
				util.ExitWithErrorMessage("Cellery setup create local command failed",
					fmt.Errorf("minikube cluster with the profile name %s already exists", setup.CelleryLocalSetup))
			}
			platform, err := minikube.NewMinikube(
				minikube.SetProfile(setup.CelleryLocalSetup),
				minikube.SetCpus(setup.MinikubeCpus),
				minikube.SetMemory(setup.MinikubeMemory),
				minikube.SetkubeVersion(setup.MinikubeKubernetesVersion))
			if err != nil {
				util.ExitWithErrorMessage("Cellery setup create local command failed",
					fmt.Errorf("failed to initialize minikube platform, %v", err))
			}
			nodeportIp, mysql, nfs, err := setup.RunSetupCreateCelleryPlatform(cli, platform)
			if err != nil {
				util.ExitWithErrorMessage("Cellery setup create local command failed",
					fmt.Errorf("failed to create minikube platform, %v", err))
			}
			if err := setup.RunSetupCreateCelleryRuntime(cli, *isComplete, false, false, false, nfs, mysql, nodeportIp); err != nil {
				util.ExitWithErrorMessage("Cellery setup create local command failed",
					fmt.Errorf("failed to create cellery runtime on minikube cluster, %v", err))
			}
		},
		Example: "  cellery setup create local",
	}
	return cmd
}

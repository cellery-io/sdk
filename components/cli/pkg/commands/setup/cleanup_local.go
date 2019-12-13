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

package setup

import (
	"fmt"

	"github.com/manifoldco/promptui"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/minikube"
	"cellery.io/cellery/components/cli/pkg/util"
)

func manageLocal(cli cli.Cli) error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}
	var items []string
	minikubeStatus, err := minikube.ClusterStatus(CelleryLocalSetup)
	if err != nil {
		return fmt.Errorf("failed to check minikube status, %v", err)
	}
	if minikubeStatus == minikube.Running {
		items = []string{stop, cleanup, setupBack}
	} else {
		items = []string{start, cleanup, setupBack}
	}
	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select `cleanup` to remove local setup",
		Items:     items,
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select an option: %v", err)
	}
	platform, err := minikube.NewMinikube(
		minikube.SetProfile(CelleryLocalSetup))
	if err != nil {
		return fmt.Errorf("failed to initialize minikube platform, %v", err)
	}

	switch value {
	case start:
		{
			if err = cli.ExecuteTask("Starting local setup", "Failed to start local setup",
				"", func() error {
					return minikube.Start(CelleryLocalSetup)
				}); err != nil {
				return fmt.Errorf("error starting local setup: %v", err)
			}
			return nil
		}
	case stop:
		{
			if err = cli.ExecuteTask("Stopping local setup", "Failed to stop local setup",
				"", func() error {
					return minikube.Stop(CelleryLocalSetup)
				}); err != nil {
				return fmt.Errorf("error stopping local setup: %v", err)
			}
			return nil
		}
	case cleanup:
		{
			return RunSetupCleanupPlatform(cli, platform, false)
		}
	default:
		{
			return manageEnvironment(cli)
		}
	}
}

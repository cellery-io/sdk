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

func manageEnvironment(cli cli.Cli) error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}
	var items []string
	minikubeStatus, err := minikube.ClusterStatus(CelleryLocalSetup)
	if err != nil {
		return fmt.Errorf("failed to check if minikube is running, %v", err)
	}
	if minikubeStatus == minikube.NonExisting {
		items = []string{celleryGcp, existingCluster, setupBack}
	} else {
		items = []string{celleryLocal, celleryGcp, existingCluster, setupBack}
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select a runtime",
		Items:     items,
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select environment option to manage: %v", err)
	}

	switch value {
	case celleryLocal:
		{
			return manageLocal(cli)
		}
	case celleryGcp:
		{
			return manageGcp(cli)
		}
	case existingCluster:
		{
			return manageExistingCluster(cli)
		}
	default:
		{
			return RunSetup(cli)
		}
	}
}

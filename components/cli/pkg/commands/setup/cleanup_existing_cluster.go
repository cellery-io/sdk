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
	"cellery.io/cellery/components/cli/pkg/util"
)

func manageExistingCluster(cli cli.Cli) error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select `cleanup` to remove existing cluster",
		Items:     []string{cleanup, setupBack},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select an option: %v", err)
	}

	switch value {
	case cleanup:
		{
			return cleanupExistingCluster(cli)
		}
	default:
		{
			return manageEnvironment(cli)
		}
	}
}

func cleanupExistingCluster(cli cli.Cli) error {
	confirmCleanup, _, err := util.GetYesOrNoFromUser("Do you want to delete the cellery runtime (This will "+
		"delete all your cells and data)", false)
	if err != nil {
		return fmt.Errorf("failed to select option, %v", err)
	}
	if confirmCleanup {
		removeKnative, _, err := util.GetYesOrNoFromUser("Remove knative-serving", false)
		if err != nil {
			return fmt.Errorf("failed to select option, %v", err)
		}
		removeIstio, _, err := util.GetYesOrNoFromUser("Remove istio", false)
		if err != nil {
			return fmt.Errorf("failed to select option, %v", err)
		}
		removeIngress, _, err := util.GetYesOrNoFromUser("Remove ingress", false)
		if err != nil {
			return fmt.Errorf("failed to select option, %v", err)
		}
		removeHpa := false
		hpaEnabled, err := cli.Runtime().IsHpaEnabled()
		if hpaEnabled {
			removeHpa, _, err = util.GetYesOrNoFromUser("Remove hpa", false)
			if err != nil {
				return fmt.Errorf("failed to select option, %v", err)
			}
		}
		return RunSetupCleanupCelleryRuntime(cli, removeKnative, removeIstio, removeIngress, removeHpa, true)
	}
	return nil
}

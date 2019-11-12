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

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/vbox"
)

func manageEnvironment() error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label: util.YellowBold("?") + " Select a runtime",
		Items: []string{constants.CellerySetupLocal, constants.CellerySetupGcp,
			constants.CellerySetupExistingCluster, constants.CellerySetupBack},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select environment option to manage: %v", err)
	}

	switch value {
	case constants.CellerySetupLocal:
		{
			manageLocal()
		}
	case constants.CellerySetupGcp:
		{
			manageGcp()
		}
	case constants.CellerySetupExistingCluster:
		{
			manageExistingCluster()
		}
	default:
		{
			RunSetup()
		}
	}
	return nil
}

func getManageLabel() string {
	var manageLabel string
	if vbox.IsVmInstalled() {
		if vbox.IsVmRunning() {
			manageLabel = constants.VmName + " is running. Select `Stop` to stop the VM"
		} else {
			manageLabel = constants.VmName + " is installed. Select `Start` to start the VM"
		}
	} else {
		manageLabel = "Cellery runtime is not installed"
	}
	return manageLabel
}

func getManageEnvOptions() []string {
	if vbox.IsVmInstalled() {
		if vbox.IsVmRunning() {
			return []string{constants.CelleryManageStop, constants.CelleryManageCleanup, constants.CellerySetupBack}
		} else {
			return []string{constants.CelleryManageStart, constants.CelleryManageCleanup, constants.CellerySetupBack}
		}
	}
	return []string{constants.CellerySetupBack}
}

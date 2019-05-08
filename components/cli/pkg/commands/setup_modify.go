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

package commands

import (
	"github.com/manifoldco/promptui"

	"github.com/cellery-io/sdk/components/cli/pkg/runtime"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunSetupModify(addApimGlobalGateway, addObservability bool) {
	err := runtime.UpdateRuntime(addApimGlobalGateway, addObservability)
	if err != nil {
		util.ExitWithErrorMessage("Fail to modify the cluster", err)
	}
	util.WaitForRuntime()
}

func modifyRuntime() {
	const enable = "Enable"
	const disable = "Disable"
	const back = "BACK"

	enableApimgt := false
	enableObservability := false

	template := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	apiMgtPrompt := promptui.Select{
		Label:     util.YellowBold("?") + "API management and global gateway",
		Items:     []string{enable, disable, back},
		Templates: template,
	}
	_, apiMgtValue, err := apiMgtPrompt.Run()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option", err)
	}

	switch apiMgtValue {
	case enable:
		enableApimgt = true
	case disable:
		enableApimgt = false
	default:
		RunSetup()
		return
	}

	observabilityPrompt := promptui.Select{
		Label:     util.YellowBold("?") + "Observability",
		Items:     []string{enable, disable, back},
		Templates: template,
	}
	_, observabilityValue, err := observabilityPrompt.Run()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option", err)
	}

	switch observabilityValue {
	case enable:
		enableObservability = true
	case disable:
		enableObservability = false
	default:
		RunSetup()
		return
	}
	RunSetupModify(enableApimgt, enableObservability)
}

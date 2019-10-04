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
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/runtime"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunSetup() {
	selectTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label: util.YellowBold("?") + " Setup Cellery runtime",
		Items: []string{constants.CELLERY_SETUP_CREATE, constants.CELLERY_SETUP_MANAGE,
			constants.CELLERY_SETUP_MODIFY, constants.CELLERY_SETUP_SWITCH, constants.CELLERY_SETUP_EXIT},
		Templates: selectTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option: %v", err)
	}

	switch value {
	case constants.CELLERY_SETUP_MANAGE:
		{
			manageEnvironment()
		}
	case constants.CELLERY_SETUP_CREATE:
		{
			createEnvironment()
		}
	case constants.CELLERY_SETUP_MODIFY:
		{
			var err error
			apimEnabled, err = runtime.IsApimEnabled()
			if err != nil {
				util.ExitWithErrorMessage("Failed check if apim is enabled", err)
			}
			enableApim = !apimEnabled
			observabilityEnabled, err = runtime.IsObservabilityEnabled()
			if err != nil {
				util.ExitWithErrorMessage("Failed check if observability is enabled", err)
			}
			enableObservability = !observabilityEnabled
			knativeEnabled, err = runtime.IsKnativeEnabled()
			if err != nil {
				util.ExitWithErrorMessage("Failed check if knative is enabled", err)
			}
			enableKnative = !knativeEnabled
			hpaEnabled, err = runtime.IsHpaEnabled()
			if err != nil {
				util.ExitWithErrorMessage("Failed check if hpa is enabled", err)
			}
			enableHpa = !hpaEnabled
			modifyRuntime()
		}
	case constants.CELLERY_SETUP_SWITCH:
		{
			selectEnvironment()
		}
	default:
		{
			os.Exit(1)
		}
	}
}

func selectEnvironment() error {
	contexts := getContexts()
	contexts = append(contexts, constants.CELLERY_SETUP_BACK)
	bold := color.New(color.Bold).SprintFunc()
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Selected: bold("Selected cluster: ") + "{{ . }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select a Cellery Installed Kubernetes Cluster",
		Items:     contexts,
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select cluster: %v", err)
	}

	if value == constants.CELLERY_SETUP_BACK {
		RunSetup()
	}

	RunSwitchCommand(value)
	fmt.Printf(util.GreenBold("\n\U00002714") + " Successfully configured Cellery.\n")
	fmt.Println()
	fmt.Println(bold("What's next ?"))
	fmt.Println("======================")
	fmt.Println("To create your first project, execute the command: ")
	fmt.Println("  $ cellery init ")
	return nil
}

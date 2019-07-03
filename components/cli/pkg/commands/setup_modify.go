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
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/cellery-io/sdk/components/cli/pkg/runtime"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

type changedComponent struct {
	component string
	change    runtime.Selection
}

var apim = "API Manager"
var autoscaling = "Autoscaler"
var knative = "Scale-to-Zero"
var hpa = "Horizontal Pod Autoscaler"
var observability = "Observability"
var apimEnabled = false
var observabilityEnabled = false
var knativeEnabled = false
var hpaEnabled = false
var apimChange = runtime.NoChange
var observabilityChange = runtime.NoChange
var knativeChange = runtime.NoChange
var hpaChange = runtime.NoChange

func RunSetupModify(addApimGlobalGateway, addObservability, knative, hpa runtime.Selection) {
	err := runtime.UpdateRuntime(addApimGlobalGateway, addObservability, knative, hpa)
	if err != nil {
		util.ExitWithErrorMessage("Fail to modify the cluster", err)
	}
	knativeEnabled, err = runtime.IsKnativeEnabled()
	if err != nil {
		util.ExitWithErrorMessage("Error while checking knative status", err)
	}
	hpaEnabled, err = runtime.IsHpaEnabled()
	if err != nil {
		util.ExitWithErrorMessage("Error while checking hpa status", err)
	}
	util.WaitForRuntime(knativeEnabled, hpaEnabled)
}

func modifyRuntime() {
	var err error
	const done = "DONE"
	const back = "BACK"
	value := getPromptValue([]string{apim, autoscaling, observability, done, back}, "Select a runtime component")
	switch value {
	case apim:
		{
			apimEnabled, err = runtime.IsApimEnabled()
			if err != nil {
				util.ExitWithErrorMessage("Failed to get select options for apim", err)
			}
			apimChange = getComponentChange(apimEnabled, apim)
		}
	case observability:
		{
			observabilityEnabled, err = runtime.IsObservabilityEnabled()
			if err != nil {
				util.ExitWithErrorMessage("Failed to get select options for observability", err)
			}
			observabilityChange = getComponentChange(observabilityEnabled, observability)
		}
	case autoscaling:
		{
			modifyAutoScalingPolicy()
		}
	case done:
		{
			changes := []changedComponent{{apim, apimChange}, {observability,
				observabilityChange}, {knative, knativeChange}, {hpa, hpaChange}}
			enabledComponents, disabledComponents := getModifiedComponents(changes)
			// If total number of enabled and disabled components is greater than zero the runtime has changed
			runtimeUpdated := len(enabledComponents) > 0 || len(disabledComponents) > 0
			if !runtimeUpdated {
				fmt.Printf("No changes will be applied to the runtime\n")
			} else {
				fmt.Printf("Following modifications to the runtime will be applied\n")
				if len(enabledComponents) > 0 {
					// Print a comma separated list of enabled components
					enabledList := strings.Join(enabledComponents, ", ")
					fmt.Printf("Enabling : %s\n", enabledList)
				}
				if len(disabledComponents) > 0 {
					// Print a comma separated list of disabled components
					disabledList := strings.Join(disabledComponents, ", ")
					fmt.Printf("Disabling : %s\n", disabledList)
				}
			}
			confirmModify, _, err := util.GetYesOrNoFromUser("Do you wish to continue", false)
			if err != nil {
				util.ExitWithErrorMessage("Failed to select confirmation", err)
			}
			if confirmModify {
				if runtimeUpdated {
					RunSetupModify(apimChange, observabilityChange, knativeChange, hpaChange)
					os.Exit(0)
				}
			} else {
				modifyRuntime()
			}
			return
		}
	default:
		{
			// If back selected, remove all the saved state transitions of each components
			// This will delete all the inputs of user
			apimChange = runtime.NoChange
			observabilityChange = runtime.NoChange
			knativeChange = runtime.NoChange
			hpaChange = runtime.NoChange
			RunSetup()
			return
		}
	}
	modifyRuntime()
}

func modifyAutoScalingPolicy() {
	var err error
	const back = "BACK"
	value := getPromptValue([]string{knative, hpa, back}, "Select a runtime component")
	switch value {
	case knative:
		{
			knativeEnabled, err = runtime.IsKnativeEnabled()
			if err != nil {
				util.ExitWithErrorMessage("Failed to get select options for knative", err)
			}
			knativeChange = getComponentChange(knativeEnabled, knative)
		}
	case hpa:
		{
			hpaEnabled, err = runtime.IsHpaEnabled()
			if err != nil {
				util.ExitWithErrorMessage("Failed to get select options for hpa", err)
			}
			hpaChange = getComponentChange(hpaEnabled, hpa)
		}
	default:
		{
			// If user selects back he will get prompted back to modify runtime section
			return
		}
	}
	// Until back button is selected modifyAutoScalingPolicy function will get recursively called
	modifyAutoScalingPolicy()
}

func getComponentChange(componentEnabled bool, label string) runtime.Selection {
	const enable = "Enable"
	const disable = "Disable"
	const back = "BACK"
	template := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}
	option := enable
	if componentEnabled {
		option = disable
	}
	prompt := promptui.Select{
		Label:     util.YellowBold("?") + " " + label,
		Items:     []string{option, back},
		Templates: template,
	}
	_, value, err := prompt.Run()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option", err)
	}
	switch value {
	case enable:
		return runtime.Enable
	case disable:
		return runtime.Disable
	default:
		{
			return runtime.NoChange
		}
	}
}

func getModifiedComponents(changes []changedComponent) ([]string, []string) {
	var enabled []string
	var disabled []string
	for _, change := range changes {
		if change.change == runtime.Enable {
			enabled = append(enabled, change.component)
		}
		if change.change == runtime.Disable {
			disabled = append(disabled, change.component)
		}
	}
	return enabled, disabled
}

func getPromptValue(items []string, label string) string {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " " + label,
		Items:     items,
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option", err)
	}
	return value
}

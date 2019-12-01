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
	"os"
	"strings"

	"github.com/manifoldco/promptui"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
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
var enableApim = false
var enableObservability = false
var enableKnative = false
var enableHpa = false

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
	isGcpRuntime := runtime.IsGcpRuntime()
	if isGcpRuntime {
		hpaEnabled = false
	}
	runtime.WaitFor(knativeEnabled, hpaEnabled)
}

func modifyRuntime(cli cli.Cli) {
	const done = "Apply changes"
	const back = "BACK"
	value := getPromptValue([]string{apim, autoscaling, observability, done, back}, "Modify system components "+
		"and select Apply changes to apply the changes")
	switch value {
	case apim:
		{
			change := getComponentChange(&enableApim, apim)
			if change != runtime.NoChange {
				apimChange = change
				if confirmModification() {
					applyChanges(cli)
				} else {
					modifyRuntime(cli)
				}
				return
			}
		}
	case observability:
		{
			change := getComponentChange(&enableObservability, observability)
			if change != runtime.NoChange {
				observabilityChange = change
				if confirmModification() {
					applyChanges(cli)
				} else {
					modifyRuntime(cli)
				}
				return
			}
		}
	case autoscaling:
		{
			modifyAutoScalingPolicy(cli)
		}
	case done:
		{
			applyChanges(cli)
			return
		}
	default:
		{
			RunSetup(cli)
			return
		}
	}
	modifyRuntime(cli)
}

func modifyAutoScalingPolicy(cli cli.Cli) {
	var label = "Select system components to modify"
	var value = ""
	const back = "BACK"
	if runtime.IsGcpRuntime() {
		value = getPromptValue([]string{knative, back}, label)
	} else {
		value = getPromptValue([]string{knative, hpa, back}, label)
	}
	switch value {
	case knative:
		{
			change := getComponentChange(&enableKnative, knative)
			if change != runtime.NoChange {
				knativeChange = change
				if confirmModification() {
					applyChanges(cli)
				} else {
					modifyRuntime(cli)
				}
				return
			}
		}
	case hpa:
		{
			change := getComponentChange(&enableHpa, hpa)
			if change != runtime.NoChange {
				hpaChange = change
				if confirmModification() {
					applyChanges(cli)
				} else {
					modifyRuntime(cli)
				}
				return
			}
		}
	default:
		{
			// If user selects back he will get prompted back to modify runtime section
			return
		}
	}
	// Until back button is selected modifyAutoScalingPolicy function will get recursively called
	modifyAutoScalingPolicy(cli)
}

func getComponentChange(enableComponent *bool, label string) runtime.Selection {
	const enable = "Enable"
	const disable = "Disable"
	const back = "BACK"
	template := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}
	option := disable
	if *enableComponent {
		option = enable
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
		*enableComponent = false
		return runtime.Enable
	case disable:
		*enableComponent = true
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

func confirmModification() bool {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " " + "Do you want to modify another component",
		Items:     []string{"Yes", "No, Apply change"},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option", err)
	}
	switch value {
	case "Yes":
		return false
	default:
		return true
	}
}

func applyChanges(cli cli.Cli) {
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
		}
		os.Exit(0)
	} else {
		modifyRuntime(cli)
	}
}

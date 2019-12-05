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
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/constants"
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

func RunSetupModify(cli cli.Cli, apiManagement, observability, knative, hpa runtime.Selection) error {
	var err error
	cli.Runtime().SetArtifactsPath(filepath.Join(cli.FileSystem().CelleryInstallationDir(), constants.K8sArtifacts))
	observabilityEnabled, err := cli.Runtime().IsComponentEnabled(runtime.Observability)
	if err != nil {
		return err
	}
	if apiManagement != runtime.NoChange {
		// Remove observability if there was a change to apim
		if observabilityEnabled {
			err = cli.Runtime().DeleteComponent(runtime.Observability)
			if err != nil {
				return err
			}
		}
		if apiManagement == runtime.Enable {
			err = cli.Runtime().DeleteComponent(runtime.IdentityProvider)
			if err != nil {
				return err
			}
			err = cli.Runtime().AddComponent(runtime.ApiManager)
			if err != nil {
				return err
			}
		} else {
			err = cli.Runtime().DeleteComponent(runtime.ApiManager)
			if err != nil {
				return err
			}
			err = cli.Runtime().AddComponent(runtime.IdentityProvider)
			if err != nil {
				return err
			}
		}
		// Add observability if there was a change to apim and there was already observability running before that
		if observabilityEnabled {
			err = cli.Runtime().AddComponent(runtime.Observability)
			if err != nil {
				return err
			}
		}
	}
	if observability != runtime.NoChange {
		if observability == runtime.Enable {
			err = cli.Runtime().AddComponent(runtime.Observability)
			if err != nil {
				return err
			}
		} else {
			err = cli.Runtime().DeleteComponent(runtime.Observability)
			if err != nil {
				return err
			}
		}
	}
	if knative != runtime.NoChange {
		if knative == runtime.Enable {
			err = cli.Runtime().AddComponent(runtime.ScaleToZero)
			if err != nil {
				return err
			}
		} else {
			err = cli.Runtime().DeleteComponent(runtime.ScaleToZero)
			if err != nil {
				return err
			}
		}
	}
	if hpa != runtime.NoChange {
		if hpa == runtime.Enable {
			err = cli.Runtime().AddComponent(runtime.HPA)
			if err != nil {
				return err
			}
		} else {
			err = cli.Runtime().DeleteComponent(runtime.HPA)
			if err != nil {
				return err
			}
		}
	}
	knativeEnabled, err = cli.Runtime().IsComponentEnabled(runtime.ScaleToZero)
	if err != nil {
		return fmt.Errorf("error while checking knative status, %v", err)
	}
	hpaEnabled, err = cli.Runtime().IsComponentEnabled(runtime.HPA)
	if err != nil {
		return fmt.Errorf("error while checking hpa status, %v", err)
	}
	isGcpRuntime := cli.Runtime().IsGcpRuntime()
	if isGcpRuntime {
		hpaEnabled = false
	}
	return cli.Runtime().WaitFor(knativeEnabled, hpaEnabled)
}

func modifyRuntime(cli cli.Cli) error {
	const done = "Apply changes"
	const back = "BACK"
	value, err := getPromptValue([]string{apim, autoscaling, observability, done, back}, "Modify system components "+
		"and select Apply changes to apply the changes")
	if err != nil {
		return err
	}
	switch value {
	case apim:
		{
			change, err := getComponentChange(&enableApim, apim)
			if err != nil {
				return err
			}
			if change != runtime.NoChange {
				apimChange = change
				modificationConfirmed, err := confirmModification()
				if err != nil {
					return err
				}
				if modificationConfirmed {
					return applyChanges(cli)
				} else {
					return modifyRuntime(cli)
				}
			}
		}
	case observability:
		{
			change, err := getComponentChange(&enableObservability, observability)
			if err != nil {
				return err
			}
			if change != runtime.NoChange {
				observabilityChange = change
				modificationConfirmed, err := confirmModification()
				if err != nil {
					return err
				}
				if modificationConfirmed {
					return applyChanges(cli)
				} else {
					return modifyRuntime(cli)
				}
			}
		}
	case autoscaling:
		{
			modifyAutoScalingPolicy(cli)
		}
	case done:
		{
			return applyChanges(cli)
		}
	default:
		{
			RunSetup(cli)
			return nil
		}
	}
	return modifyRuntime(cli)
}

func modifyAutoScalingPolicy(cli cli.Cli) error {
	var label = "Select system components to modify"
	var value = ""
	var err error
	const back = "BACK"
	if cli.Runtime().IsGcpRuntime() {
		value, err = getPromptValue([]string{knative, back}, label)
		if err != nil {
			return err
		}
	} else {
		value, err = getPromptValue([]string{knative, hpa, back}, label)
		if err != nil {
			return err
		}
	}
	switch value {
	case knative:
		{
			change, err := getComponentChange(&enableKnative, knative)
			if err != nil {
				return err
			}
			if change != runtime.NoChange {
				knativeChange = change
				modificationConfirmed, err := confirmModification()
				if err != nil {
					return err
				}
				if modificationConfirmed {
					return applyChanges(cli)
				} else {
					return modifyRuntime(cli)
				}
			}
		}
	case hpa:
		{
			change, err := getComponentChange(&enableHpa, hpa)
			if err != nil {
				return err
			}
			if change != runtime.NoChange {
				hpaChange = change
				modificationConfirmed, err := confirmModification()
				if err != nil {
					return err
				}
				if modificationConfirmed {
					return applyChanges(cli)
				} else {
					return modifyRuntime(cli)
				}
			}
		}
	default:
		{
			// If user selects back he will get prompted back to modify runtime section
			return nil
		}
	}
	// Until back button is selected modifyAutoScalingPolicy function will get recursively called
	return modifyAutoScalingPolicy(cli)
}

func getComponentChange(enableComponent *bool, label string) (runtime.Selection, error) {
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
		return runtime.NoChange, fmt.Errorf("failed to select an option, %v", err)
	}
	switch value {
	case enable:
		*enableComponent = false
		return runtime.Enable, nil
	case disable:
		*enableComponent = true
		return runtime.Disable, nil
	default:
		{
			return runtime.NoChange, nil
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

func getPromptValue(items []string, label string) (string, error) {
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
		return "", fmt.Errorf("failed to select an option, %v", err)
	}
	return value, nil
}

func confirmModification() (bool, error) {
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
		return false, fmt.Errorf("failed to select an option, %v", err)
	}
	switch value {
	case "Yes":
		return false, nil
	default:
		return true, nil
	}
}

func applyChanges(cli cli.Cli) error {
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
		return fmt.Errorf("failed to select confirmation, %v", err)
	}
	if confirmModify {
		if runtimeUpdated {
			err := RunSetupModify(cli, apimChange, observabilityChange, knativeChange, hpaChange)
			if err != nil {
				return fmt.Errorf("failed to update cellery runtime, %v", err)
			}
		}
		os.Exit(0)
	} else {
		modifyRuntime(cli)
	}
	return nil
}

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

	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
	"cellery.io/cellery/components/cli/pkg/vbox"

	"github.com/manifoldco/promptui"
)

func manageLocal() error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " " + getManageLabel(),
		Items:     getManageEnvOptions(),
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select an option: %v", err)
	}

	switch value {
	case constants.CelleryManageStop:
		{
			spinner := util.StartNewSpinner("Stopping Cellery Runtime")
			defer func() {
				spinner.Stop(true)
			}()
			vbox.StopVm()
		}
	case constants.CelleryManageStart:
		{
			vbox.StartVm()
			runtime.WaitFor(false, false)
		}
	case constants.CelleryManageCleanup:
		{
			RunCleanupLocal(false, false)
		}
	default:
		{
			manageEnvironment()
		}
	}
	return nil
}

func RunCleanupLocal(removeVm, removeVmImage bool) error {
	var err error
	// Get the confirmation to remove cellery local runtime
	if !removeVm {
		removeVm, _, err = util.GetYesOrNoFromUser("Do you want to delete the cellery runtime (This will "+
			"delete all your cells and data)", false)
		if err != nil {
			util.ExitWithErrorMessage("failed get user confirmation", err)
		}
	}
	// Get the confirmation to delete vm image
	if !removeVmImage {
		existingImages := vbox.ImageExists()
		var imagesToDelete string
		if existingImages > vbox.None {
			if existingImages == vbox.Basic {
				imagesToDelete = vmBasic
			} else if existingImages == vbox.Complete {
				imagesToDelete = vmComplete
			} else {
				imagesToDelete = fmt.Sprintf("%s and %s", vmBasic, vmComplete)
			}
			removeVmImage, _, err = util.GetYesOrNoFromUser(fmt.Sprintf("Do you want to remove downloaded images. "+
				"This will delete %s", imagesToDelete), false)
			if err != nil {
				util.ExitWithErrorMessage("failed to get user permission to remove downloaded image", err)
			}
		}
	}
	// remove cellery-runtime-local vm
	if removeVm {
		err = removeLocalSetup()
		if err != nil {
			return err
		}
	}
	// remove downloaded image of vm
	if removeVmImage {
		vbox.RemoveVmImage()
	}
	return nil
}

func removeLocalSetup() error {
	var err error
	spinner := util.StartNewSpinner("Removing Cellery Runtime")
	defer func() {
		spinner.Stop(true)
	}()
	err = vbox.RemoveVm()
	if err != nil {
		return err
	}
	return nil
}

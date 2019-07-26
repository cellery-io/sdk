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

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/runtime"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/vbox"

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
	case constants.CELLERY_MANAGE_STOP:
		{
			spinner := util.StartNewSpinner("Stopping Cellery Runtime")
			defer func() {
				spinner.Stop(true)
			}()
			vbox.StopVm()
		}
	case constants.CELLERY_MANAGE_START:
		{
			vbox.StartVm()
			runtime.WaitFor(false, false)
		}
	case constants.CELLERY_MANAGE_CLEANUP:
		{
			RunCleanupLocal(false)
		}
	default:
		{
			manageEnvironment()
		}
	}
	return nil
}

func RunCleanupLocal(confirmed bool) error {
	var err error
	var confirmCleanup = confirmed
	removeImage := false
	if !confirmed {
		confirmCleanup, _, err = util.GetYesOrNoFromUser("Do you want to delete the cellery runtime (This will "+
			"delete all your cells and data)", false)
		if err != nil {
			util.ExitWithErrorMessage("failed get user confirmation", err)
		}
	}
	if confirmCleanup {
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
			removeImage, _, err = util.GetYesOrNoFromUser(fmt.Sprintf("Do you want to remove downloaded images. "+
				"This will delete %s", imagesToDelete), false)
			if err != nil {
				util.ExitWithErrorMessage("failed to get user permission to remove downloaded image", err)
			}
		}
		err = CleanupLocal(removeImage)
		if err != nil {
			return err
		}
	}
	return nil
}

func CleanupLocal(removeImage bool) error {
	spinner := util.StartNewSpinner("Removing Cellery Runtime")
	defer func() {
		spinner.Stop(true)
	}()
	err := vbox.RemoveVm(removeImage)
	if err != nil {
		return err
	}
	return nil
}

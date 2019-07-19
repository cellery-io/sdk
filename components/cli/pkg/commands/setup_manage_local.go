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
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/runtime"

	"github.com/manifoldco/promptui"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
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
			util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "controlvm", constants.VM_NAME, "acpipowerbutton"))
		}
	case constants.CELLERY_MANAGE_START:
		{
			util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "startvm", constants.VM_NAME, "--type", "headless"))
			runtime.WaitForRuntime(false, false)
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
	if !confirmed {
		confirmCleanup, _, err = util.GetYesOrNoFromUser("Do you want to delete the cellery runtime (This will "+
			"delete all your cells and data)", false)
		if err != nil {
			util.ExitWithErrorMessage("failed to select option", err)
		}
	}
	if confirmCleanup {
		err = CleanupLocal()
		if err != nil {
			return err
		}
	}
	return nil
}

func CleanupLocal() error {
	spinner := util.StartNewSpinner("Removing Cellery Runtime")
	defer func() {
		spinner.Stop(true)
	}()
	if isVmRuning() {
		util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "controlvm", constants.VM_NAME, "acpipowerbutton"))
	}
	for isVmRuning() {
		time.Sleep(2 * time.Second)
	}
	err := util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "unregistervm", constants.VM_NAME, "--delete"))
	if err != nil {
		return err
	}
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, vmComplete))
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, vmBasic))
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, configComplete))
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, configBasic))
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.VM_FILE_NAME))
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.VM_DISK_NAME))
	return nil
}

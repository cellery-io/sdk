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
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/manifoldco/promptui"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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
		return fmt.Errorf("Failed to select an option: %v", err)
	}

	switch value {
	case constants.CELLERY_MANAGE_STOP:
		{
			spinner := util.StartNewSpinner("Stopping Cellery Runtime")
			defer func() {
				spinner.Stop(true)
			}()
			err := util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "controlvm", constants.VM_NAME, "acpipowerbutton"))
			if err != nil {
				fmt.Printf("cellery : %v:\n", err)
				os.Exit(1)
			}
		}
	case constants.CELLERY_MANAGE_START:
		{
			err := util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "startvm", constants.VM_NAME, "--type", "headless"))
			if err != nil {
				fmt.Printf("cellery : %v:\n", err)
				os.Exit(1)
			}
		}
	case constants.CELLERY_MANAGE_CLEANUP:
		{
			err := cleanupLocal()
			if err != nil {
				fmt.Printf("cellery : %v:\n", err)
				os.Exit(1)
			}
		}
	default:
		{
			manageEnvironment()
		}
	}
	return nil
}

func cleanupLocal() error {
	spinner := util.StartNewSpinner("Removing Cellery Runtime")
	defer func() {
		spinner.Stop(true)
	}()
	if isVmRuning() {
		err := util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "controlvm", constants.VM_NAME, "acpipowerbutton"))
		if err != nil {
			return err
		}
	}
	for isVmRuning() {
		time.Sleep(2 * time.Second)
	}
	err := util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "unregistervm", constants.VM_NAME, "--delete"))
	if err != nil {
		return err
	}
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.AWS_S3_ITEM_VM))
	return nil
}

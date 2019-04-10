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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
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
		Items: []string{constants.CELLERY_SETUP_LOCAL, constants.CELLERY_SETUP_GCP,
			constants.CELLERY_SETUP_EXISTING_CLUSTER, constants.CELLERY_SETUP_BACK},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("Failed to select an option: %v", err)
	}

	switch value {
	case constants.CELLERY_SETUP_LOCAL:
		{
			manageLocal()
		}
	case constants.CELLERY_SETUP_GCP:
		{
			manageGcp()
		}
	case constants.CELLERY_SETUP_EXISTING_CLUSTER:
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
	if isVmInstalled() {
		if isVmRuning() {
			manageLabel = constants.VM_NAME + " is running. Select `Stop` to stop the VM"
		} else {
			manageLabel = constants.VM_NAME + " is installed. Select `Start` to start the VM"
		}
	} else {
		manageLabel = "Cellery runtime is not installed"
	}
	return manageLabel
}

func isVmRuning() bool {
	if isVmInstalled() {
		cmd := exec.Command(constants.VBOX_MANAGE, "showvminfo", constants.VM_NAME)
		stdoutReader, _ := cmd.StdoutPipe()
		stdoutScanner := bufio.NewScanner(stdoutReader)
		output := ""
		go func() {
			for stdoutScanner.Scan() {
				output = output + stdoutScanner.Text()
			}
		}()
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error occurred while checking VM status: %v \n", err)
			os.Exit(1)
		}
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("\x1b[31;1m Error occurred while checking VM status \x1b[0m %v \n", err)
			os.Exit(1)
		}
		if strings.Contains(output, "running (since") {
			return true
		}
	}
	return false
}

func getManageEnvOptions() []string {
	if isVmInstalled() {
		if isVmRuning() {
			return []string{constants.CELLERY_MANAGE_STOP, constants.CELLERY_MANAGE_CLEANUP, constants.CELLERY_SETUP_BACK}
		} else {
			return []string{constants.CELLERY_MANAGE_START, constants.CELLERY_MANAGE_CLEANUP, constants.CELLERY_SETUP_BACK}
		}
	}
	return []string{constants.CELLERY_SETUP_BACK}
}

func isVmInstalled() bool {
	cmd := exec.Command(constants.VBOX_MANAGE, "list", "vms")
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	output := ""
	go func() {
		for stdoutScanner.Scan() {
			output = output + stdoutScanner.Text()
		}
	}()
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error occurred while checking if VMs installed: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while checking if VMs installed \x1b[0m %v \n", err)
		os.Exit(1)
	}

	if strings.Contains(output, constants.VM_NAME) {
		return true
	}
	return false
}

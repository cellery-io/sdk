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

	"github.com/manifoldco/promptui"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunSetupCreateLocal(isCompleteSelected bool) {
	if util.IsCommandAvailable("VBoxManage") {
		util.ExitWithErrorMessage("Error creating VM", fmt.Errorf("VBoxManage not installed"))
	}
	if IsVmInstalled() {
		util.ExitWithErrorMessage("Error creating VM", fmt.Errorf("installed VM already exists"))
	}
	vmLocation := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM)
	repoCreateErr := util.CreateDir(vmLocation)
	if repoCreateErr != nil {
		os.Exit(1)
	}

	if isCompleteSelected {
		confirmDownload, _, err := util.GetYesOrNoFromUser(fmt.Sprintf("Downloading %s will take %s from your machine. Do you want to continue",
			constants.AWS_S3_ITEM_VM_COMPLETE,
			util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_VM_COMPLETE))), false)
		if err != nil {
			util.ExitWithErrorMessage("Failed to select an option", err)
		}
		if !confirmDownload {
			os.Exit(1)
		}
		fmt.Println("Downloading " + constants.AWS_S3_ITEM_VM_COMPLETE)
		util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_VM_COMPLETE, vmLocation,
			true)
		util.ExtractTarGzFile(vmLocation, filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM,
			constants.AWS_S3_ITEM_VM_COMPLETE))
		util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_CONFIG_COMPLETE, vmLocation,
			false)
		err = util.MergeKubeConfig(filepath.Join(util.UserHomeDir(),
			constants.CELLERY_HOME, constants.VM, constants.AWS_S3_ITEM_CONFIG_COMPLETE))
		if err != nil {
			util.ExitWithErrorMessage("Failed to merge kube-config file", err)
		}
	} else {
		confirmDownload, _, err := util.GetYesOrNoFromUser(fmt.Sprintf("Downloading %s will take %s from your machine. Do you want to continue",
			constants.AWS_S3_ITEM_VM_MINIMAL,
			util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_VM_MINIMAL))), false)
		if err != nil {
			util.ExitWithErrorMessage("Failed to select an option", err)
		}
		if !confirmDownload {
			os.Exit(1)
		}
		fmt.Println("Downloading " + constants.AWS_S3_ITEM_VM_MINIMAL)
		util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_VM_MINIMAL, vmLocation,
			true)
		util.ExtractTarGzFile(vmLocation, filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM,
			constants.AWS_S3_ITEM_VM_MINIMAL))
		util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_CONFIG_MINIMAL, vmLocation,
			false)
		err = util.MergeKubeConfig(filepath.Join(util.UserHomeDir(),
			constants.CELLERY_HOME, constants.VM, constants.AWS_S3_ITEM_CONFIG_MINIMAL))
		if err != nil {
			util.ExitWithErrorMessage("Failed to merge kube-config file", err)
		}
	}
	installVM()
	util.WaitForRuntime()
}

func createLocal() error {
	var isCompleteSelected = false
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	sizeMinimal := util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_VM_MINIMAL))
	sizeComplete := util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_VM_COMPLETE))

	cellPrompt := promptui.Select{
		Label: util.YellowBold("?") + " Select the type of runtime",
		Items: []string{
			fmt.Sprintf("%s (size: %s)", constants.BASIC, sizeMinimal),
			fmt.Sprintf("%s (size: %s)", constants.COMPLETE, sizeComplete),
			constants.CELLERY_SETUP_BACK,
		},
		Templates: cellTemplate,
	}
	index, _, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("Failed to select an option: %v", err)
	}
	if index == 2 {
		createEnvironment()
		return nil
	}
	if index == 1 {
		isCompleteSelected = true
	}
	RunSetupCreateLocal(isCompleteSelected)

	return nil
}

func installVM() error {
	vmPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.VM_FILE_NAME)
	spinner := util.StartNewSpinner("Installing Cellery Runtime")

	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "hostonlyif", "create"),
		"Error Installing VM")
	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "hostonlyif", "ipconfig", "vboxnet0",
		"--ip", "192.168.56.1"), "Error Installing VM")

	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "import", vmPath), "Error Installing VM")
	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "modifyvm", constants.VM_NAME,
		"--ostype", "Ubuntu_64", "--cpus", "2", "--memory", "8000", "--natpf1", "guestkube,tcp,,6443,,6443", "--natpf1",
		"guestssh,tcp,,2222,,22", "--natpf1", "guesthttps,tcp,,443,,443", "--natpf1", "guesthttp,tcp,,80,,80"),
		"Error Installing VM")
	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "startvm", constants.VM_NAME, "--type", "headless"),
		"Error Installing VM")

	spinner.Stop(true)
	return nil
}

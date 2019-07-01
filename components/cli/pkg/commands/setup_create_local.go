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
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

var vmComplete = fmt.Sprintf("cellery-runtime-complete-%s.tar.gz", version.BuildVersion())
var vmBasic = fmt.Sprintf("cellery-runtime-basic-%s.tar.gz", version.BuildVersion())
var configComplete = fmt.Sprintf("config-cellery-runtime-complete-%s", version.BuildVersion())
var configBasic = fmt.Sprintf("config-cellery-runtime-basic-%s", version.BuildVersion())

func RunSetupCreateLocal(isCompleteSelected, confirmed bool) {
	var err error
	var confirmDownload = confirmed
	if !util.IsCommandAvailable("VBoxManage") {
		util.ExitWithErrorMessage("Error creating VM", fmt.Errorf("VBoxManage not installed"))
	}
	if IsVmInstalled() {
		util.ExitWithErrorMessage("Error creating VM", fmt.Errorf("installed VM already exists"))
	}
	vmLocation := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM)
	repoCreateErr := util.CreateDir(vmLocation)
	if repoCreateErr != nil {
		util.ExitWithErrorMessage("Failed to create vm directory", err)
	}

	if isCompleteSelected {
		if !confirmed {
			confirmDownload, _, err = util.GetYesOrNoFromUser(fmt.Sprintf(
				"Downloading %s will take %s from your machine. Do you want to continue",
				vmComplete,
				util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, vmComplete))),
				false)
			if err != nil {
				util.ExitWithErrorMessage("Failed to select an option", err)
			}
		}
		if !confirmDownload {
			os.Exit(1)
		}
		fmt.Println("Downloading " + vmComplete)
		util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, vmComplete, vmLocation,
			true)
		util.ExtractTarGzFile(vmLocation, filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM,
			vmComplete))
		util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, configComplete, vmLocation,
			false)
		err = util.MergeKubeConfig(filepath.Join(util.UserHomeDir(),
			constants.CELLERY_HOME, constants.VM, configComplete))
		if err != nil {
			util.ExitWithErrorMessage("Failed to merge kube-config file", err)
		}
	} else {
		if !confirmed {
			confirmDownload, _, err = util.GetYesOrNoFromUser(fmt.Sprintf("Downloading %s will take %s from your "+
				"machine. Do you want to continue",
				vmBasic,
				util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, vmBasic))),
				false)
			if err != nil {
				util.ExitWithErrorMessage("Failed to select an option", err)
			}
		}
		if !confirmDownload {
			os.Exit(1)
		}
		fmt.Println("Downloading " + vmBasic)
		util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, vmBasic, vmLocation,
			true)
		util.ExtractTarGzFile(vmLocation, filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM,
			vmBasic))
		util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, configBasic, vmLocation,
			false)
		err = util.MergeKubeConfig(filepath.Join(util.UserHomeDir(),
			constants.CELLERY_HOME, constants.VM, configBasic))
		if err != nil {
			util.ExitWithErrorMessage("Failed to merge kube-config file", err)
		}
	}
	installVM()
	util.WaitForRuntime(false, false)
}

func createLocal() error {
	var isCompleteSelected = false
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}
	sizeMinimal := util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, vmBasic))
	sizeComplete := util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, vmComplete))

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
		return fmt.Errorf("failed to select an option: %v", err)
	}
	if index == 2 {
		createEnvironment()
		return nil
	}
	if index == 1 {
		isCompleteSelected = true
	}
	RunSetupCreateLocal(isCompleteSelected, false)

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

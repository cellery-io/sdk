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
	"os"
	"os/exec"
	"path/filepath"
)

func RunSetupCreateLocal() {
	
}

func createLocal() error {
	var addGlobalGateway bool
	var addObservability bool

	addGlobalGateway, err := util.GetYesOrNoFromUser("Add Global Gateway")
	if err != nil {
		fmt.Printf("Error while creating VM location: %v", err)
		os.Exit(1)
	}
	if addGlobalGateway {
		addObservability, err = util.GetYesOrNoFromUser("Add Observability")
	}

	vmLocation := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM)
	repoCreateErr := util.CreateDir(vmLocation)
	if repoCreateErr != nil {
		fmt.Println("Error while creating VM location: " + repoCreateErr.Error())
		os.Exit(1)
	}
	spinner := util.StartNewSpinner("Downloading Cellery Runtime")
	defer func() {
		spinner.Stop(true)
	}()

	if !addGlobalGateway && !addObservability {
		createBasicLocalEnvironment(vmLocation)
	}

	if addGlobalGateway && !addObservability {
		createLocalWithGlobalGateway(vmLocation)
	}

	if addGlobalGateway && addObservability {
		createLocalEnvironmentWithObservability(vmLocation)
	}
	util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_CONFIG, vmLocation)
	util.ReplaceFile(filepath.Join(util.UserHomeDir(), ".kube", "config"), filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.AWS_S3_ITEM_CONFIG))
	installVM()

	return nil
}

func createBasicLocalEnvironment(vmLocation string) error {
	util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_VM, vmLocation)
	util.ExtractTarGzFile(vmLocation, filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.AWS_S3_ITEM_VM))

	return nil
}

func createLocalWithGlobalGateway(vmLocation string) error {
	util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_VM, vmLocation)
	util.ExtractTarGzFile(vmLocation, filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.AWS_S3_ITEM_VM))

	return nil
}

func createLocalEnvironmentWithObservability(vmLocation string) error {
	util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_VM, vmLocation)
	util.ExtractTarGzFile(vmLocation, filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.AWS_S3_ITEM_VM))

	return nil
}

func installVM() error {
	spinner := util.StartNewSpinner("Installing Cellery Runtime")
	defer func() {
		spinner.Stop(true)
	}()
	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "createvm", "--name", constants.VM_NAME, "--ostype", "Ubuntu_64", "--register"), "Error Installing VM")
	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "modifyvm", constants.VM_NAME, "--ostype", "Ubuntu_64", "--cpus", "4", "--memory", "8000", "--natpf1", "guestkube,tcp,,6443,,6443", "--natpf1", "guestssh,tcp,,2222,,22", "--natpf1", "guesthttps,tcp,,443,,443", "--natpf1", "guesthttp,tcp,,80,,80"), "Error Installing VM")
	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "storagectl", constants.VM_NAME, "--name", "hd1", "--add", "sata", "--portcount", "2"), "Error Installing VM")

	vmLocation := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.VM_FILE_NAME)
	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "storageattach", constants.VM_NAME, "--storagectl", "hd1", "--port", "1", "--type", "hdd", "--medium", vmLocation), "Error Installing VM")
	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "startvm", constants.VM_NAME, "--type", "headless"), "Error Installing VM")

	return nil
}

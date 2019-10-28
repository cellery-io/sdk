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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/runtime"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/vbox"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

var vmComplete = fmt.Sprintf("cellery-runtime-complete-%s.tar.gz", version.BuildVersion())
var vmBasic = fmt.Sprintf("cellery-runtime-basic-%s.tar.gz", version.BuildVersion())
var configComplete = fmt.Sprintf("config-cellery-runtime-complete-%s", version.BuildVersion())
var configBasic = fmt.Sprintf("config-cellery-runtime-basic-%s", version.BuildVersion())
var md5Complete = fmt.Sprintf("cellery-runtime-complete-%s.md5", version.BuildVersion())
var md5Basic = fmt.Sprintf("cellery-runtime-basic-%s.md5", version.BuildVersion())

var downloadLocation = filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM)
var downloadTempLocation = filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.TMP, constants.VM)

var vmPath string
var md5Path string
var vm string
var config string
var md5 string
var etag string

func RunSetupCreateLocal(isCompleteSelected, forceDownload, confirmed bool) {
	var err error
	var downloadVm = false
	if !util.IsCommandAvailable("VBoxManage") {
		util.ExitWithErrorMessage("Error creating VM", fmt.Errorf("VBoxManage not installed"))
	}
	if vbox.IsVmInstalled() {
		util.ExitWithErrorMessage("Error creating VM", fmt.Errorf("installed VM already exists"))
	}
	if isCompleteSelected {
		vm = vmComplete
		config = configComplete
		md5 = md5Complete
	} else {
		vm = vmBasic
		config = configBasic
		md5 = md5Basic
	}
	// Set global variables
	vmPath = filepath.Join(downloadLocation, vm)
	md5Path = filepath.Join(downloadLocation, md5)
	if forceDownload {
		// If force-download flag is set download the vm regardless whether already downloaded image exists or not
		downloadVm = true
	} else {
		// Check if the vm should be downloaded
		downloadVm = downloadVmConfirmation(confirmed)
	}
	if downloadVm {
		// If the download confirmed proceed to download for aws
		fmt.Println("Downloading " + vm)
		// Create a temporary location to download vm
		util.RemoveDir(downloadTempLocation)
		repoCreateErr := util.CreateDir(downloadTempLocation)
		if repoCreateErr != nil {
			util.ExitWithErrorMessage("Failed to create vm directory", err)
		}
		etag = util.GetS3ObjectEtag(constants.AWS_S3_BUCKET, vm)
		util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, vm, downloadTempLocation, true)
		util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, config, downloadTempLocation, false)
		// Copy the files if the download was successful
		util.CopyFile(filepath.Join(downloadTempLocation, vm), filepath.Join(downloadLocation, vm))
		util.CopyFile(filepath.Join(downloadTempLocation, config), filepath.Join(downloadLocation, config))
		util.RemoveDir(md5Path)
		err = ioutil.WriteFile(md5Path, []byte(etag), 0666)
		if err != nil {
			util.ExitWithErrorMessage(fmt.Sprintf("Error writing md5 value to %s", md5), err)
		}
		util.RemoveDir(downloadTempLocation)
	}
	// Merge the kube config file
	err = util.MergeKubeConfig(filepath.Join(downloadLocation, config))
	if err != nil {
		util.ExitWithErrorMessage("Failed to merge kube-config file", err)
	}
	fmt.Printf("Extracting %s ...", vm)
	err = util.ExtractTarGzFile(downloadLocation, vmPath)
	if err != nil {
		util.ExitWithErrorMessage("Failed to extract vm", err)
	}
	err = vbox.InstallVM(isCompleteSelected)
	if err != nil {
		util.ExitWithErrorMessage("Failed to install vm", err)
	}
	runtime.WaitFor(false, false)
}

func downloadVmConfirmation(confirmed bool) bool {
	var err error
	vmFileExists := false
	md5FileExists := false
	downloadVm := false
	vmFileExists, err = util.FileExists(vmPath)
	if err != nil {
		util.ExitWithErrorMessage(fmt.Sprintf("Error checking if %s exists", vm), err)
	}
	// Check if previously downloaded image of same version exists
	if vmFileExists {
		// If vm image is already downloaded check whether an updated version exists
		md5FileExists, err = util.FileExists(md5Path)
		if err != nil {
			util.ExitWithErrorMessage(fmt.Sprintf("Error checking if %s exists", md5), err)
		}
		if md5FileExists {
			md5, err := ioutil.ReadFile(md5Path)
			if err != nil {
				util.ExitWithErrorMessage(fmt.Sprintf("Failed to read %s", md5Path), err)
			}
			// Check if updated version of vm is available
			fmt.Println("Checking if updated cellery runtime is available ...")
			etag = util.GetS3ObjectEtag(constants.AWS_S3_BUCKET, vm)
			if !(strings.Contains(string(md5), etag)) {
				// md5 file confirms the availability of an updated image of vm
				if !confirmed {
					// If the user wish to get the updated vm download the latest vm
					downloadVm, _, err = util.GetYesOrNoFromUser(fmt.Sprintf(
						"Updated version of %s is available and will take %s from your "+
							"machine. Do you want to use the upgraded version",
						vm, util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, vm))), false)
					if !downloadVm {
						// Not exiting CLI since user response "No" suggests using old image
						fmt.Println("Installing existing cellery-local-setup")
						downloadVm = false
					}
				} else {
					// When inline command (cellery setup create local) is executed download confirmation will not be prompted
					downloadVm = true
				}
			} else {
				// If the existing vm is same as the one available in aws, then do not download
				fmt.Println("Latest cellery-local-setup is already downloaded")
				downloadVm = false
			}
		} else {
			// If md5 file cannot be found download the latest vm
			if !confirmed {
				confirmDownload()
			}
			downloadVm = true
		}
	} else {
		// If previously downloaded image does not exist download the latest image
		if !confirmed {
			confirmDownload()
		}
		downloadVm = true
	}
	return downloadVm
}

func createLocal() error {
	var isCompleteSelected = false
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}
	sizeBasic := util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, vmBasic))
	sizeComplete := util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, vmComplete))

	cellPrompt := promptui.Select{
		Label: util.YellowBold("?") + " Select the type of runtime",
		Items: []string{
			fmt.Sprintf("%s (size: %s)", constants.BASIC, sizeBasic),
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
	RunSetupCreateLocal(isCompleteSelected, false, false)
	return nil
}

func confirmDownload() {
	confirm, _, err := util.GetYesOrNoFromUser(fmt.Sprintf(
		"Downloading %s will take %s from your machine. Do you want to continue", vm,
		util.FormatBytesToString(util.GetS3ObjectSize(constants.AWS_S3_BUCKET, vm))),
		false)
	if err != nil {
		util.ExitWithErrorMessage("Failed to get user confirmation to download", err)
	}
	if !confirm {
		os.Exit(0)
	}
}

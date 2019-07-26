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
 *
 */

package vbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/osexec"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

type Exists int

const (
	None Exists = iota
	Basic
	Complete
	Both
)

var vmComplete = fmt.Sprintf("cellery-runtime-complete-%s.tar.gz", version.BuildVersion())
var vmBasic = fmt.Sprintf("cellery-runtime-basic-%s.tar.gz", version.BuildVersion())
var configComplete = fmt.Sprintf("config-cellery-runtime-complete-%s", version.BuildVersion())
var configBasic = fmt.Sprintf("config-cellery-runtime-basic-%s", version.BuildVersion())

func InstallVM(isComplete bool) error {
	var err error
	vmPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, vm, vmFileName)
	spinner := util.StartNewSpinner("Installing Cellery Runtime")

	if _, err = osexec.GetCommandOutput(exec.Command(vBoxManage, hostOnlyIf, create)); err != nil {
		spinner.Stop(false)
		return err
	}
	if _, err = osexec.GetCommandOutput(exec.Command(vBoxManage, hostOnlyIf, ipconfig, vboxnet0,
		argIp, vboxIp,
	)); err != nil {
		spinner.Stop(false)
		return err
	}
	if _, err = osexec.GetCommandOutput(exec.Command(vBoxManage, importVm, vmPath)); err != nil {
		spinner.Stop(false)
		return err
	}
	cores := "2"
	if isComplete {
		cores = "4"
	}
	if _, err = osexec.GetCommandOutput(exec.Command(vBoxManage, modifyVm, vmName,
		argOsType, ubuntu64,
		argCpus, cores,
		argMemory, "8192",
		argNatpf1, "guestkube,tcp,,6443,,6443",
		argNatpf1, "guestssh,tcp,,2222,,22",
		argNatpf1, "guesthttps,tcp,,443,,443",
		argNatpf1, "guesthttp,tcp,,80,,80",
	)); err != nil {
		spinner.Stop(false)
		return err
	}
	if _, err = osexec.GetCommandOutput(exec.Command(vBoxManage, startVm, vmName,
		argType, headless)); err != nil {
		spinner.Stop(false)
		return err
	}
	spinner.Stop(true)
	return nil
}

func IsVmRunning() bool {
	var err error
	var output string
	if IsVmInstalled() {
		cmd := exec.Command(vBoxManage, "showvminfo", vmName)
		output, err = osexec.GetCommandOutput(cmd)
		if strings.Contains(output, "running (since") {
			return true
		}
		if err != nil {
			util.ExitWithErrorMessage("Error checking if vm is running", err)
		}
	}
	return false
}

func IsVmInstalled() bool {
	var err error
	var output string
	cmd := exec.Command(vBoxManage, argList, vms)
	output, err = osexec.GetCommandOutput(cmd)
	if strings.Contains(output, "running (since") {
		return true
	}
	if err != nil {
		util.ExitWithErrorMessage("Error checking if vm is installed", err)
	}
	if strings.Contains(output, vmName) {
		return true
	}
	return false
}

func StartVm() {
	var err error
	cmd := exec.Command(vBoxManage, startVm, vmName, argType, headless)
	_, err = osexec.GetCommandOutput(cmd)
	if err != nil {
		util.ExitWithErrorMessage("Error starting vm", err)
	}
}

func StopVm() {
	var err error
	cmd := exec.Command(vBoxManage, controlvm, vmName, acpipowerbutton)
	_, err = osexec.GetCommandOutput(cmd)
	if err != nil {
		util.ExitWithErrorMessage("Error stopping vm", err)
	}
}

func RemoveVm(removeImage bool) error {
	var err error
	if IsVmRunning() {
		if _, err = osexec.GetCommandOutput(exec.Command(vBoxManage, controlvm, vmName, acpipowerbutton)); err != nil {
			return err
		}
	}
	for IsVmRunning() {
		time.Sleep(2 * time.Second)
	}
	if _, err = osexec.GetCommandOutput(exec.Command(vBoxManage, unregistervm, vmName, argDelete)); err != nil {
		return err
	}
	if removeImage {
		// If user does not wish to retain the downloaded image delete the file from the system
		os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, vm, vmComplete))
		os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, vm, vmBasic))
		os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, vm, configComplete))
		os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, vm, configBasic))
	}
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, vm, vmFileName))
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, vm, vmDiskName))
	return nil
}
func ImageExists() Exists {
	exists := None
	var err error
	vmCompleteExists := false
	vmBasicExists := false
	var vmCompletePath = filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, vm, vmComplete)
	var vmBasicPath = filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, vm, vmBasic)

	vmCompleteExists, err = util.FileExists(vmCompletePath)
	if err != nil {
		util.ExitWithErrorMessage(fmt.Sprintf("Error checking if %s exists", vmComplete), err)
	}
	vmBasicExists, err = util.FileExists(vmBasicPath)
	if err != nil {
		util.ExitWithErrorMessage(fmt.Sprintf("Error checking if %s exists", vmBasic), err)
	}
	if vmBasicExists {
		exists = Basic
	}
	if vmCompleteExists {
		exists = Complete
	}
	if vmBasicExists && vmCompleteExists {
		exists = Both
	}
	return exists
}

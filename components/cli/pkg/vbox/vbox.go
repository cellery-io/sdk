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
	"os/exec"
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/osexec"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func InstallVM(isComplete bool) error {
	var err error
	vmPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.VM_FILE_NAME)
	spinner := util.StartNewSpinner("Installing Cellery Runtime")

	if _, err = osexec.GetCommandOutput(exec.Command(constants.VBOX_MANAGE, hostOnlyIf, create)); err != nil {
		spinner.Stop(false)
		return err
	}
	if _, err = osexec.GetCommandOutput(exec.Command(constants.VBOX_MANAGE, hostOnlyIf, ipconfig, vboxnet0,
		argIp, vboxIp,
	)); err != nil {
		spinner.Stop(false)
		return err
	}
	if _, err = osexec.GetCommandOutput(exec.Command(constants.VBOX_MANAGE, importVm, vmPath)); err != nil {
		spinner.Stop(false)
		return err
	}
	cores := "2"
	if isComplete {
		cores = "4"
	}
	if _, err = osexec.GetCommandOutput(exec.Command(constants.VBOX_MANAGE, modifyVm, constants.VM_NAME,
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
	if _, err = osexec.GetCommandOutput(exec.Command(constants.VBOX_MANAGE, startVm, constants.VM_NAME,
		argType, headless)); err != nil {
		spinner.Stop(false)
		return err
	}
	spinner.Stop(true)
	return nil
}

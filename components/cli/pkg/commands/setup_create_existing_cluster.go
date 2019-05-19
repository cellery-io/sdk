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
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/runtime"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/oxequa/interact"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func createOnExistingCluster() error {
	var isCompleteSetup = false
	var isPersistentVolume = false
	var hasNfsStorage = false
	var isLoadBalancerIngressMode = false
	var isBackSelected = false
	var nfs runtime.Nfs
	var db runtime.MysqlDb
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select the type of runtime",
		Items:     []string{constants.PERSISTENT_VOLUME, constants.NON_PERSISTENT_VOLUME, constants.CELLERY_SETUP_BACK},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("Failed to select an option: %v", err)
	}
	if value == constants.CELLERY_SETUP_BACK {
		createEnvironment()
		return nil
	}
	if value == constants.PERSISTENT_VOLUME {
		isPersistentVolume = true

		hasNfsStorage, isBackSelected, err = util.GetYesOrNoFromUser(fmt.Sprintf("Use NFS server"),
			true)
		if err != nil {
			util.ExitWithErrorMessage("Failed to select an option", err)
		}
		if hasNfsStorage {
			nfs, db, err = getPersistentVolumeDataWithNfs()
		}
		if isBackSelected {
			createOnExistingCluster()
			return nil
		}
	}
	isCompleteSetup, isBackSelected = util.IsCompleteSetupSelected()
	if isBackSelected {
		createOnExistingCluster()
		return nil
	}
	isLoadBalancerIngressMode, isBackSelected = util.IsLoadBalancerIngressTypeSelected()
	if isBackSelected {
		createOnExistingCluster()
		return nil
	}

	if err != nil {
		return fmt.Errorf("Failed to get user input: %v", err)
	}
	RunSetupCreateOnExistingCluster(isCompleteSetup, isPersistentVolume, hasNfsStorage, isLoadBalancerIngressMode, nfs, db)

	return nil
}

func RunSetupCreateOnExistingCluster(isCompleteSetup, isPersistentVolume, hasNfsStorage, isLoadBalancerIngressMode bool,
	nfs runtime.Nfs, db runtime.MysqlDb) {
	artifactsPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.K8S_ARTIFACTS)
	os.RemoveAll(artifactsPath)
	util.CopyDir(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS), artifactsPath)
	if err := runtime.CreateRuntime(artifactsPath, isCompleteSetup, isPersistentVolume, hasNfsStorage,
		isLoadBalancerIngressMode, nfs, db); err != nil {
		fmt.Printf("Error deploying cellery runtime: %v", err)
	}
	util.WaitForRuntime()
}

func getPersistentVolumeDataWithNfs() (runtime.Nfs, runtime.MysqlDb, error) {
	prefix := util.CyanBold("?")
	nfsServerIp := ""
	fileShare := ""
	dbHostName := ""
	dbUserName := ""
	dbPassword := ""
	err := interact.Run(&interact.Interact{
		Before: func(c interact.Context) error {
			c.SetPrfx(color.Output, prefix)
			return nil
		},
		Questions: []*interact.Question{
			{
				Before: func(c interact.Context) error {
					c.SetPrfx(nil, util.CyanBold("?"))
					return nil
				},
				Quest: interact.Quest{
					Msg: util.Bold("NFS server IP: "),
				},
				Action: func(c interact.Context) interface{} {
					nfsServerIp, _ = c.Ans().String()
					return nil
				},
			},
			{
				Before: func(c interact.Context) error {
					c.SetPrfx(nil, util.CyanBold("?"))
					return nil
				},
				Quest: interact.Quest{
					Msg: util.Bold("File share name: "),
				},
				Action: func(c interact.Context) interface{} {
					fileShare, _ = c.Ans().String()
					return nil
				},
			},
			{
				Before: func(c interact.Context) error {
					c.SetPrfx(nil, util.CyanBold("?"))
					return nil
				},
				Quest: interact.Quest{
					Msg: util.Bold("Database host: "),
				},
				Action: func(c interact.Context) interface{} {
					dbHostName, _ = c.Ans().String()
					return nil
				},
			},
		},
	})
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while getting user input", err)
	}
	dbUserName, dbPassword, err = util.RequestCredentials("Mysql", "")
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while getting user input", err)
	}
	return runtime.Nfs{nfsServerIp, fileShare},
		runtime.MysqlDb{dbHostName, dbUserName, dbPassword}, nil
}

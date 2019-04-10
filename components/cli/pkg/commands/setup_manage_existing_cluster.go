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

	"github.com/manifoldco/promptui"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func manageExistingCluster() error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select `cleanup` to remove an existing GCP cluster",
		Items:     []string{constants.CELLERY_MANAGE_CLEANUP, constants.CELLERY_SETUP_BACK},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("Failed to select an option: %v", err)
	}

	switch value {
	case constants.CELLERY_MANAGE_CLEANUP:
		{
			cleanupExistingCluster()
		}
	default:
		{
			manageEnvironment()
		}
	}
	return nil
}

func cleanupExistingCluster() error {
	confirmCleanup, _, err := util.GetYesOrNoFromUser("Do you want to delete the cellery runtime (This will "+
		"delete all your cells and data)", false)
	if err != nil {
		util.ExitWithErrorMessage("failed to select option", err)
	}
	if confirmCleanup {
		removeIstio, _, err := util.GetYesOrNoFromUser("Remove istio", false)
		if err != nil {
			util.ExitWithErrorMessage("failed to select option", err)
		}
		removeIngress, _, err := util.GetYesOrNoFromUser("Remove ingress", false)
		if err != nil {
			util.ExitWithErrorMessage("failed to select option", err)
		}
		gcpSpinner := util.StartNewSpinner("Cleaning up cluster")

		kubectl.DeleteNameSpace("cellery-system")
		if removeIstio {
			kubectl.DeleteNameSpace("istio-system")
		}
		if removeIngress {
			kubectl.DeleteNameSpace("ingress-nginx")
		}
		kubectl.DeleteAllCells()
		kubectl.DeletePersistedVolume("wso2apim-local-pv")
		kubectl.DeletePersistedVolume("wso2apim-with-analytics-mysql-pv")
		gcpSpinner.Stop(true)
	}
	return nil
}

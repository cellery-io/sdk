/*
 * Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package setup

import (
	"fmt"

	"github.com/manifoldco/promptui"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
)

func manageExistingCluster(cli cli.Cli) error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select `cleanup` to remove existing cluster",
		Items:     []string{cleanup, setupBack},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select an option: %v", err)
	}

	switch value {
	case cleanup:
		{
			cleanupExistingCluster(cli)
		}
	default:
		{
			manageEnvironment(cli)
		}
	}
	return nil
}

func cleanupExistingCluster(cli cli.Cli) error {
	confirmCleanup, _, err := util.GetYesOrNoFromUser("Do you want to delete the cellery runtime (This will "+
		"delete all your cells and data)", false)
	if err != nil {
		util.ExitWithErrorMessage("failed to select option", err)
	}
	if confirmCleanup {
		removeKnative, _, err := util.GetYesOrNoFromUser("Remove knative-serving", false)
		if err != nil {
			util.ExitWithErrorMessage("failed to select option", err)
		}
		removeIstio, _, err := util.GetYesOrNoFromUser("Remove istio", false)
		if err != nil {
			util.ExitWithErrorMessage("failed to select option", err)
		}
		removeIngress, _, err := util.GetYesOrNoFromUser("Remove ingress", false)
		if err != nil {
			util.ExitWithErrorMessage("failed to select option", err)
		}
		removeHpa := false
		hpaEnabled, err := cli.Runtime().IsHpaEnabled()
		if hpaEnabled {
			removeHpa, _, err = util.GetYesOrNoFromUser("Remove hpa", false)
			if err != nil {
				util.ExitWithErrorMessage("failed to select option", err)
			}
		}
		spinner := util.StartNewSpinner("Cleaning up cluster")
		cleanupCluster(cli, removeKnative, removeIstio, removeIngress, removeHpa)
		spinner.Stop(true)
	}
	return nil
}

func RunCleanupExisting(cli cli.Cli, removeKnative, removeIstio, removeIngress, removeHpa, confirmed bool) error {
	var err error
	var confirmCleanup = confirmed
	if !confirmed {
		confirmCleanup, _, err = util.GetYesOrNoFromUser("Do you want to delete the cellery runtime (This will "+
			"delete all your cells and data)", false)
		if err != nil {
			util.ExitWithErrorMessage("failed to select option", err)
		}
	}
	if confirmCleanup {
		spinner := util.StartNewSpinner("Cleaning up cluster")
		if removeKnative {
			kubernetes.DeleteNameSpace("knative-serving")
		}
		cleanupCluster(cli, removeKnative, removeIstio, removeIngress, removeHpa)
		spinner.Stop(true)
	}
	return nil
}

func cleanupCluster(cli cli.Cli, removeKnative, removeIstio, removeIngress, removeHpa bool) {
	kubernetes.DeleteNameSpace("cellery-system")
	if removeKnative {
		out, err := kubernetes.DeleteResource("apiservices.apiregistration.k8s.io", "v1beta1.custom.metrics.k8s.io")
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while deleting the knative apiservice", fmt.Errorf(out))
		}
		kubernetes.DeleteNameSpace("knative-serving")
	}
	if removeIstio {
		kubernetes.DeleteNameSpace("istio-system")
	}
	if removeIngress {
		kubernetes.DeleteNameSpace("ingress-nginx")
	}
	if removeHpa {
		cli.Runtime().DeleteComponent(runtime.HPA)
	}
	kubernetes.DeleteAllCells()
	kubernetes.DeletePersistedVolume("wso2apim-local-pv")
	kubernetes.DeletePersistedVolume("wso2apim-with-analytics-mysql-pv")
}

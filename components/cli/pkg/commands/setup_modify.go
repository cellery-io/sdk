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
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunSetupModify(addApimGlobalGateway, addObservability bool) {
	util.CopyDir(filepath.Join(util.CelleryInstallationDir(), "artifacts"),
		filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS))
	artifactPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS)
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"

	if addApimGlobalGateway {
		// Remove IDP
		removeIdp(artifactPath, errorDeployingCelleryRuntime)

		// Create APIM global gateway
		executeAPIMArtifacts(artifactPath, errorDeployingCelleryRuntime)
	}
	if addObservability {
		executeObservabilityArtifacts(artifactPath, errorDeployingCelleryRuntime, false)
	}
}

func modifyRuntime() {
	addApimGlobalGateway, err := util.GetYesOrNoFromUser("Add apim global gateway")
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option", err)
	}
	if !addApimGlobalGateway {
		os.Exit(1)
	}
	addObservability, err := util.GetYesOrNoFromUser("Add observability")
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option", err)
	}
	if !addObservability {
		os.Exit(1)
	}
	RunSetupModify(addApimGlobalGateway, addObservability)
}

func removeIdp(artifactPath, errorDeployingCelleryRuntime string) {
	// Delete the IDP config maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.DELETE, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-idp/conf", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.DELETE, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-idp/conf/datasources", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.DELETE, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-idp/conf/identity", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.DELETE, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-idp/conf/tomcat", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Delete IDP deployment and the service
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.DELETE, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-idp/global-idp.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)

}

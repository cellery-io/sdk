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

func RunSetupModify(addObservability bool) {
	util.CopyDir(util.CelleryInstallationDir(), filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS))
	artifactPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS)
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"

	if addObservability {
		executeObservabilityArtifacts(artifactPath, errorDeployingCelleryRuntime)
	}
}

func executeControllerArtifacts(artifactPath, errorDeployingCelleryRuntime string) error {
	// Give permission
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, "clusterrolebinding", "cluster-admin-binding", "--clusterrole", "cluster-admin", "--user", accountName), errorDeployingCelleryRuntime)

	// Setup Celley namespace, create service account and the docker registry credentials
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/ns-init.yaml"), errorDeployingCelleryRuntime)

	// Istio
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/istio-crds.yaml"), errorDeployingCelleryRuntime)

	// Enabling Istio injection
	util.ExecuteCommand(exec.Command(constants.KUBECTL, "label", "namespace", "default", "istio-injection=enabled"), errorDeployingCelleryRuntime)

	// Without security
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/istio-demo-cellery.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/istio-gateway.yaml"), errorDeployingCelleryRuntime)

	// Install Cellery crds
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/01-cluster-role.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/02-service-account.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/03-cluster-role-binding.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/04-crd-cell.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/05-crd-gateway.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/06-crd-token-service.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/07-crd-service.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/08-config.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/09-controller.yaml"), errorDeployingCelleryRuntime)
	return nil
}

func executeAPIMArtifacts(artifactPath, errorDeployingCelleryRuntime string) error {
	// Create apim NFS volumes and volume claims
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-apim/artifacts-persistent-volume.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-apim/artifacts-persistent-volume-claim.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Create the gw config maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "gw-conf", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "gw-conf-datasources", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/datasources/", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Create KM config maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "conf-identity", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/identity", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "apim-template", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/resources/api_templates", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "apim-tomcat", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/tomcat", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "apim-security", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/security", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	//Create gateway deployment and the service
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-apim/global-apim.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Install nginx-ingress for control plane ingress
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/mandatory.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/cloud-generic.yaml"), errorDeployingCelleryRuntime)
	return nil
}

func executeObservabilityArtifacts(artifactPath, errorDeployingCelleryRuntime string) error {
	//// Create SP worker configmaps
	//util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "sp-worker-conf", "--from-file", artifactPath+"/k8s-artefacts/observability/sp/conf", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	//util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "sp-worker-siddhi", "--from-file", artifactPath+"/k8s-artefacts/observability/siddhi", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	//util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "observability-portal-config", "--from-file", artifactPath+"/k8s-artefacts/observability/node-server/config", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Create SP worker deployment
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/observability/sp/sp-worker.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Create observability portal deployment, service and ingress.
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/observability/portal/observability-portal.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Create K8s Metrics Config-maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "k8s-metrics-prometheus-conf", "--from-file", artifactPath+"/k8s-artefacts/observability/prometheus/config", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "k8s-metrics-grafana-conf", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/config", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "k8s-metrics-grafana-datasources", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/datasources", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "k8s-metrics-grafana-dashboards", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/dashboards", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "k8s-metrics-grafana-dashboards-default", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/dashboards/default", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Create K8s Metrics deployment, service and ingress.
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/observability/prometheus/k8s-metrics-prometheus.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/observability/grafana/k8s-metrics-grafana.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	return nil
}

func modifyRuntime() {
	addObservability, err := util.GetYesOrNoFromUser("Add observability")
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option", err)
	}
	if !addObservability {
		os.Exit(1)
	}
	RunSetupModify(addObservability)
}

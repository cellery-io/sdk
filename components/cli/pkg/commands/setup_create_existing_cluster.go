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
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"

	"github.com/manifoldco/promptui"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func createOnExistingCluster() error {
	var isPersistedVolumeVolume = false
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select the type of runtime",
		Items:     []string{constants.PERSISTENT_VOLUME, constants.NON_PERSISTENT_VOLUME},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("Failed to select an option: %v", err)
	}
	if value == constants.PERSISTENT_VOLUME {
		isPersistedVolumeVolume = true
	}
	RunSetupCreateOnExistingCluster(isPersistedVolumeVolume)

	return nil
}

func RunSetupCreateOnExistingCluster(isPersistedVolumeVolume bool) {
	if isPersistedVolumeVolume {
		createRuntimeOnExistingClusterWithPersistedVolume()
	} else {
		createRuntimeOnExistingClusterWithNonPersistedVolume()
	}
}

func createRuntimeOnExistingClusterWithPersistedVolume() {
	gcpSpinner := util.StartNewSpinner("Creating cellery runtime")
	// Backup folders
	util.RenameFile(filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP, constants.CELLERY, constants.MYSQL),
		filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP, constants.CELLERY, constants.MYSQL)+"-old")
	util.RenameFile(filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP, constants.CELLERY,
		constants.APIM_REPOSITORY_DEPLOYMENT_SERVER), filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP,
		constants.CELLERY, constants.APIM_REPOSITORY_DEPLOYMENT_SERVER)+"-old")

	// Create folders required by the mysql PVC
	util.CreateDir(filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP, constants.CELLERY, constants.MYSQL))

	// Create folders required by the APIM PVC
	util.CreateDir(filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP, constants.CELLERY,
		constants.APIM_REPOSITORY_DEPLOYMENT_SERVER))

	updateK8sArtifacts()

	var artifactPath = filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS)
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"

	gcpSpinner.SetNewAction("Creating controller")
	executeControllerArtifacts(artifactPath, errorDeployingCelleryRuntime)

	gcpSpinner.SetNewAction("Creating APIM")
	executeAPIMArtifacts(artifactPath, errorDeployingCelleryRuntime)

	gcpSpinner.SetNewAction("Creating Observability")
	executeObservabilityArtifacts(artifactPath, errorDeployingCelleryRuntime, true)

	// Create ingress-nginx deployment
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/system/mandatory.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/system/service-nodeport.yaml"), errorDeployingCelleryRuntime)
	gcpSpinner.Stop(true)
}

func createRuntimeOnExistingClusterWithNonPersistedVolume() {
	gcpSpinner := util.StartNewSpinner("Creating cellery runtime")

	updateK8sArtifacts()

	var artifactPath = filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS)
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"

	gcpSpinner.SetNewAction("Creating controller")
	executeControllerArtifacts(artifactPath, errorDeployingCelleryRuntime)

	gcpSpinner.SetNewAction("Creating APIM")
	executeAPIMArtifactsForNonPersistedVolume(artifactPath, errorDeployingCelleryRuntime)

	gcpSpinner.SetNewAction("Creating Observability")
	executeObservabilityArtifacts(artifactPath, errorDeployingCelleryRuntime, true)

	// Create ingress-nginx deployment
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/system/mandatory.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/system/service-nodeport.yaml"), errorDeployingCelleryRuntime)
	gcpSpinner.Stop(true)
}

func updateK8sArtifacts() {
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS_OLD))
	util.CopyDir(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS),
		filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS_OLD))
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS))
	util.CopyDir(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS), filepath.Join(util.UserHomeDir(),
		constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS))
	// Replace username, password, host in /global-apim/conf/datasources/master-datasources.xml
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP,
		constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.GLOBAL_APIM, constants.CONF, constants.DATA_SOURCES,
		constants.MASTER_DATA_SOURCES_XML), constants.DATABASE_USERNAME, constants.GCP_SQL_USER_NAME, -1); err != nil {
		fmt.Printf("%v: %v", constants.ERROR_REPLACING_APIM_MASTER_DATASOURCES_XML, err)
	}

	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP,
		constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.GLOBAL_APIM, constants.CONF, constants.DATA_SOURCES,
		constants.MASTER_DATA_SOURCES_XML), constants.DATABASE_PASSWORD, constants.GCP_SQL_PASSWORD, -1); err != nil {
		fmt.Printf("%v: %v", constants.ERROR_REPLACING_APIM_MASTER_DATASOURCES_XML, err)
	}

	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP,
		constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.GLOBAL_APIM, constants.CONF, constants.DATA_SOURCES,
		constants.MASTER_DATA_SOURCES_XML), constants.MYSQL_DATABASE_HOST,
		constants.MYSQL_HOST_NAME_FOR_EXISTING_CLUSTER, -1); err != nil {
		fmt.Printf("%v: %v", constants.ERROR_REPLACING_APIM_MASTER_DATASOURCES_XML, err)
	}
	// Replace username username, password, host in /observability/sp/conf/deployment.yaml
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME,
		constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.OBSERVABILITY,
		constants.SP, constants.CONF, constants.DEPLOYMENT_YAML), constants.DATABASE_USERNAME,
		constants.GCP_SQL_USER_NAME, -1); err != nil {
		fmt.Printf("%V: %v", constants.ERROR_REPLACING_OBSERVABILITY_YAML, err)
	}
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP,
		constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.OBSERVABILITY, constants.SP,
		constants.CONF, constants.DEPLOYMENT_YAML), constants.DATABASE_PASSWORD,
		constants.GCP_SQL_PASSWORD, -1); err != nil {
		fmt.Printf("%V: %v", constants.ERROR_REPLACING_OBSERVABILITY_YAML, err)
	}
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP,
		constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.OBSERVABILITY, constants.SP, constants.CONF,
		constants.DEPLOYMENT_YAML), constants.MYSQL_DATABASE_HOST, constants.MYSQL_HOST_NAME_FOR_EXISTING_CLUSTER,
		-1); err != nil {
		fmt.Printf("%V: %v", constants.ERROR_REPLACING_OBSERVABILITY_YAML, err)
	}

	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP,
		constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.MYSQL, constants.DB_SCRIPTS, constants.INIT_SQL),
		constants.DATABASE_USERNAME, constants.GCP_SQL_USER_NAME, -1); err != nil {
		fmt.Printf("%V: %v", constants.ERROR_REPLACING_INIT_SQL, err)
	}

	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP,
		constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.MYSQL, constants.DB_SCRIPTS, constants.INIT_SQL),
		constants.DATABASE_PASSWORD, constants.GCP_SQL_PASSWORD, -1); err != nil {
		fmt.Printf("%V: %v", constants.ERROR_REPLACING_INIT_SQL, err)
	}
}

func executeControllerArtifacts(artifactPath, errorDeployingCelleryRuntime string) {
	nodeName, err := kubectl.GetMasterNodeName()
	if err != nil {
		util.ExitWithErrorMessage("Failed to create controller", err)
	}
	// Setup Celley namespace, create service account and the docker registry credentials
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/system/ns-init.yaml"), errorDeployingCelleryRuntime)

	// Label the node
	util.ExecuteCommand(exec.Command(constants.KUBECTL, "label", "nodes", nodeName, "disk=local"),
		errorDeployingCelleryRuntime)
	time.Sleep(60 * time.Second)

	// Istio
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/system/istio-crds.yaml"), errorDeployingCelleryRuntime)
	// Install Istio
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/system/istio-demo-cellery.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/system/istio-gateway.yaml"), errorDeployingCelleryRuntime)

	// Enabling Istio injection
	util.ExecuteCommand(exec.Command(constants.KUBECTL, "label", "namespace", "default", "istio-injection=enabled"),
		errorDeployingCelleryRuntime)

	// Install Cellery crds
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/controller/01-cluster-role.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/controller/02-service-account.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/controller/03-cluster-role-binding.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/controller/04-crd-cell.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/controller/05-crd-gateway.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/controller/06-crd-token-service.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/controller/07-crd-service.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/controller/08-config.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/controller/09-autoscale-policy.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/controller/10-controller.yaml"), errorDeployingCelleryRuntime)

	time.Sleep(120 * time.Second)
}

func executeAPIMArtifacts(artifactPath, errorDeployingCelleryRuntime string) {
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"mysql-dbscripts", "--from-file", artifactPath+"/k8s-artefacts/mysql/dbscripts/", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/mysql/mysql-persistent-volumes-local-dev.yaml", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/mysql/mysql-persistent-volume-claim.yaml", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/mysql/mysql-deployment.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Wait till the mysql deployment availability
	util.ExecuteCommand(exec.Command(constants.KUBECTL, "wait", "deployment/wso2apim-with-analytics-mysql-deployment",
		"--for", "condition=available", "--timeout", "300s", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/mysql/mysql-service.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Create apim NFS volumes and volume claims
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/global-apim/persistent-volume-local-dev.yaml", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/global-apim/persistent-volume-claim-local.yaml", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)

	// Create the gw config maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "gw-conf", "--from-file",
		artifactPath+"/k8s-artefacts/global-apim/conf", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "gw-conf-datasources",
		"--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/datasources/", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)

	// Create KM config maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"conf-identity", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/identity",
		"-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"apim-template", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/resources/api_templates",
		"-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"apim-tomcat", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/tomcat", "-n",
		"cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"apim-security", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/security", "-n",
		"cellery-system"), errorDeployingCelleryRuntime)

	//Create gateway deployment and the service
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/global-apim/global-apim.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Wait till the gateway deployment availability
	util.ExecuteCommand(exec.Command(constants.KUBECTL, "wait", "deployment.apps/gateway", "--for",
		"condition=available", "--timeout", "600s", "-n", "cellery-system"), errorDeployingCelleryRuntime)
}

func executeAPIMArtifactsForNonPersistedVolume(artifactPath, errorDeployingCelleryRuntime string) {
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"mysql-dbscripts", "--from-file", artifactPath+"/k8s-artefacts/mysql/dbscripts/", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)

	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/mysql/mysql-deployment-volatile.yaml", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)

	// Wait till the mysql deployment availability
	util.ExecuteCommand(exec.Command(constants.KUBECTL, "wait", "deployment/wso2apim-with-analytics-mysql-deployment",
		"--for", "condition=available", "--timeout", "300s", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/mysql/mysql-service.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Create the gw config maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "gw-conf", "--from-file",
		artifactPath+"/k8s-artefacts/global-apim/conf", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "gw-conf-datasources",
		"--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/datasources/", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)

	// Create KM config maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"conf-identity", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/identity",
		"-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"apim-template", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/resources/api_templates",
		"-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"apim-tomcat", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/tomcat", "-n",
		"cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"apim-security", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/security", "-n",
		"cellery-system"), errorDeployingCelleryRuntime)

	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/global-apim/global-apim-volatile.yaml", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)
	// Wait till the gateway deployment availability
	util.ExecuteCommand(exec.Command(constants.KUBECTL, "wait", "deployment.apps/gateway", "--for",
		"condition=available", "--timeout", "600s", "-n", "cellery-system"), errorDeployingCelleryRuntime)
}

func executeObservabilityArtifacts(artifactPath, errorDeployingCelleryRuntime string, existingCluster bool) error {
	if existingCluster {
		// Create SP worker configmaps
		util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
			"sp-worker-conf", "--from-file", artifactPath+"/k8s-artefacts/observability/sp/conf",
			"-n", "cellery-system"), errorDeployingCelleryRuntime)
		util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
			"sp-worker-siddhi", "--from-file", artifactPath+"/k8s-artefacts/observability/siddhi",
			"-n", "cellery-system"), errorDeployingCelleryRuntime)
		util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
			"observability-portal-config", "--from-file", artifactPath+"/k8s-artefacts/observability/node-server/config",
			"-n", "cellery-system"), errorDeployingCelleryRuntime)
	}
	// Create SP worker deployment
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/observability/sp/sp-worker.yaml", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)
	// Create observability portal deployment, service and ingress.
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/observability/portal/observability-portal.yaml", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)
	// Create K8s Metrics Config-maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"k8s-metrics-prometheus-conf", "--from-file", artifactPath+"/k8s-artefacts/observability/prometheus/config",
		"-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"k8s-metrics-grafana-conf", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/config", "-n",
		"cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"k8s-metrics-grafana-datasources", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/datasources",
		"-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"k8s-metrics-grafana-dashboards", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/dashboards",
		"-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP,
		"k8s-metrics-grafana-dashboards-default", "--from-file",
		artifactPath+"/k8s-artefacts/observability/grafana/dashboards/default", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)
	// Create K8s Metrics deployment, service and ingress.
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/observability/prometheus/k8s-metrics-prometheus.yaml", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG,
		artifactPath+"/k8s-artefacts/observability/grafana/k8s-metrics-grafana.yaml", "-n", "cellery-system"),
		errorDeployingCelleryRuntime)

	return nil
}

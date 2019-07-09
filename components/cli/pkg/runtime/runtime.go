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

package runtime

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

type Selection int

const (
	NoChange Selection = iota
	Enable
	Disable
)

var isCompleteSetup = false

func SetCompleteSetup(completeSetup bool) {
	isCompleteSetup = completeSetup
}

func CreateRuntime(artifactsPath string, isPersistentVolume, hasNfsStorage, isLoadBalancerIngressMode bool, nfs Nfs,
	db MysqlDb) error {
	spinner := util.StartNewSpinner("Creating cellery runtime")
	if isPersistentVolume && !hasNfsStorage {
		createFoldersRequiredForMysqlPvc()
		createFoldersRequiredForApimPvc()
	}
	dbHostName := constants.MYSQL_HOST_NAME_FOR_EXISTING_CLUSTER
	dbUserName := constants.CELLERY_SQL_USER_NAME
	dbPassword := constants.CELLERY_SQL_PASSWORD
	if hasNfsStorage {
		dbHostName = db.DbHostName
		dbUserName = db.DbUserName
		dbPassword = db.DbPassword
		updateNfsServerDetails(nfs.NfsServerIp, nfs.FileShare, artifactsPath)
	}
	if err := updateMysqlCredentials(dbUserName, dbPassword, dbHostName, artifactsPath); err != nil {
		spinner.Stop(false)
		return fmt.Errorf("error updating mysql credentials: %v", err)
	}
	if err := updateInitSql(dbUserName, dbPassword, artifactsPath); err != nil {
		spinner.Stop(false)
		return fmt.Errorf("error updating mysql init script: %v", err)
	}

	if isPersistentVolume && !IsGcpRuntime() {
		nodeName, err := kubectl.GetMasterNodeName()
		if err != nil {
			return fmt.Errorf("error getting master node name: %v", err)
		}
		if err := kubectl.ApplyLable("nodes", nodeName, "disk=local", true); err != nil {
			return fmt.Errorf("error applying master node lable: %v", err)
		}
	}
	// Setup Cellery namespace
	spinner.SetNewAction("Setting up cellery namespace")
	if err := CreateCelleryNameSpace(); err != nil {
		return fmt.Errorf("error creating cellery namespace: %v", err)
	}

	// Apply Istio CRDs
	spinner.SetNewAction("Applying istio crds")
	if err := ApplyIstioCrds(artifactsPath); err != nil {
		return fmt.Errorf("error creating istio crds: %v", err)
	}
	// Apply nginx resources
	spinner.SetNewAction("Creating ingress-nginx")
	if err := installNginx(artifactsPath, isLoadBalancerIngressMode); err != nil {
		return fmt.Errorf("error installing ingress-nginx: %v", err)
	}
	// sleep for few seconds - this is to make sure that the CRDs are properly applied
	time.Sleep(20 * time.Second)

	// Enabling Istio injection
	spinner.SetNewAction("Enabling istio injection")
	if err := kubectl.ApplyLable("namespace", "default", "istio-injection=enabled",
		true); err != nil {
		return fmt.Errorf("error enabling istio injection: %v", err)
	}

	// Install istio
	spinner.SetNewAction("Installing istio")
	if err := InstallIstio(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS)); err != nil {
		return fmt.Errorf("error installing istio: %v", err)
	}

	// Fix for applying only knative serving CRD's
	if err := InstallKnativeServing(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS)); err != nil {
		return fmt.Errorf("error installing knative serving: %v", err)
	}

	if err := kubectl.DeleteNameSpace("knative-serving"); err != nil {
		return fmt.Errorf("error removing knative-serving namespace: %v", err)
	}

	// Apply controller CRDs
	spinner.SetNewAction("Creating controller")
	if err := InstallController(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS)); err != nil {
		return fmt.Errorf("error creating cellery controller: %v", err)
	}

	spinner.SetNewAction("Configuring mysql")
	if err := AddMysql(artifactsPath, isPersistentVolume); err != nil {
		return fmt.Errorf("error configuring mysql: %v", err)
	}

	spinner.SetNewAction("Creating ConfigMaps")
	if err := CreateGlobalGatewayConfigMaps(artifactsPath); err != nil {
		return fmt.Errorf("error creating gateway configmaps: %v", err)
	}
	if err := CreateObservabilityConfigMaps(artifactsPath); err != nil {
		return fmt.Errorf("error creating observability configmaps: %v", err)
	}
	if err := CreateIdpConfigMaps(artifactsPath); err != nil {
		return fmt.Errorf("error creating idp configmaps: %v", err)
	}

	if isPersistentVolume {
		spinner.SetNewAction("Creating Persistent Volume")
		if err := createPersistentVolume(artifactsPath, hasNfsStorage); err != nil {
			return fmt.Errorf("error creating persistent volume: %v", err)
		}
	}

	if isCompleteSetup {
		spinner.SetNewAction("Adding apim")
		if err := addApim(artifactsPath, isPersistentVolume); err != nil {
			return fmt.Errorf("error creating apim deployment: %v", err)
		}
		spinner.SetNewAction("Adding observability")
		if err := addObservability(artifactsPath); err != nil {
			return fmt.Errorf("error creating observability deployment: %v", err)
		}
	} else {
		spinner.SetNewAction("Adding idp")
		if err := addIdp(artifactsPath); err != nil {
			return fmt.Errorf("error creating idp deployment: %v", err)
		}
	}
	spinner.Stop(true)
	return nil
}

func UpdateRuntime(apiManagement, observability, knative, hpa Selection) error {
	spinner := util.StartNewSpinner("Updating cellery runtime")
	var err error
	observabilityEnabled, err := IsObservabilityEnabled()
	if err != nil {
		spinner.Stop(false)
		return err
	}
	if apiManagement != NoChange {
		// Remove observability if there was a change to apim
		if observabilityEnabled {
			err = DeleteComponent(Observability)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		}
		if apiManagement == Enable {
			err = DeleteComponent(IdentityProvider)
			if err != nil {
				spinner.Stop(false)
				return err
			}
			err = AddComponent(ApiManager)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		} else {
			err = DeleteComponent(ApiManager)
			if err != nil {
				spinner.Stop(false)
				return err
			}
			err = AddComponent(IdentityProvider)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		}
		// Add observability if there was a change to apim and there was already observability running before that
		if observabilityEnabled {
			err = AddComponent(Observability)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		}
	}
	if observability != NoChange {
		if observability == Enable {
			err = AddComponent(Observability)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		} else {
			err = DeleteComponent(Observability)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		}
	}
	if knative != NoChange {
		if knative == Enable {
			err = AddComponent(ScaleToZero)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		} else {
			err = DeleteComponent(ScaleToZero)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		}
	}
	if hpa != NoChange {
		if hpa == Enable {
			err = AddComponent(HPA)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		} else {
			err = DeleteComponent(HPA)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		}
	}
	spinner.Stop(true)
	return nil
}

func AddComponent(component SystemComponent) error {
	switch component {
	case ApiManager:
		return addApim(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS), false)
	case IdentityProvider:
		return addIdp(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS))
	case Observability:
		return addObservability(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS))
	case ScaleToZero:
		return InstallKnativeServing(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS))
	case HPA:
		return InstallHPA(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS))
	default:
		return fmt.Errorf("unknown system componenet %q", component)
	}
}

func DeleteComponent(component SystemComponent) error {
	switch component {
	case ApiManager:
		return deleteApim(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS))
	case IdentityProvider:
		return deleteIdp(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS))
	case Observability:
		return deleteObservability(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS))
	case ScaleToZero:
		return deleteKnative()
	case HPA:
		return deleteHpa(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS))
	default:
		return fmt.Errorf("unknown system componenet %q", component)
	}
}

func IsComponentEnabled(component SystemComponent) (bool, error) {
	switch component {
	case ApiManager:
		return IsApimEnabled()
	case Observability:
		return IsObservabilityEnabled()
	case ScaleToZero:
		return IsKnativeEnabled()
	case HPA:
		return IsHpaEnabled()
	default:
		return false, fmt.Errorf("unknown system componenet %q", component)
	}
}

func createFoldersRequiredForMysqlPvc() {
	// Backup folders
	util.RenameFile(filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP, constants.CELLERY, constants.MYSQL),
		filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP, constants.CELLERY, constants.MYSQL)+"-old")
	// Create folders required by the mysql PVC
	util.CreateDir(filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP, constants.CELLERY, constants.MYSQL))
}

func createFoldersRequiredForApimPvc() {
	// Backup folders
	util.RenameFile(filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP, constants.CELLERY,
		constants.APIM_REPOSITORY_DEPLOYMENT_SERVER), filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP,
		constants.CELLERY, constants.APIM_REPOSITORY_DEPLOYMENT_SERVER)+"-old")
	// Create folders required by the APIM PVC
	util.CreateDir(filepath.Join(constants.ROOT_DIR, constants.VAR, constants.TMP, constants.CELLERY,
		constants.APIM_REPOSITORY_DEPLOYMENT_SERVER))
}

func buildArtifactsPath(component SystemComponent, artifactsPath string) string {
	switch component {
	case ApiManager:
		return filepath.Join(artifactsPath, "global-apim")
	case IdentityProvider:
		return filepath.Join(artifactsPath, "global-idp")
	case Observability:
		return filepath.Join(artifactsPath, "observability")
	case Controller:
		return filepath.Join(artifactsPath, "controller")
	case System:
		return filepath.Join(artifactsPath, "system")
	case Mysql:
		return filepath.Join(artifactsPath, "mysql")
	case HPA:
		return filepath.Join(artifactsPath, "metrics-server/")
	default:
		return filepath.Join(artifactsPath)
	}
}

func IsGcpRuntime() bool {
	nodes, err := kubectl.GetNodes()
	if err != nil {
		util.ExitWithErrorMessage("failed to check if runtime is gcp", err)
	}
	for _, node := range nodes.Items {
		version := node.Status.NodeInfo.KubeletVersion
		if strings.Contains(version, "gke") {
			return true
		}
	}
	return false
}

func GetInstancesNames() ([]string, error) {
	var instances []string
	runningInstances, err := kubectl.GetCells()
	if err != nil {
		return nil, err
	}
	for _, runningInstance := range runningInstances.Items {
		instances = append(instances, runningInstance.CellMetaData.Name)
	}
	return instances, nil
}

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
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

type ConfigMap struct {
	Name string
	Path string
}

type Nfs struct {
	NfsServerIp string
	FileShare   string
}
type MysqlDb struct {
	DbHostName string
	DbUserName string
	DbPassword string
}

func CreateRuntime(artifactsPath string, isCompleteSetup, isPersistentVolume, hasNfsStorage,
	isLoadBalancerIngressMode bool, nfs Nfs, db MysqlDb) error {
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
		fmt.Printf("Error updating file: %v", err)
	}
	if err := updateInitSql(dbUserName, dbPassword, artifactsPath); err != nil {
		spinner.Stop(false)
		fmt.Printf("Error updating file: %v", err)
	}
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"
	labelMasterNode(errorDeployingCelleryRuntime)

	// Setup Cellery namespace
	spinner.SetNewAction("Setting up cellery name space")
	if err := CreateCelleryNameSpace(); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}

	// Apply Istio CRDs
	spinner.SetNewAction("Applying istio crds")
	if err := ApplyIstioCrds(artifactsPath); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}
	// sleep for few seconds - this is to make sure that the CRDs are properly applied
	time.Sleep(20 * time.Second)

	// Enabling Istio injection
	spinner.SetNewAction("Enabling istio injection")
	if err := kubectl.ApplyLable("namespace", "default", "istio-injection=enabled",
		true); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}

	// Install istio
	spinner.SetNewAction("Installing istio")
	if err := InstallIstio(filepath.Join(util.CelleryInstallationDir(), "k8s-artefacts")); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}

	// Apply controller CRDs
	spinner.SetNewAction("Creating controller")
	if err := InstallController(filepath.Join(util.CelleryInstallationDir(), "k8s-artefacts")); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}

	spinner.SetNewAction("Configuring mysql")
	AddMysql(artifactsPath, isPersistentVolume)

	CreateGlobalGatewayConfigMaps(artifactsPath)
	if isPersistentVolume {
		createPersistentVolume(artifactsPath, hasNfsStorage)
	}
	CreateObservabilityConfigMaps(artifactsPath)
	CreateIdpConfigMaps(artifactsPath)

	if isCompleteSetup {
		spinner.SetNewAction("Adding apim")
		addApim(artifactsPath, isPersistentVolume)
		spinner.SetNewAction("Adding observability")
		addObservability(artifactsPath)
	} else {
		spinner.SetNewAction("Adding idp")
		addIdp(artifactsPath)
	}
	spinner.SetNewAction("Creating ingress-nginx")
	installNginx(artifactsPath, isLoadBalancerIngressMode)
	spinner.Stop(true)

	return nil
}

func UpdateRuntime(apiManagement, observability bool) error {
	var err error
	err = DeleteComponent(Observability)
	if err != nil {
		return err
	}
	if apiManagement {
		err = DeleteComponent(IdentityProvider)
		if err != nil {
			return err
		}
		err = AddComponent(ApiManager)
		if err != nil {
			return err
		}
	} else {
		err = DeleteComponent(ApiManager)
		if err != nil {
			return err
		}
		err = AddComponent(IdentityProvider)
		if err != nil {
			return err
		}
	}

	if observability {
		err = AddComponent(Observability)
		if err != nil {
			return err
		}
	} else {
		err = DeleteComponent(Observability)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddComponent(component SystemComponent) error {
	switch component {
	case ApiManager:
		return addApim(filepath.Join(util.CelleryInstallationDir(), "k8s-artefacts"), false)
	case IdentityProvider:
		return addIdp(filepath.Join(util.CelleryInstallationDir(), "k8s-artefacts"))
	case Observability:
		return addObservability(filepath.Join(util.CelleryInstallationDir(), "k8s-artefacts"))

	default:
		return fmt.Errorf("unknown system componenet %q", component)
	}
}

func DeleteComponent(component SystemComponent) error {
	switch component {
	case ApiManager:
		return deleteApim(filepath.Join(util.CelleryInstallationDir(), "k8s-artefacts"))
	case IdentityProvider:
		return deleteIdp(filepath.Join(util.CelleryInstallationDir(), "k8s-artefacts"))
	case Observability:
		return deleteObservability(filepath.Join(util.CelleryInstallationDir(), "k8s-artefacts"))
	default:
		return fmt.Errorf("unknown system componenet %q", component)
	}
}

func labelMasterNode(errorDeployingCelleryRuntime string) {
	nodeName, err := kubectl.GetMasterNodeName()
	if err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}
	if err := kubectl.ApplyLable("nodes", nodeName, "disk=local", true); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
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
	default:
		return filepath.Join(artifactsPath)
	}
}

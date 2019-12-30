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

package runtime

import (
	"fmt"
	"path/filepath"
	"strings"

	"cellery.io/cellery/components/cli/pkg/util"

	"cellery.io/cellery/components/cli/pkg/kubernetes"
)

func (runtime *CelleryRuntime) AddApim(isPersistentVolume bool, nfs Nfs, db MysqlDb) error {
	//for _, v := range buildApimYamlPaths(runtime.artifactsPath, isPersistentVolume) {
	//	err := kubernetes.ApplyFileWithNamespace(v, "cellery-system")
	//	if err != nil {
	//		return err
	//	}
	//}
	runtime.UnmarshalHelmValues("cellery-runtime")
	runtime.celleryRuntimeVals.ApiManager.Enabled = true
	if runtime.IsGcpRuntime() {
		runtime.celleryRuntimeVals.Global.CelleryRuntime.Db.Hostname = db.DbHostName
		runtime.celleryRuntimeVals.Global.CelleryRuntime.Db.CarbonDb.Username = db.DbUserName
		runtime.celleryRuntimeVals.Global.CelleryRuntime.Db.CarbonDb.Password = db.DbPassword
	}
	if isPersistentVolume {
		runtime.celleryRuntimeVals.ApiManager.Persistence.Enabled = true
		runtime.celleryRuntimeVals.ApiManager.Persistence.Media = "local-filesystem"
		if nfs.NfsServerIp != "" {
			runtime.celleryRuntimeVals.ApiManager.Persistence.Media = "nfs"
			runtime.celleryRuntimeVals.ApiManager.Persistence.NfsServerIp = nfs.NfsServerIp
			runtime.celleryRuntimeVals.ApiManager.Persistence.SharedLocation = nfs.FileShare
		}
	} else {
		runtime.celleryRuntimeVals.ApiManager.Persistence.Enabled = false
		runtime.celleryRuntimeVals.ApiManager.Persistence.Media = "volatile"
	}
	runtime.MarshalHelmValues("cellery-runtime")
	if err := util.ApplyHelmChartWithCustomValues("cellery-runtime", "cellery-system",
		"apply", runtime.celleryRuntimeYaml); err != nil {
		return err
	}
	return nil
}

func deleteApim(artifactsPath string) error {
	for _, v := range buildApimYamlPaths(artifactsPath, false) {
		err := kubernetes.DeleteFileWithNamespace(v, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func (runtime *CelleryRuntime) DeleteApim() error {
	runtime.UnmarshalHelmValues("cellery-runtime")
	runtime.celleryRuntimeVals.ApiManager.Enabled = true
	runtime.MarshalHelmValues("cellery-runtime")
	if err := util.ApplyHelmChartWithCustomValues("cellery-runtime", "cellery-system",
		"delete", runtime.celleryRuntimeYaml); err != nil {
		return err
	}
	return nil
}

func IsApimEnabled() (bool, error) {
	enabled := true
	_, err := kubernetes.GetDeployment("cellery-system", "gateway")
	if err != nil {
		if strings.Contains(err.Error(), "No resources found") ||
			strings.Contains(err.Error(), "Error from server (NotFound)") {
			enabled = false
		} else {
			return enabled, fmt.Errorf("error checking if apim is enabled")
		}
	}
	return enabled, nil
}

func createGlobalGatewayConfigMaps(artifactsPath string) error {
	for _, confMap := range buildGlobalGatewayConfigMaps(artifactsPath) {
		err := kubernetes.CreateConfigMapWithNamespace(confMap.Name, confMap.Path, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func buildApimYamlPaths(artifactsPath string, isPersistentVolume bool) []string {
	base := buildArtifactsPath(ApiManager, artifactsPath)
	if isPersistentVolume {
		return []string{
			filepath.Join(base, "global-apim.yaml"),
		}
	}
	return []string{
		filepath.Join(base, "global-apim-volatile.yaml"),
	}
}

func buildGlobalGatewayConfigMaps(artifactsPath string) []ConfigMap {
	base := buildArtifactsPath(ApiManager, artifactsPath)
	return []ConfigMap{
		{"gw-conf", filepath.Join(base, "conf")},
		{"gw-conf-datasources", filepath.Join(base, "conf", "datasources")},
		{"conf-identity", filepath.Join(base, "conf", "identity")},
		{"apim-template", filepath.Join(base, "conf", "resources", "api_templates")},
		{"apim-tomcat", filepath.Join(base, "conf", "tomcat")},
		{"apim-security", filepath.Join(base, "conf", "security")},
	}
}

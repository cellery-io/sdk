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
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
)

func addApim(artifactsPath string, isPersistentVolume bool) error {
	for _, v := range buildApimYamlPaths(artifactsPath, isPersistentVolume) {
		err := kubectl.ApplyFileWithNamespace(v, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteApim(artifactsPath string) error {
	for _, v := range buildApimYamlPaths(artifactsPath, false) {
		err := kubectl.DeleteFileWithNamespace(v, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateGlobalGatewayConfigMaps(artifactsPath string) error {
	for _, confMap := range buildGlobalGatewayConfigMaps(artifactsPath) {
		err := kubectl.CreateConfigMapWithNamespace(confMap.Name, confMap.Path, "cellery-system")
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

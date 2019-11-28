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

package gcp

import (
	"path/filepath"

	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/runtime"
)

func CreateGlobalGatewayConfigMaps() error {
	for _, confMap := range buildGlobalGatewayConfigMaps() {
		err := kubernetes.CreateConfigMapWithNamespace(confMap.Name, confMap.Path, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func buildGlobalGatewayConfigMaps() []ConfigMap {
	base := buildArtifactsPath(runtime.ApiManager)
	return []ConfigMap{
		{"gw-conf", filepath.Join(base, "conf")},
		{"gw-conf-datasources", filepath.Join(base, "conf", "datasources")},
		{"conf-identity", filepath.Join(base, "conf", "identity")},
		{"apim-template", filepath.Join(base, "conf", "resources", "api_templates")},
		{"apim-tomcat", filepath.Join(base, "conf", "tomcat")},
		{"apim-security", filepath.Join(base, "conf", "security")},
	}
}

func AddApim() error {
	for _, v := range buildApimYamlPaths() {
		err := kubernetes.ApplyFileWithNamespace(v, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func buildApimYamlPaths() []string {
	base := buildArtifactsPath(runtime.ApiManager)
	return []string{
		filepath.Join(base, "global-apim.yaml"),
	}
}

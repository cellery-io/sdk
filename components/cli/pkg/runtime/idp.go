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

func addIdp(artifactsPath string) error {
	for _, v := range buildIdpYamlPaths(artifactsPath) {
		err := kubectl.ApplyFileWithNamespace(v, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteIdp(artifactsPath string) error {
	for _, v := range buildIdpYamlPaths(artifactsPath) {
		err := kubectl.DeleteFileWithNamespace(v, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateIdpConfigMaps(artifactsPath string) error {
	for _, confMap := range buildIdpConfigMaps(artifactsPath) {
		err := kubectl.CreateConfigMapWithNamespace(confMap.Name, confMap.Path, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func buildIdpYamlPaths(artifactsPath string) []string {
	base := buildArtifactsPath(IdentityProvider, artifactsPath)
	return []string{
		filepath.Join(base, "global-idp.yaml"),
	}
}

func buildIdpConfigMaps(artifactsPath string) []ConfigMap {
	base := buildArtifactsPath(IdentityProvider, artifactsPath)
	return []ConfigMap{
		{"identity-server-conf", filepath.Join(base, "conf")},
		{"identity-server-conf-datasources", filepath.Join(base, "conf", "datasources")},
		{"identity-server-conf-identity", filepath.Join(base, "conf", "identity")},
		{"identity-server-tomcat", filepath.Join(base, "conf", "tomcat")},
	}
}

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

	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/runtime"
)

func CreateIdpConfigMaps() error {
	for _, confMap := range buildIdpConfigMaps() {
		err := kubernetes.CreateConfigMapWithNamespace(confMap.Name, confMap.Path, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateIdp() error {
	for _, v := range buildIdpYamlPaths() {
		err := kubernetes.ApplyFileWithNamespace(v, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func buildIdpConfigMaps() []ConfigMap {
	base := buildArtifactsPath(runtime.IdentityProvider)
	return []ConfigMap{
		{"identity-server-conf", filepath.Join(base, "conf")},
		{"identity-server-conf-datasources", filepath.Join(base, "conf", "datasources")},
		{"identity-server-conf-identity", filepath.Join(base, "conf", "identity")},
		{"identity-server-tomcat", filepath.Join(base, "conf", "tomcat")},
	}
}

func buildIdpYamlPaths() []string {
	base := buildArtifactsPath(runtime.IdentityProvider)
	return []string{
		filepath.Join(base, "global-idp.yaml"),
	}
}

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
	"path/filepath"

	"cellery.io/cellery/components/cli/pkg/kubernetes"
)

func ApplyIstioCrds(artifactsPath string) error {
	for _, v := range buildIstioCrdsYamlPaths(artifactsPath) {
		err := kubernetes.ApplyFile(v)
		if err != nil {
			return err
		}
	}
	return nil

}

func InstallIstio(artifactsPath string) error {
	for _, v := range buildIstioYamlPaths(artifactsPath) {
		err := kubernetes.ApplyFile(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildIstioYamlPaths(artifactsPath string) []string {
	base := buildArtifactsPath(System, artifactsPath)
	return []string{
		filepath.Join(base, "istio-demo-cellery.yaml"),
		filepath.Join(base, "istio-gateway.yaml"),
	}
}

func buildIstioCrdsYamlPaths(artifactsPath string) []string {
	base := buildArtifactsPath(System, artifactsPath)
	return []string{
		filepath.Join(base, "istio-crds.yaml"),
	}
}

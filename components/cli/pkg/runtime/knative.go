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
	"time"

	"cellery.io/cellery/components/cli/pkg/kubernetes"
)

func InstallKnativeServing(artifactsPath string) error {
	for _, v := range buildKnativeYamlPaths(artifactsPath) {
		err := kubernetes.ApplyFile(v)
		if err != nil {
			time.Sleep(10 * time.Second)
			err = kubernetes.ApplyFile(v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ApplyKnativeCrds(artifactsPath string) error {
	for _, v := range buildKnativeCrdsYamlPaths(artifactsPath) {
		err := kubernetes.ApplyFile(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteKnative() error {
	out, err := kubernetes.DeleteResource("apiservices.apiregistration.k8s.io", "v1beta1.custom.metrics.k8s.io")
	if err != nil {
		return fmt.Errorf("error occurred while deleting the knative apiservice: %s", fmt.Errorf(out))
	}
	return kubernetes.DeleteNameSpace("knative-serving")
}

func IsKnativeEnabled() (bool, error) {
	enabled := true
	_, err := kubernetes.GetDeployment("knative-serving", "activator")
	if err != nil {
		if strings.Contains(err.Error(), "No resources found") ||
			strings.Contains(err.Error(), "not found") {
			enabled = false
		} else {
			return enabled, fmt.Errorf("error checking if knative serving is enabled")
		}
	}
	return enabled, nil
}

func buildKnativeYamlPaths(artifactsPath string) []string {
	base := buildArtifactsPath(System, artifactsPath)
	return []string{
		filepath.Join(base, "knative-serving.yaml"),
	}
}

func buildKnativeCrdsYamlPaths(artifactsPath string) []string {
	base := buildArtifactsPath(System, artifactsPath)
	return []string{
		filepath.Join(base, "knative-serving-crds.yaml"),
	}
}

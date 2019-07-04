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
 *
 */

package runtime

import (
	"fmt"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
)

func InstallHPA(artifactsPath string) error {
	for _, v := range buildHPAYamlPaths(artifactsPath) {
		err := kubectl.CreateFile(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteHpa(artifactsPath string) error {
	for _, v := range buildHPAYamlPaths(artifactsPath) {
		err := kubectl.DeleteFile(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func IsHpaEnabled() (bool, error) {
	enabled := true
	_, err := kubectl.GetDeployment("kube-system", "metrics-server")
	if err != nil {
		if strings.Contains(err.Error(), "No resources found") ||
			strings.Contains(err.Error(), "not found") {
			enabled = false
		} else {
			return enabled, fmt.Errorf("error checking if hpa is enabled")
		}
	}
	return enabled, nil
}

func buildHPAYamlPaths(artifactsPath string) []string {
	base := buildArtifactsPath(HPA, artifactsPath)
	return []string{
		base,
	}
}

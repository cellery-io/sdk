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
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/mattbaird/jsonpatch"

	"cellery.io/cellery/components/cli/pkg/kubernetes"
)

func (runtime *CelleryRuntime) InstallIngressNginx(isLoadBalancerIngressMode bool) error {
	for _, file := range buildNginxYamlPaths(runtime.artifactsPath, isLoadBalancerIngressMode) {
		err := kubernetes.ApplyFile(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildNginxYamlPaths(artifactsPath string, isLoadBalancerIngressMode bool) []string {
	base := buildArtifactsPath(System, artifactsPath)
	yamls := []string{filepath.Join(base, "mandatory.yaml")}
	if isLoadBalancerIngressMode {
		yamls = append(yamls, filepath.Join(base, "cloud-generic.yaml"))
	} else {
		yamls = append(yamls, filepath.Join(base, "service-nodeport.yaml"))
	}
	return yamls
}

func (runtime *CelleryRuntime) UpdateNodePortIpAddress(nodePortIpAddress string) error {
	originalIngressNginx, err := kubernetes.GetService("ingress-nginx", "ingress-nginx")
	if err != nil {
		return fmt.Errorf("error getting original ingress-nginx: %v", err)
	}
	updatedIngressNginx, err := kubernetes.GetService("ingress-nginx", "ingress-nginx")
	if err != nil {
		return fmt.Errorf("error getting updated ingress-nginx: %v", err)
	}
	updatedIngressNginx.Spec.ExternalIPs = append(updatedIngressNginx.Spec.ExternalIPs, nodePortIpAddress)

	originalData, err := json.Marshal(originalIngressNginx)
	if err != nil {
		return fmt.Errorf("error marshalling original data: %v", err)
	}
	desiredData, err := json.Marshal(updatedIngressNginx)
	if err != nil {
		return fmt.Errorf("error marshalling desired data: %v", err)
	}
	patch, err := jsonpatch.CreatePatch(originalData, desiredData)
	if err != nil {
		return fmt.Errorf("error creating json patch: %v", err)
	}
	if len(patch) == 0 {
		return nil
	}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("error marshalling json patch: %v", err)
	}
	kubernetes.JsonPatchWithNameSpace("svc", "ingress-nginx", string(patchBytes), "ingress-nginx")
	return nil
}

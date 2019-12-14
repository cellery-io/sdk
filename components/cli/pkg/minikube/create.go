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

package minikube

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/runtime"
)

const ciliumNetworkPlugin = "https://raw.githubusercontent.com/cilium/cilium/1.5.6/examples/kubernetes/1.15/cilium-minikube.yaml"

func (minikube *Minikube) CreateK8sCluster() (string, error) {
	cmd := exec.Command(
		minikubeCmd,
		"start",
		"--cpus", minikube.cpus,
		"--memory", minikube.memory,
		"--kubernetes-version", minikube.kubeVersion,
		"--profile", minikube.profile,
		"--network-plugin=cni",
		"--enable-default-cni",
	)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to start minikube, %v", err)
	}
	cmd = exec.Command(
		minikubeCmd,
		"ip",
		"--profile", celleryLocalSetup,
	)
	minikubeIp, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get minikube ip, %v", err)
	}
	kubernetes.CreateFile(ciliumNetworkPlugin)
	return strings.TrimSuffix(string(minikubeIp), "\n"), nil
}
func (minikube *Minikube) ConfigureSqlInstance() (runtime.MysqlDb, error) {
	return runtime.MysqlDb{}, nil
}
func (minikube *Minikube) CreateStorage() error {
	return nil
}
func (minikube *Minikube) CreateNfs() (runtime.Nfs, error) {
	return runtime.Nfs{}, nil
}
func (minikube *Minikube) UpdateKubeConfig() error {
	return nil
}
func (minikube *Minikube) ClusterName() string {
	return ""
}

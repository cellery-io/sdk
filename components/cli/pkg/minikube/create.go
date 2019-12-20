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
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/runtime"
)

const ciliumNetworkPlugin = "https://raw.githubusercontent.com/cilium/cilium/1.5.6/examples/kubernetes/1.15/cilium-minikube.yaml"

func (minikube *Minikube) CreateK8sCluster() (string, error) {
	var cmd *exec.Cmd
	if minikube.driver == "" {
		cmd = exec.Command(
			minikubeCmd,
			"start",
			"--cpus", minikube.cpus,
			"--memory", minikube.memory,
			"--kubernetes-version", minikube.kubeVersion,
			"--profile", minikube.profile,
			"--embed-certs=true",
			"--network-plugin=cni",
			"--enable-default-cni",
		)
	} else {
		cmd = exec.Command(
			minikubeCmd,
			"start",
			"--vm-driver", minikube.driver,
			"--cpus", minikube.cpus,
			"--memory", minikube.memory,
			"--kubernetes-version", minikube.kubeVersion,
			"--profile", minikube.profile,
			"--embed-certs=true",
			"--network-plugin=cni",
			"--enable-default-cni",
		)
	}
	var stderr bytes.Buffer
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Printf("\r\x1b[2K\033[36m%s\033[m\n", stdoutScanner.Text())
		}
	}()
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			fmt.Printf("\r\x1b[2K\033[36m%s\033[m\n", stderrScanner.Text())
			fmt.Fprintf(&stderr, stderrScanner.Text())
		}
	}()
	err := cmd.Start()
	if err != nil {
		errStr := string(stderr.Bytes())
		return "", fmt.Errorf("error occurred while starting minikube command, %v", errStr)
	}
	err = cmd.Wait()
	if err != nil {
		errStr := string(stderr.Bytes())
		return "", fmt.Errorf("error occurred while starting minikube, %v", errStr)
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

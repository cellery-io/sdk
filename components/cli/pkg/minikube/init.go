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
	"strconv"
)

const celleryLocalSetup = "cellery-local-setup"

type Minikube struct {
	cpus        string
	memory      string
	kubeVersion string
	profile     string
}

func NewMinikube(opts ...func(*Minikube)) (*Minikube, error) {
	minikube := &Minikube{}
	for _, opt := range opts {
		opt(minikube)
	}
	return minikube, nil
}

func SetCpus(cpus int) func(*Minikube) {
	return func(minikube *Minikube) {
		minikube.cpus = strconv.Itoa(cpus)
	}
}

func SetMemory(memory int) func(*Minikube) {
	return func(minikube *Minikube) {
		minikube.memory = strconv.Itoa(memory)
	}
}

func SetkubeVersion(kubeVersion string) func(*Minikube) {
	return func(minikube *Minikube) {
		minikube.kubeVersion = kubeVersion
	}
}

func SetProfile(profile string) func(*Minikube) {
	return func(minikube *Minikube) {
		minikube.profile = profile
	}
}

func ClusterExists(profile string) (bool, error) {
	cmd := exec.Command(
		"minikube",
		"status",
		"--profile", profile,
		"--format='{{.Host}}'",
	)
	var stderr bytes.Buffer
	var err error
	output := ""
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			output += stdoutScanner.Text()
		}
	}()
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			fmt.Fprintf(&stderr, stderrScanner.Text())
		}
	}()
	err = cmd.Start()
	if err != nil {
		errStr := string(stderr.Bytes())
		return false, fmt.Errorf("error occurred while starting to check minikube status, %v", errStr)
	}
	err = cmd.Wait()
	if err != nil {
		if output == "''" {
			return false, nil
		} else if output == "'Stopped'" {
			return true, nil
		} else {
			return false, fmt.Errorf("failed to check status of minikube profile %s, %v", profile, err)
		}
	}
	return true, nil
}

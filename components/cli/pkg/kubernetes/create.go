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

package kubernetes

import (
	"os"
	"os/exec"
)

func CreateFile(file string) error {
	cmd := exec.Command(
		kubectl,
		"create",
		"-f",
		file,
	)
	displayVerboseOutput(cmd)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CreateConfigMapWithNamespace(name, confFile, namespace string) error {
	cmd := exec.Command(
		kubectl,
		"create",
		"configmap",
		name,
		"--from-file",
		confFile,
		"-n", namespace,
	)
	displayVerboseOutput(cmd)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CreateClusterRoleBinding(clusterRole, user string) error {
	cmd := exec.Command(
		kubectl,
		"create",
		"clusterrolebinding",
		"cluster-admin-binding",
		"--clusterrole",
		clusterRole,
		"--user",
		user,
	)
	displayVerboseOutput(cmd)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (kubeCli *CelleryKubeCli) CreateNamespace(namespace string) error {
	cmd := exec.Command(
		kubectl,
		"create",
		"namespace",
		namespace,
	)
	displayVerboseOutput(cmd)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

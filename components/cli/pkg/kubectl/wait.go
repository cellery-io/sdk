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

package kubectl

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

func WaitForDeployment(condition string, timeoutSeconds int, resourceName string, namespace string) error {
	return WaitForCondition(condition, timeoutSeconds, fmt.Sprintf("deployment/%s", resourceName), namespace)
}

func WaitForCondition(condition string, timeoutSeconds int, resourceName string, namespace string) error {
	cmd := exec.Command(
		constants.KUBECTL,
		"wait",
		fmt.Sprintf("--for=condition=%s", condition),
		fmt.Sprintf("--timeout=%ds", timeoutSeconds),
		resourceName,
		"-n", namespace,
	)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func WaitForCluster(timeout time.Duration) error {
	exitCode := 0
	for start := time.Now(); time.Since(start) < timeout; {
		cmd := exec.Command(constants.KUBECTL, "get", "nodes", "--request-timeout=10s")
		err := cmd.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				ws := exitError.Sys().(syscall.WaitStatus)
				exitCode = ws.ExitStatus()
				if exitCode == 1 {
					continue
				} else {
					return fmt.Errorf("kubectl exit with unknown error code %d", exitCode)
				}
			} else {
				return err
			}
		} else {
			return nil
		}
	}
	return fmt.Errorf("cluster ready check timeout")
}

func WaitForDeployments(namespace string) error {
	names, err := GetDeploymentNames(namespace)
	if err != nil {
		return err
	}
	for _, v := range names {
		if len(v) == 0 {
			continue
		}
		err := WaitForDeployment("available", -1, v, namespace)
		if err != nil {
			return err
		}
	}
	return nil
}

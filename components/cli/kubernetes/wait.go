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

package kubernetes

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

func WaitForDeployment(condition string, timeoutSeconds int, resourceName string, namespace string) error {
	return WaitForCondition(condition, timeoutSeconds, fmt.Sprintf("deployment/%s", resourceName), namespace)
}

func WaitForCell(condition string, timeoutSeconds int, resourceName string, namespace string) error {
	return WaitForCondition(condition, timeoutSeconds, fmt.Sprintf("cells.mesh.cellery.io/%s", resourceName), namespace)
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
	displayVerboseOutput(cmd)
	return cmd.Run()
}

func WaitForCluster(timeout time.Duration) error {
	exitCode := 0
	for start := time.Now(); time.Since(start) < timeout; {
		cmd := exec.Command(constants.KUBECTL,
			"get",
			"nodes",
			"--request-timeout=10s",
		)
		displayVerboseOutput(cmd)
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

func WaitForDeployments(namespace string, timeout time.Duration) error {
	names, err := GetDeploymentNames(namespace)
	if err != nil {
		return err
	}
	start := time.Now()
	for _, v := range names {
		if len(v) == 0 {
			continue
		}
		// Check whether there is a timeout
		if time.Since(start) > timeout {
			return fmt.Errorf("deployment status check timeout. Please use 'kubectl get pods -n %s' to check the status", namespace)
		}

		for time.Since(start) < timeout {
			err := WaitForDeployment("available", int(timeout.Seconds()), v, namespace)
			if err != nil {
				continue
				// we ignore this error due to an unexpected error
				// ".status.conditions accessor error: Failure is of the type string, expected map[string]interface{}"
			}
			break
		}
	}
	return nil
}

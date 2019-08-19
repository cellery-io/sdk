/*
 * Copyright (c) 2018 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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

package commands

import (
	"fmt"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/runtime"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunTerminate(terminatingInstances []string, terminateAll bool) {
	runningInstances, err := runtime.GetInstancesNames()
	if err != nil {
		util.ExitWithErrorMessage("Error getting running cell instances", err)
	}
	// Delete all running instances
	if terminateAll {
		for _, runningInstance := range runningInstances {
			terminateInstance(runningInstance)
		}
	} else {
		// Check if any given instance is not running
		for _, terminatingInstance := range terminatingInstances {
			if util.ContainsInStringArray(runningInstances, terminatingInstance) {
				continue
			} else {
				util.ExitWithErrorMessage("Error terminating cell instances", fmt.Errorf("instance: %s does not exist", terminatingInstance))
			}
		}
		// If all given instances are running terminate them all
		for _, terminatingInstance := range terminatingInstances {
			terminateInstance(terminatingInstance)
		}
	}
}

func terminateInstance(instance string) {
	output, err := kubectl.DeleteResource("cell", instance)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while stopping the cell instance: "+instance, fmt.Errorf(output))
	}

	output, err = kubectl.DeleteResource("composite", instance)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while stopping the composite instance: "+instance, fmt.Errorf(output))
	}
	// Delete the TLS Secret
	secretName := instance + "--tls-secret"
	output, err = kubectl.DeleteResource("secret", secretName)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while deleting the secret: "+secretName, fmt.Errorf(output))
	}
}

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

package instance

import (
	"fmt"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/util"
)

func RunTerminate(cli cli.Cli, terminatingInstances []string, terminateAll bool) error {
	var err error
	var runningInstances []string
	if runningInstances, err = cli.KubeCli().GetInstancesNames(); err != nil {
		return fmt.Errorf("error getting running cell instances, %v", err)
	}
	if terminateAll {
		// Terminate all running instances
		for _, runningInstance := range runningInstances {
			terminateInstance(cli, runningInstance)
		}
	} else {
		// Check if any given instance is not running
		for _, terminatingInstance := range terminatingInstances {
			if util.ContainsInStringArray(runningInstances, terminatingInstance) {
				continue
			} else {
				return fmt.Errorf("error terminating cell instances, %v", fmt.Errorf("instance: %s does "+
					"not exist", terminatingInstance))
			}
		}
		// If all given instances are running terminate them all
		for _, terminatingInstance := range terminatingInstances {
			terminateInstance(cli, terminatingInstance)
		}
	}
	return nil
}

func terminateInstance(cli cli.Cli, instance string) error {
	var err error
	var output string
	if output, err = cli.KubeCli().DeleteResource("cell", instance); err != nil {
		return fmt.Errorf("error occurred while stopping the cell instance %s, %v", instance, fmt.Errorf(output))
	}
	if output, err = cli.KubeCli().DeleteResource("composite", instance); err != nil {
		return fmt.Errorf("error occurred while stopping the composite instance %s, %v", instance, fmt.Errorf(output))
	}
	// Delete the TLS Secret
	secretName := instance + "--tls-secret"
	if output, err = cli.KubeCli().DeleteResource("secret", secretName); err != nil {
		return fmt.Errorf("error occurred while deleting the secret: %s, %v", secretName, fmt.Errorf(output))
	}
	return nil
}

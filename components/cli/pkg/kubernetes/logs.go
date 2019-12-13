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
	"fmt"
	"os/exec"

	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/osexec"
)

func (kubeCli *CelleryKubeCli) StreamCellLogsUserComponents(instanceName string, follow bool) error {
	cmd := exec.Command(kubectl,
		"logs",
		"-l",
		constants.GroupName+"/cell="+instanceName+","+constants.GroupName+"/component",
		"--all-containers=true",
	)
	if follow {
		noOfContainers, err := getContainerCount(instanceName, false)
		if err != nil {
			return err
		}
		cmd.Args = append(cmd.Args, "-f", fmt.Sprintf("--max-log-requests=%d", noOfContainers))
	}
	displayVerboseOutput(cmd)
	return osexec.PrintCommandOutput(cmd)
}

func (kubeCli *CelleryKubeCli) StreamCellLogsAllComponents(instanceName string, follow bool) error {
	cmd := exec.Command(kubectl,
		"logs",
		"-l",
		constants.GroupName+"/cell="+instanceName,
		"--all-containers=true",
	)
	if follow {
		noOfContainers, err := getContainerCount(instanceName, true)
		if err != nil {
			return err
		}
		cmd.Args = append(cmd.Args, "-f", fmt.Sprintf("--max-log-requests=%d", noOfContainers))
	}
	displayVerboseOutput(cmd)
	return osexec.PrintCommandOutput(cmd)
}

func (kubeCli *CelleryKubeCli) StreamComponentLogs(instanceName string, componentName string, follow bool) error {
	cmd := exec.Command(kubectl,
		"logs",
		"-l",
		constants.GroupName+"/cell="+instanceName+","+constants.GroupName+"/component="+instanceName+"--"+componentName,
		"-c",
		componentName,
	)
	if follow {
		cmd.Args = append(cmd.Args, "-f")
	}
	displayVerboseOutput(cmd)
	return osexec.PrintCommandOutput(cmd)
}

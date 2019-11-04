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
)

func RunLogs(cellName, componentName string, sysLog bool) error {
	if componentName == "" {
		var logs string
		var err error
		if sysLog {
			logs, err = kubectl.GetCellLogsAllComponents(cellName)
		} else {
			logs, err = kubectl.GetCellLogsUserComponents(cellName)
		}

		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Error getting logs for instance %s", cellName), err)
		}
		if logs == "" {
			return fmt.Errorf(fmt.Sprintf("No logs found"), fmt.Errorf("cannot find cell "+
				"instance %s", cellName))
		}
	} else {
		logs, err := kubectl.GetComponentLogs(cellName, componentName)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Error getting logs for component %s of instance %s",
				componentName, cellName), err)
		}
		if logs == "" {
			return fmt.Errorf(fmt.Sprintf("No logs found"), fmt.Errorf("cannot find component "+
				"%s of cell instance %s", componentName, cellName))
		}
	}
	return nil
}

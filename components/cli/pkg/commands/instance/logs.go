/*
 * Copyright (c) 2018 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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
)

func RunLogs(cli cli.Cli, cellName, componentName string, sysLog bool, follow bool) error {
	if componentName == "" {
		var logsFound bool
		var err error
		if sysLog {
			logsFound, err = cli.KubeCli().GetCellLogsAllComponents(cellName, follow)
		} else {
			logsFound, err = cli.KubeCli().GetCellLogsUserComponents(cellName, follow)
		}

		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Error getting logs for instance %s", cellName), err)
		}
		if !logsFound {
			return fmt.Errorf(fmt.Sprintf("No logs found"), fmt.Errorf("cannot find cell "+
				"instance %s", cellName))
		}
	} else {
		logsFound, err := cli.KubeCli().GetComponentLogs(cellName, componentName, follow)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Error getting logs for component %s of instance %s",
				componentName, cellName), err)
		}
		if !logsFound {
			return fmt.Errorf(fmt.Sprintf("No logs found"), fmt.Errorf("cannot find component "+
				"%s of cell instance %s", componentName, cellName))
		}
	}
	return nil
}

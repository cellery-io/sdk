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

	"github.com/hashicorp/go-version"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/util"
)

const minimumKubernetesVersionToFollowLogs = "v1.14.0"

func RunLogs(cli cli.Cli, instanceName string, componentName string, sysLog bool, follow bool) error {
	if err := cli.KubeCli().IsInstanceAvailable(instanceName); err != nil {
		return fmt.Errorf(fmt.Sprintf("No logs found"), fmt.Errorf("cannot find running "+
			"instance %s", instanceName))
	}

	if follow {
		_, clientVersion, err := cli.KubeCli().Version()
		if err != nil {
			return err
		}
		kubectlClientVersion, err := version.NewVersion(clientVersion)
		minimumKubernetesClientVersion, err := version.NewVersion(minimumKubernetesVersionToFollowLogs)

		if kubectlClientVersion.LessThan(minimumKubernetesClientVersion) {
			util.PrintWarningMessage(fmt.Sprintf("Unable to follow cellery logs for instance %s."+
				"Please upgrade kubernetes client version to 1.14 or higher", instanceName))
			return nil
		}
	}

	if componentName == "" {
		var err error
		if sysLog {
			err = cli.KubeCli().StreamCellLogsAllComponents(instanceName, follow)
		} else {
			err = cli.KubeCli().StreamCellLogsUserComponents(instanceName, follow)
		}

		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Error getting logs for instance %s", instanceName), err)
		}
	} else {
		if err := cli.KubeCli().IsComponentAvailable(instanceName, componentName); err != nil {
			return fmt.Errorf(fmt.Sprintf("No logs found"), fmt.Errorf("cannot find component "+
				"%s of cell instance %s", componentName, instanceName))
		}
		if err := cli.KubeCli().StreamComponentLogs(instanceName, componentName, follow); err != nil {
			return fmt.Errorf(fmt.Sprintf("Error getting logs for component %s of instance %s",
				componentName, instanceName), err)
		}
	}
	return nil
}

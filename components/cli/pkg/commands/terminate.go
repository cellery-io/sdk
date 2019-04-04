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
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunTerminate(instanceName string) {
	// Delete the Cell
	output, err := util.ExecuteKubeCtlCmd(constants.DELETE, "cell", instanceName)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while stopping the cell instance: "+instanceName, fmt.Errorf(output))
	}

	// Delete the TLS Secret
	secretName := instanceName + "--tls-secret"
	output, err = util.ExecuteKubeCtlCmd(constants.DELETE, "secret", secretName, constants.IGNORE_NOT_FOUND)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while deleting the secret: "+secretName, fmt.Errorf(output))
	}
}

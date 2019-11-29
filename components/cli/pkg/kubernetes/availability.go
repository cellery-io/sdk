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
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/osexec"
	"fmt"
	"os/exec"
	"strings"

	errorpkg "cellery.io/cellery/components/cli/pkg/error"
)

func (kubeCli *CelleryKubeCli) IsInstanceAvailable(instanceName string) error {
	var canBeComposite bool
	_, err := kubeCli.GetCell(instanceName)
	if err != nil {
		if cellNotFound, _ := errorpkg.IsCellInstanceNotFoundError(instanceName, err); cellNotFound {
			canBeComposite = true
		} else {
			return fmt.Errorf("failed to check available Cells, %v", err)
		}
	} else {
		return nil
	}
	if canBeComposite {
		_, err := kubeCli.GetComposite(instanceName)
		if err != nil {
			if compositeNotFound, _ := errorpkg.IsCompositeInstanceNotFoundError(instanceName, err); compositeNotFound {
				return fmt.Errorf("instance %s not available in the runtime", instanceName)
			} else {
				return fmt.Errorf("failed to check available Composites, %v", err)
			}
		} else {
			return nil
		}
	}
	return nil
}

func (kubeCli *CelleryKubeCli) IsComponentAvailable(instanceName, componentName string) error {
	cmd := exec.Command(constants.KubeCtl,
		"get",
		"component",
		instanceName+"--"+componentName,
	)
	displayVerboseOutput(cmd)
	_, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("component %s not found", componentName)
		}
		return fmt.Errorf("unknown error: %v", err)
	}
	return nil
}

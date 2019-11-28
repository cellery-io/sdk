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

	errorpkg "github.com/cellery-io/sdk/components/cli/pkg/error"
)

func GetInstanceAvailability(instanceName string) (bool, error) {
	var canBeComposite bool
	_, err := GetCell(instanceName)
	if err != nil {
		if cellNotFound, _ := errorpkg.IsCellInstanceNotFoundError(instanceName, err); cellNotFound {
			canBeComposite = true
		} else {
			return false, fmt.Errorf("failed to check available Cells, %v", err)
		}
	} else {
		return true, nil
	}
	if canBeComposite {
		_, err := GetComposite(instanceName)
		if err != nil {
			if compositeNotFound, _ := errorpkg.IsCompositeInstanceNotFoundError(instanceName, err); compositeNotFound {
				return false, fmt.Errorf("failed to retrieve dependencies of %s, instance not available in the runtime", instanceName)
			} else {
				return false, fmt.Errorf("failed to check available Composites, %v", err)
			}
		} else {
			return true, nil
		}
	}
	return false, nil
}

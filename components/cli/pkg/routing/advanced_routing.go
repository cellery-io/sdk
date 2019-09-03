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

package routing

import (
	"fmt"
	"regexp"
)

func isCellInstanceNotFoundError(srcInst string, cellErr error) (bool, error) {
	matches, err := regexp.MatchString(buildCellInstanceNonExistErrorMatcher(srcInst),
		cellErr.Error())
	if err != nil {
		return false, err
	}
	return matches, nil
}

func getCellGatewayHost(instance string) string {
	return fmt.Sprintf("%s--gateway-service", instance)
}

func getGatewayName(instance string) string {
	return fmt.Sprintf("%s--gateway", instance)
}

func getVsName(instance string) string {
	return fmt.Sprintf("%s--vs", instance)
}

func getCompositeServiceHost(instance string, component string) string {
	return fmt.Sprintf("%s--%s-service", instance, component)
}

func buildCellInstanceNonExistErrorMatcher(name string) string {
	return fmt.Sprintf("cell(.)+(%s)(.)+not found", name)
}

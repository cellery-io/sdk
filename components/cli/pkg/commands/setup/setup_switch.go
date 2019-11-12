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

package setup

import (
	"fmt"

	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunSwitchCommand(context string) error {
	if err := kubernetes.UseContext(context); err != nil {
		util.ExitWithErrorMessage("Failed to switch cluster", err)
	}
	return nil
}

func ValidateCluster(cluster string) error {
	if !util.ContainsInStringArray(getContexts(), cluster) {
		return fmt.Errorf("cluster %s doesn't exist", cluster)
	}
	return nil
}

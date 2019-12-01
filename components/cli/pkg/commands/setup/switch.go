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

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/util"
)

func RunSetupSwitch(cli cli.Cli, context string) error {
	if err := cli.KubeCli().UseContext(context); err != nil {
		return fmt.Errorf("failed to switch cluster, %v", err)
	}
	return nil
}

func ValidateCluster(cli cli.Cli, cluster string) error {
	contexts, err := getContexts(cli)
	if err != nil {
		return fmt.Errorf("failed to get contexts, %v", err)
	}
	if !util.ContainsInStringArray(contexts, cluster) {
		return fmt.Errorf("cluster %s doesn't exist", cluster)
	}
	return nil
}

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
	"encoding/json"
	"fmt"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
)

func RunSetupListClusters(cli cli.Cli) error {
	contexts, err := getContexts(cli)
	if err != nil {
		return fmt.Errorf("failed to get contexts, %v", err)
	}
	for _, v := range contexts {
		fmt.Fprintln(cli.Out(), v)
	}
	return nil
}

func getContexts(cli cli.Cli) ([]string, error) {
	var contexts []string
	jsonOutput := &kubernetes.Config{}
	output, err := cli.KubeCli().GetContexts()
	if err != nil {
		return nil, fmt.Errorf("error getting context list, %v", err)
	}
	err = json.Unmarshal(output, jsonOutput)
	if err != nil {
		return nil, fmt.Errorf("error trying to unmarshal contexts output, %v", err)
	}
	for i := 0; i < len(jsonOutput.Contexts); i++ {
		contexts = append(contexts, jsonOutput.Contexts[i].Name)
	}
	return contexts, nil
}

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

package commands

import (
	"encoding/json"
	"fmt"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunSetupListClusters() error {
	for _, v := range getContexts() {
		fmt.Println(v)
	}
	return nil
}

func getContexts() []string {
	var contexts []string
	jsonOutput := &kubectl.Config{}
	output, err := kubectl.GetContexts()
	if err != nil {
		util.ExitWithErrorMessage("Error getting context list", err)
	}
	err = json.Unmarshal(output, jsonOutput)
	if err != nil {
		util.ExitWithErrorMessage("Error trying to unmarshal contexts output", err)
	}
	for i := 0; i < len(jsonOutput.Contexts); i++ {
		contexts = append(contexts, jsonOutput.Contexts[i].Name)
	}
	return contexts
}

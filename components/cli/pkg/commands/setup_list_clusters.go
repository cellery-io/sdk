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
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/cellery-io/sdk/components/cli/pkg/util"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

func RunSetupListClusters() error {
	for _, v := range getContexts() {
		fmt.Println(v)
	}
	return nil
}

func getContexts() []string {
	contexts := []string{}
	cmd := exec.Command(constants.KUBECTL, "config", "view", "-o", "json")
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	output := ""
	go func() {
		for stdoutScanner.Scan() {
			output = output + stdoutScanner.Text()
		}
	}()
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)

	execError := ""
	go func() {
		for stderrScanner.Scan() {
			execError += stderrScanner.Text()
		}
	}()
	err := cmd.Start()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option", err)
	}
	err = cmd.Wait()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while configuring cellery", err)
	}
	jsonOutput := &Config{}
	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson != nil {
		fmt.Println(errJson)
	}

	for i := 0; i < len(jsonOutput.Contexts); i++ {
		contexts = append(contexts, jsonOutput.Contexts[i].Name)
	}
	return contexts
}

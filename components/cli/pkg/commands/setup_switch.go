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
	"fmt"
	"os"
	"os/exec"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunSwitchCommand(context string) error {
	cmd := exec.Command(constants.KUBECTL, "config", "use-context", context)
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
		fmt.Printf("Error in executing cellery setup: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while configuring cellery: \x1b[0m %v \n", execError)
		os.Exit(1)
	}
	return nil
}

func ValidateCluster(cluster string) error {
	if !util.ContainsInStringArray(getContexts(), cluster) {
		return fmt.Errorf("cluster %s doesn't exist", cluster)
	}
	return nil
}

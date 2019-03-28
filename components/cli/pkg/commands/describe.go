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
	"bufio"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunDescribe(cellImage string) {
	instancePattern, _ := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), cellImage)
	if instancePattern {
		cmd := exec.Command("kubectl", "describe", "cells", cellImage)
		stdoutReader, _ := cmd.StdoutPipe()
		stdoutScanner := bufio.NewScanner(stdoutReader)
		go func() {
			for stdoutScanner.Scan() {
				fmt.Println(stdoutScanner.Text())
			}
		}()
		stderrReader, _ := cmd.StderrPipe()
		stderrScanner := bufio.NewScanner(stderrReader)
		go func() {
			for stderrScanner.Scan() {
				fmt.Println(stderrScanner.Text())
			}
		}()
		err := cmd.Start()
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while fetching cell details", err)
		}
		err = cmd.Wait()
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while fetching cell details", err)
		}
	} else {
		fmt.Println(string(util.ReadCellImageYaml(cellImage)))
	}
}

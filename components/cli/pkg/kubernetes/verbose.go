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
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

var verboseMode = false
var verboseColor = color.New(color.FgWhite).Add(color.Bold).SprintFunc()

func getCommandString(cmd *exec.Cmd) string {
	var verboseModePrefix = ">>"
	var commandArgs []string
	commandArgs = append(commandArgs, verboseModePrefix)
	commandArgs = append(commandArgs, cmd.Args...)
	return strings.Join(commandArgs, " ")
}

func displayVerboseOutput(cmd *exec.Cmd) {
	// If running on verbose mode expose the kubectl commands.
	if verboseMode {
		fmt.Println(verboseColor(getCommandString(cmd)))
	}
}

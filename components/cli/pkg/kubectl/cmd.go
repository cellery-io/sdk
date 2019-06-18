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
 *
 */

package kubectl

import (
	"bufio"
	"os/exec"
	"strings"
)

func getCommandString(cmd *exec.Cmd) string {
	var verboseModePrefix = ">>"
	var commandArgs []string
	commandArgs = append(commandArgs, verboseModePrefix)
	commandArgs = append(commandArgs, cmd.Args...)
	return strings.Join(commandArgs, " ")
}

func getCommandOutput(cmd *exec.Cmd) (string, error) {
	var output string
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			output += stdoutScanner.Text()
		}
	}()
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			output += stderrScanner.Text()
		}
	}()
	err := cmd.Start()
	if err != nil {
		return output, err
	}
	err = cmd.Wait()
	if err != nil {
		return output, err
	}
	return output, nil
}

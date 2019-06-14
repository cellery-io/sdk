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

package kubectl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

func GetDeploymentNames(namespace string) ([]string, error) {
	cmd := exec.Command(
		constants.KUBECTL,
		"get",
		"deployments",
		"-o",
		"jsonpath={.items[*].metadata.name}",
		"-n", namespace,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(out), " "), nil
}

func GetMasterNodeName() (string, error) {
	cmd := exec.Command(constants.KUBECTL, "get", "node", "--selector", "node-role.kubernetes.io/master",
		"-o", "json")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	jsonOutput := &Node{}
	err = json.Unmarshal([]byte(out), jsonOutput)
	if err != nil {
		return "", err
	}
	if len(jsonOutput.Items) > 0 {
		return jsonOutput.Items[0].Metadata.Name, nil
	}
	return "", fmt.Errorf("node with master role does not exist")
}

func GetNodes() (Node, error) {
	cmd := exec.Command(
		constants.KUBECTL,
		"get",
		"nodes",
		"-o",
		"json",
	)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	jsonOutput := Node{}
	if err != nil {
		return jsonOutput, err
	}
	errJson := json.Unmarshal([]byte(out), &jsonOutput)
	if errJson != nil {
		return jsonOutput, errJson
	}
	return jsonOutput, nil
}

func GetCells(verboseMode bool) (Cells, error) {
	var Red = color.New(color.FgWhite).Add(color.Bold).SprintFunc()
	cmd := exec.Command(
		constants.KUBECTL,
		"get",
		"cells",
		"-o",
		"json",
	)
	// If running on verbose mode expose the kubectl commands.
	if verboseMode {
		fmt.Println(Red(getCommandString(cmd)))
		fmt.Println()
	}
	jsonOutput := Cells{}
	outfile, err := os.Create("./out.txt")
	if err != nil {
		return jsonOutput, fmt.Errorf("unable to create file: %v", err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()
	err = cmd.Start()
	if err != nil {
		return jsonOutput, fmt.Errorf("error getting cell data: %v", err)
	}
	err = cmd.Wait()
	if err != nil {
		return jsonOutput, fmt.Errorf("error waiting to get cell data: %v", err)
	}
	out, err := ioutil.ReadFile("./out.txt")
	os.Remove("./out.txt")
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput, err
}

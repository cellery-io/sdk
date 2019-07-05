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
	"encoding/json"
	"fmt"
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
	displayVerboseOutput(cmd)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(out), " "), nil
}

func GetMasterNodeName() (string, error) {
	cmd := exec.Command(constants.KUBECTL,
		"get",
		"node",
		"--selector",
		"node-role.kubernetes.io/master",
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
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
	displayVerboseOutput(cmd)
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

func GetCells() (Cells, error) {
	cmd := exec.Command(
		constants.KUBECTL,
		"get",
		"cells",
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := Cells{}
	out, err := getCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput, err
}

func GetCell(cellName string) (Cell, error) {
	cmd := exec.Command(constants.KUBECTL,
		"get",
		"cells",
		cellName,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	out, err := getCommandOutput(cmd)
	jsonOutput := Cell{}
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return jsonOutput, fmt.Errorf("cell instance %s not found", cellName)
		}
		return jsonOutput, fmt.Errorf("unknown error: %v", err)
	}
	err = json.Unmarshal([]byte(out), &jsonOutput)
	if err != nil {
		return jsonOutput, err
	}
	return jsonOutput, err
}

func GetPods(cellName string) (Pods, error) {
	cmd := exec.Command(constants.KUBECTL,
		"get",
		"pods",
		"-l",
		constants.GROUP_NAME+"/cell="+cellName,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := Pods{}
	out, err := getCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput, err
}

func GetServices(cellName string) (Services, error) {
	cmd := exec.Command(
		constants.KUBECTL,
		"get",
		"services",
		"-l",
		constants.GROUP_NAME+"/cell="+cellName,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := Services{}
	out, err := getCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput, err
}

func GetGateways(cellName string) (Gateway, error) {
	cmd := exec.Command(constants.KUBECTL,
		"get",
		"gateways",
		cellName+"--gateway",
		"-o", ""+
			"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := Gateway{}
	out, err := getCommandOutput(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal([]byte(out), &jsonOutput)
	return jsonOutput, err
}

func GetVirtualService(vs string) (VirtualService, error) {
	cmd := exec.Command(constants.KUBECTL,
		"get",
		"virtualservice",
		vs,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := VirtualService{}
	out, err := getCommandOutput(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal([]byte(out), &jsonOutput)
	return jsonOutput, err
}

func GetAutoscalePolicy(autoscalepolicy string) (*AutoscalePolicy, error) {
	cmd := exec.Command(constants.KUBECTL,
		"get",
		"ap",
		autoscalepolicy,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := &AutoscalePolicy{}
	out, err := getCommandOutput(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal([]byte(out), &jsonOutput)
	return jsonOutput, err
}

func GetDeployment(namespace, deployment string) (string, error) {
	cmd := exec.Command(
		constants.KUBECTL,
		"get",
		"deployments",
		deployment,
		"-n", namespace,
	)
	displayVerboseOutput(cmd)
	out, err := getCommandOutput(cmd)
	return out, err
}

func GetGatewayAsMapInterface(gw string) (map[string]interface{}, error) {
	cmd := exec.Command(constants.KUBECTL,
		"get",
		"gateway",
		gw,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	var output map[string]interface{}
	out, err := getCommandOutput(cmd)
	if err != nil {
		return output, err
	}
	err = json.Unmarshal([]byte(out), &output)
	return output, err
}

func GetCellInstanceAsMapInterface(cell string) (map[string]interface{}, error) {
	cmd := exec.Command(constants.KUBECTL,
		"get",
		"cell",
		cell,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	var output map[string]interface{}
	out, err := getCommandOutput(cmd)
	if err != nil {
		return output, err
	}
	err = json.Unmarshal([]byte(out), &output)
	return output, err
}

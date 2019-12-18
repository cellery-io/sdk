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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/osexec"
)

func GetDeploymentNames(namespace string) ([]string, error) {
	cmd := exec.Command(
		kubectl,
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

func (kubeCli *CelleryKubeCli) GetMasterNodeName() (string, error) {
	cmd := exec.Command(kubectl,
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
		kubectl,
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

// GetCells returns mock cell instances array.
func (kubeCli *CelleryKubeCli) GetCells() ([]Cell, error) {
	cmd := exec.Command(
		kubectl,
		"get",
		"cells",
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := Cells{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput.Items, err
	}
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput.Items, err
}

func (kubeCli *CelleryKubeCli) GetComposites() ([]Composite, error) {
	cmd := exec.Command(
		kubectl,
		"get",
		"composites",
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := Composites{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput.Items, err
	}
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput.Items, err
}

func (kubeCli *CelleryKubeCli) GetCell(cellName string) (Cell, error) {
	cmd := exec.Command(kubectl,
		"get",
		"cells",
		cellName,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	jsonOutput := Cell{}
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return jsonOutput, fmt.Errorf("cell instance %s not found", cellName)
		}
		return jsonOutput, fmt.Errorf("unknown error: %v", err)
	}
	err = json.Unmarshal(out, &jsonOutput)
	if err != nil {
		return jsonOutput, err
	}
	return jsonOutput, err
}

func (kubeCli *CelleryKubeCli) GetComposite(compositeName string) (Composite, error) {
	cmd := exec.Command(kubectl,
		"get",
		"composite",
		compositeName,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	jsonOutput := Composite{}
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return jsonOutput, fmt.Errorf("composite instance %s not found", compositeName)
		}
		return jsonOutput, fmt.Errorf("unknown error: %v", err)
	}
	err = json.Unmarshal(out, &jsonOutput)
	if err != nil {
		return jsonOutput, err
	}
	return jsonOutput, err
}

func (kubeCli *CelleryKubeCli) GetPodsForCell(cellName string) (Pods, error) {
	cmd := exec.Command(kubectl,
		"get",
		"pods",
		"-l",
		constants.GroupName+"/cell="+cellName,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := Pods{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput, err
}

func (kubeCli *CelleryKubeCli) GetPodsForComposite(compName string) (Pods, error) {
	cmd := exec.Command(kubectl,
		"get",
		"pods",
		"-l",
		constants.GroupName+"/composite="+compName,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := Pods{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput, err
}

func (kubeCli *CelleryKubeCli) GetServices(cellName string) (Services, error) {
	cmd := exec.Command(
		kubectl,
		"get",
		"services",
		"-l",
		constants.GroupName+"/cell="+cellName,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := Services{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput, err
}

func (kubeCli *CelleryKubeCli) GetVirtualService(vs string) (VirtualService, error) {
	cmd := exec.Command(kubectl,
		"get",
		"virtualservice",
		vs,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := VirtualService{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal([]byte(out), &jsonOutput)
	return jsonOutput, err
}

func GetDeployment(namespace, deployment string) (string, error) {
	cmd := exec.Command(
		kubectl,
		"get",
		"deployments",
		deployment,
		"-n", namespace,
	)
	displayVerboseOutput(cmd)
	out, err := osexec.GetCommandOutput(cmd)
	return out, err
}

func GetGatewayAsMapInterface(gw string) (map[string]interface{}, error) {
	cmd := exec.Command(kubectl,
		"get",
		"gateways",
		gw,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	var output map[string]interface{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return output, err
	}
	err = json.Unmarshal([]byte(out), &output)
	return output, err
}

func (kubeCli *CelleryKubeCli) GetCellInstanceAsMapInterface(cell string) (map[string]interface{}, error) {
	cmd := exec.Command(kubectl,
		"get",
		"cell",
		cell,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	var output map[string]interface{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return output, err
	}
	err = json.Unmarshal([]byte(out), &output)
	return output, err
}

func (kubeCli *CelleryKubeCli) GetCompositeInstanceAsMapInterface(composite string) (map[string]interface{}, error) {
	cmd := exec.Command(kubectl,
		"get",
		"composite",
		composite,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	var output map[string]interface{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return output, err
	}
	err = json.Unmarshal([]byte(out), &output)
	return output, err
}

func GetService(service, namespace string) (Service, error) {
	cmd := exec.Command(
		kubectl,
		"get",
		"services",
		"-n",
		namespace,
		service,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	jsonOutput := Service{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput, err
	}
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput, err
}

func getContainerCount(cellName string, considerSystemPods bool) (int, error) {
	cmd := exec.Command(
		kubectl,
		"get",
		"pods",
		"-o",
		"jsonpath={.items[*].spec['containers','initContainers'][*].name}",
	)
	if considerSystemPods {
		cmd.Args = append(cmd.Args, "-l", constants.GroupName+"/cell="+cellName)
	} else {
		cmd.Args = append(cmd.Args, "-l",
			constants.GroupName+"/cell="+cellName+","+constants.GroupName+"/component")
	}
	out, err := cmd.Output()
	displayVerboseOutput(cmd)
	if err != nil {
		return 0, err
	}
	return len(strings.Split(string(out), " ")), nil
}

func (kubeCli *CelleryKubeCli) GetInstancesNames() ([]string, error) {
	var instances []string
	runningCellInstances, err := kubeCli.GetCells()
	if err != nil {
		return nil, err
	}
	runningCompositeInstances, err := kubeCli.GetComposites()
	if err != nil {
		return nil, err
	}
	for _, runningInstance := range runningCellInstances {
		instances = append(instances, runningInstance.CellMetaData.Name)
	}
	for _, runningInstance := range runningCompositeInstances {
		instances = append(instances, runningInstance.CompositeMetaData.Name)
	}
	return instances, nil
}

func (kubeCli *CelleryKubeCli) GetInstanceBytes(instanceKind, InstanceName string) ([]byte, error) {
	cmd := exec.Command(kubectl,
		"get",
		instanceKind,
		InstanceName,
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	return []byte(out), err
}

func (kubeCli *CelleryKubeCli) GetNamespace(namespace string) ([]byte, error) {
	cmd := exec.Command(
		kubectl,
		"get",
		"namespace",
		namespace,
	)
	displayVerboseOutput(cmd)
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return nil, err
	}
	return out, nil
}

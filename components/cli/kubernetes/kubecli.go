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

package kubernetes

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/osexec"
)

// KubeCli represents kubernetes client.
type KubeCli interface {
	GetCells() ([]Cell, error)
	SetVerboseMode(enable bool)
	DeleteResource(kind, instance string) (string, error)
	GetComposites() ([]Composite, error)
	GetInstancesNames() ([]string, error)
	GetCell(cellName string) (Cell, error)
	GetComposite(compositeName string) (Composite, error)
}

type CelleryKubeCli struct {
}

// NewCelleryCli returns a CelleryCli instance.
func NewCelleryKubeCli() *CelleryKubeCli {
	kubeCli := &CelleryKubeCli{}
	return kubeCli
}

// GetCells returns mock cell instances array.
func (kubecli *CelleryKubeCli) GetCells() ([]Cell, error) {
	cmd := exec.Command(
		constants.KUBECTL,
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

func (kubecli *CelleryKubeCli) GetComposites() ([]Composite, error) {
	cmd := exec.Command(
		constants.KUBECTL,
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

func (kubecli *CelleryKubeCli) GetCell(cellName string) (Cell, error) {
	cmd := exec.Command(constants.KUBECTL,
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

func (kubecli *CelleryKubeCli) GetComposite(compositeName string) (Composite, error) {
	cmd := exec.Command(constants.KUBECTL,
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

func (kubecli *CelleryKubeCli) DeleteResource(kind, instance string) (string, error) {
	cmd := exec.Command(
		constants.KUBECTL,
		"delete",
		kind,
		instance,
		"--ignore-not-found",
	)
	displayVerboseOutput(cmd)
	return osexec.GetCommandOutput(cmd)
}

func (kubecli *CelleryKubeCli) SetVerboseMode(enable bool) {
	verboseMode = enable
}

func (kubecli *CelleryKubeCli) GetInstancesNames() ([]string, error) {
	var instances []string
	runningCellInstances, err := kubecli.GetCells()
	if err != nil {
		return nil, err
	}
	runningCompositeInstances, err := kubecli.GetComposites()
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

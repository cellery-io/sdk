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
	"os/exec"

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
	GetInstanceBytes(instanceKind, InstanceName string) ([]byte, error)
	DescribeCell(cellName string) error
	Version() (string, string, error)
	GetServices(cellName string) (Services, error)
	GetCellLogsUserComponents(cellName string) (string, error)
	GetCellLogsAllComponents(cellName string) (string, error)
	GetComponentLogs(cellName, componentName string) (string, error)
	JsonPatch(kind, instance, jsonPatch string) error
	ApplyFile(file string) error
	GetCellInstanceAsMapInterface(cell string) (map[string]interface{}, error)
	GetCompositeInstanceAsMapInterface(composite string) (map[string]interface{}, error)
	GetPodsForCell(cellName string) (Pods, error)
	GetPodsForComposite(compName string) (Pods, error)
	GetVirtualService(vs string) (VirtualService, error)
}

type CelleryKubeCli struct {
}

// NewCelleryCli returns a CelleryCli instance.
func NewCelleryKubeCli() *CelleryKubeCli {
	kubeCli := &CelleryKubeCli{}
	return kubeCli
}

func (kubeCli *CelleryKubeCli) SetVerboseMode(enable bool) {
	verboseMode = enable
}

func (kubeCli *CelleryKubeCli) DeleteResource(kind, instance string) (string, error) {
	cmd := exec.Command(
		constants.KubeCtl,
		"delete",
		kind,
		instance,
		"--ignore-not-found",
	)
	displayVerboseOutput(cmd)
	return osexec.GetCommandOutput(cmd)
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
	cmd := exec.Command(constants.KubeCtl,
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

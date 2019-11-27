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

package test

import (
	"encoding/json"
	"fmt"

	"cellery.io/cellery/components/cli/pkg/kubernetes"
)

const celleryInstance = "cells.mesh.cellery.io"
const celleryComposite = "composites.mesh.cellery.io"

type MockKubeCli struct {
	cells            kubernetes.Cells
	composites       kubernetes.Composites
	cellsBytes       map[string][]byte
	k8sServerVersion string
	k8sClientVersion string
	services         map[string]kubernetes.Services
	cellLogs         map[string]string
	componentsLogs   map[string]string
	virtualServices  map[string]kubernetes.VirtualService
}

func (kubeCli *MockKubeCli) SetVerboseMode(enable bool) {
}

func WithCells(cells kubernetes.Cells) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.cells = cells
	}
}

func WithComposites(composites kubernetes.Composites) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.composites = composites
	}
}

func WithCellsAsBytes(cellsBytes map[string][]byte) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.cellsBytes = cellsBytes
	}
}

func WithServices(services map[string]kubernetes.Services) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.services = services
	}
}

func WithCellLogs(cellLogs map[string]string) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.cellLogs = cellLogs
	}
}

func WithComponentLogs(componentLogs map[string]string) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.componentsLogs = componentLogs
	}
}

func WithVirtualServices(virtualServices map[string]kubernetes.VirtualService) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.virtualServices = virtualServices
	}
}

func SetK8sVersions(serverVersion, clientVersion string) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.k8sServerVersion = serverVersion
		cli.k8sClientVersion = clientVersion
	}
}

// NewMockKubeCli returns a mock cli for the cli.KubeCli interface.
func NewMockKubeCli(opts ...func(*MockKubeCli)) *MockKubeCli {
	cli := &MockKubeCli{}
	for _, opt := range opts {
		opt(cli)
	}
	return cli
}

// GetCells returns cell instances array.
func (kubeCli *MockKubeCli) GetCells() ([]kubernetes.Cell, error) {
	return kubeCli.cells.Items, nil
}

func (kubeCli *MockKubeCli) GetComposites() ([]kubernetes.Composite, error) {
	return kubeCli.composites.Items, nil
}

func (kubeCli *MockKubeCli) GetCell(cellName string) (kubernetes.Cell, error) {
	for _, cell := range kubeCli.cells.Items {
		if cell.CellMetaData.Name == cellName {
			return cell, nil
		}
	}
	if kubeCli.cellsBytes[cellName] != nil {
		cell := kubernetes.Cell{}
		err := json.Unmarshal(kubeCli.cellsBytes[cellName], &cell)
		if err != nil {
			return kubernetes.Cell{}, err
		}
		return cell, nil
	}
	return kubernetes.Cell{}, fmt.Errorf("cell %s not found", cellName)
}

func (kubeCli *MockKubeCli) GetComposite(compositeName string) (kubernetes.Composite, error) {
	for _, composite := range kubeCli.composites.Items {
		if composite.CompositeMetaData.Name == compositeName {
			return composite, nil
		}
	}
	return kubernetes.Composite{}, fmt.Errorf("composite %s not found", compositeName)
}

func (kubeCli *MockKubeCli) DeleteResource(kind, instance string) (string, error) {
	return "", nil
}

func (kubeCli *MockKubeCli) GetInstancesNames() ([]string, error) {
	var instanceNames []string
	for _, cell := range kubeCli.cells.Items {
		instanceNames = append(instanceNames, cell.CellMetaData.Name)
	}
	for _, composites := range kubeCli.composites.Items {
		instanceNames = append(instanceNames, composites.CompositeMetaData.Name)
	}
	return instanceNames, nil
}

func (kubeCli *MockKubeCli) GetInstanceBytes(instanceKind, InstanceName string) ([]byte, error) {
	if instanceKind == celleryInstance {
		return kubeCli.cellsBytes[InstanceName], nil
	} else if instanceKind == celleryComposite {
		return kubeCli.cellsBytes[InstanceName], nil
	}
	return nil, nil
}

func (kubeCli *MockKubeCli) DescribeCell(cellName string) error {
	for _, cell := range kubeCli.cells.Items {
		if cell.CellMetaData.Name == cellName {
			return nil
		}
	}
	return fmt.Errorf(fmt.Sprintf("cell instance %s not found", cellName))
}

func (kubeCli *MockKubeCli) Version() (string, string, error) {
	return kubeCli.k8sServerVersion, kubeCli.k8sClientVersion, nil
}

func (kubeCli *MockKubeCli) GetServices(cellName string) (kubernetes.Services, error) {
	return kubeCli.services[cellName], nil
}

func (kubeCli *MockKubeCli) GetCellLogsUserComponents(instanceName string, follow bool) (bool, error) {
	return true, nil
}

func (kubeCli *MockKubeCli) GetCellLogsAllComponents(instanceName string, follow bool) (bool, error) {
	return false, nil
}

func (kubeCli *MockKubeCli) GetComponentLogs(instanceName, componentName string, follow bool) (bool, error) {
	return false, nil
}

func (kubeCli *MockKubeCli) JsonPatch(kind, instance, jsonPatch string) error {
	return nil
}

func (kubeCli *MockKubeCli) ApplyFile(file string) error {
	return nil
}

func (kubeCli *MockKubeCli) GetCellInstanceAsMapInterface(cell string) (map[string]interface{}, error) {
	var output map[string]interface{}
	out := kubeCli.cellsBytes[cell]
	err := json.Unmarshal(out, &output)
	return output, err
}

func (kubeCli *MockKubeCli) GetCompositeInstanceAsMapInterface(composite string) (map[string]interface{}, error) {
	return nil, nil
}

func (kubeCli *MockKubeCli) GetPodsForCell(cellName string) (kubernetes.Pods, error) {
	return kubernetes.Pods{}, nil
}

func (kubeCli *MockKubeCli) GetPodsForComposite(compName string) (kubernetes.Pods, error) {
	return kubernetes.Pods{}, nil
}

func (kubeCli *MockKubeCli) GetVirtualService(vs string) (kubernetes.VirtualService, error) {
	return kubeCli.virtualServices[vs], nil
}

func (kubeCli *MockKubeCli) IsInstanceAvailable(instanceName string) (bool, error) {
	return true, nil
}

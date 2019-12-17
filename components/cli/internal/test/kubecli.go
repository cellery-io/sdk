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

	errorpkg "cellery.io/cellery/components/cli/pkg/error"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
)

const celleryInstance = "cells.mesh.cellery.io"
const celleryComposite = "composites.mesh.cellery.io"

type MockKubeCli struct {
	clusterName      string
	contexts         []string
	config           []byte
	cells            kubernetes.Cells
	components       kubernetes.Components
	composites       kubernetes.Composites
	cellsBytes       map[string][]byte
	k8sServerVersion string
	k8sClientVersion string
	services         map[string]kubernetes.Services
	virtualServices  map[string]kubernetes.VirtualService
}

// NewMockKubeCli returns a mock cli for the cli.KubeCli interface.
func NewMockKubeCli(opts ...func(*MockKubeCli)) *MockKubeCli {
	cli := &MockKubeCli{}
	for _, opt := range opts {
		opt(cli)
	}
	return cli
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

func WithComponents(components kubernetes.Components) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.components = components
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

func SetClusterName(clusterName string) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.clusterName = clusterName
	}
}

func SetContexts(contexts []string) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.contexts = contexts
	}
}

func SetConfig(config []byte) func(*MockKubeCli) {
	return func(cli *MockKubeCli) {
		cli.config = config
	}
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
	if kubeCli.k8sServerVersion != "" && kubeCli.k8sClientVersion != "" {
		return kubeCli.k8sServerVersion, kubeCli.k8sClientVersion, nil
	}
	return "Unable to connect to the server", kubeCli.k8sClientVersion, fmt.Errorf("failed to get k8s version")
}

func (kubeCli *MockKubeCli) GetServices(cellName string) (kubernetes.Services, error) {
	return kubeCli.services[cellName], nil
}

func (kubeCli *MockKubeCli) StreamCellLogsUserComponents(instanceName string, follow bool) error {
	return nil
}

func (kubeCli *MockKubeCli) StreamCellLogsAllComponents(instanceName string, follow bool) error {
	return nil
}

func (kubeCli *MockKubeCli) StreamComponentLogs(instanceName, componentName string, follow bool) error {
	return nil
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

func (kubeCli *MockKubeCli) IsInstanceAvailable(instanceName string) error {
	var canBeComposite bool
	_, err := kubeCli.GetCell(instanceName)
	if err != nil {
		if cellNotFound, _ := errorpkg.IsCellInstanceNotFoundError(instanceName, err); cellNotFound {
			canBeComposite = true
		} else {
			return fmt.Errorf("failed to check available Cells, %v", err)
		}
	} else {
		return nil
	}
	if canBeComposite {
		_, err := kubeCli.GetComposite(instanceName)
		if err != nil {
			if compositeNotFound, _ := errorpkg.IsCompositeInstanceNotFoundError(instanceName, err); compositeNotFound {
				return fmt.Errorf("instance %s not available in the runtime", instanceName)
			} else {
				return fmt.Errorf("failed to check available Composites, %v", err)
			}
		} else {
			return nil
		}
	}
	return nil
}

func (kubeCli *MockKubeCli) IsComponentAvailable(instanceName, componentName string) error {
	for _, component := range kubeCli.components.Items {
		if component.ComponentMetaData.Name == componentName {
			return nil
		}
	}
	return fmt.Errorf("component %s not found", componentName)
}

func (kubeCli *MockKubeCli) GetContext() (string, error) {
	if kubeCli.clusterName != "" {
		return kubeCli.clusterName, nil
	}
	return "", fmt.Errorf("not connected to a cluster")
}

func (kubeCli *MockKubeCli) GetContexts() ([]byte, error) {
	if kubeCli.config != nil {
		return kubeCli.config, nil
	}
	return nil, fmt.Errorf("failed to get config")
}

func (kubeCli *MockKubeCli) UseContext(context string) error {
	for _, c := range kubeCli.contexts {
		if c == context {
			return nil
		}
	}
	return fmt.Errorf("failed to use context")
}

func (kubeCli *MockKubeCli) GetMasterNodeName() (string, error) {
	return "", nil
}

func (kubeCli *MockKubeCli) ApplyLabel(itemType, itemName, labelName string, overWrite bool) error {
	return nil
}

func (kubeCli *MockKubeCli) DeletePersistedVolume(persistedVolume string) error {
	return nil
}

func (kubeCli *MockKubeCli) DeleteAllCells() error {
	return nil
}

func (kubeCli *MockKubeCli) DeleteNameSpace(nameSpace string) error {
	return nil
}

func (kubeCli *MockKubeCli) SetNamespace(namespace string) error {
	return nil
}

func (kubeCli *MockKubeCli) CreateNamespace(namespace string) error {
	return nil
}

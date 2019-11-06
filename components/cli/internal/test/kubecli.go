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

package test

import (
	"fmt"

	"github.com/cellery-io/sdk/components/cli/kubernetes"
)

type MockKubeCli struct {
	cells      kubernetes.Cells
	composites kubernetes.Composites
}

type option func(*MockKubeCli)

func WithCells(cells kubernetes.Cells) option {
	return func(cli *MockKubeCli) {
		cli.cells = cells
	}
}

func WithComposites(composites kubernetes.Composites) option {
	return func(cli *MockKubeCli) {
		cli.composites = composites
	}
}

// NewMockKubeCli returns a mock cli for the cli.KubeCli interface.
func NewMockKubeCli(opts ...option) *MockKubeCli {
	cli := &MockKubeCli{}
	for _, opt := range opts {
		opt(cli)
	}
	return cli
}

// GetCells returns cell instances array.
func (kubecli *MockKubeCli) GetCells() ([]kubernetes.Cell, error) {
	return kubecli.cells.Items, nil
}

func (kubecli *MockKubeCli) GetComposites() ([]kubernetes.Composite, error) {
	return kubecli.composites.Items, nil
}

func (kubecli *MockKubeCli) GetCell(cellName string) (kubernetes.Cell, error) {
	for _, cell := range kubecli.cells.Items {
		if cell.CellMetaData.Name == cellName {
			return cell, nil
		}
	}
	return kubernetes.Cell{}, fmt.Errorf("cell %s does not exist", cellName)
}

func (kubecli *MockKubeCli) GetComposite(compositeName string) (kubernetes.Composite, error) {
	for _, composite := range kubecli.composites.Items {
		if composite.CompositeMetaData.Name == compositeName {
			return composite, nil
		}
	}
	return kubernetes.Composite{}, fmt.Errorf("composite %s does not exist", compositeName)
}

func (kubecli *MockKubeCli) DeleteResource(kind, instance string) (string, error) {
	return "", nil
}

func (kubecli *MockKubeCli) SetVerboseMode(enable bool) {
}

func (kubecli *MockKubeCli) GetInstancesNames() ([]string, error) {
	var instanceNames []string
	for _, cell := range kubecli.cells.Items {
		instanceNames = append(instanceNames, cell.CellMetaData.Name)
	}
	for _, composites := range kubecli.composites.Items {
		instanceNames = append(instanceNames, composites.CompositeMetaData.Name)
	}
	return instanceNames, nil
}

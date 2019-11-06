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
	"github.com/cellery-io/sdk/components/cli/kubernetes"
)

type MockKubeCli struct {
	cells kubernetes.Cells
	composites kubernetes.Composites
}

// NewMockKubeCli returns a mock cli for the cli.KubeCli interface.
func NewMockKubeCli(cells kubernetes.Cells, composites kubernetes.Composites) *MockKubeCli {
	cli := &MockKubeCli{
		cells: cells,
		composites:composites,
	}
	return cli
}

// GetCells returns cell instances array.
func (kubecli *MockKubeCli) GetCells() ([]kubernetes.Cell, error) {
	return kubecli.cells.Items, nil
}

func (kubecli *MockKubeCli) GetComposites() (kubernetes.Composites, error) {
	return kubecli.composites, nil
}

func (kubecli *MockKubeCli)DeleteResource(kind, instance string) (string, error) {
	return "", nil
}

func (kubecli *MockKubeCli)SetVerboseMode(enable bool) {
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

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

package instance

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/cellery-io/sdk/components/cli/internal/test"
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
)

func TestRunPatchForSingleComponent(t *testing.T) {
	petBeAutoCell, err := ioutil.ReadFile(filepath.Join("testdata", "cells", "pet-be-auto.json"))
	if err != nil {
		t.Errorf("failed to read mock cell yaml file")
	}
	cellMap := make(map[string][]byte)
	cellMap["pet-be-auto"] = petBeAutoCell
	cells := kubernetes.Cells{
		Items: []kubernetes.Cell{
			{
				CellMetaData: kubernetes.K8SMetaData{
					Name:              "pet-be-auto",
					CreationTimestamp: "2019-10-18T11:40:36Z",
				},
				CellStatus: kubernetes.CellStatus{
					Status: "Ready",
				},
			},
		},
	}
	mockKubeCli := test.NewMockKubeCli(test.WithCells(cells), test.WithCellsAsBytes(cellMap))
	mockCli := test.NewMockCli(test.SetKubeCli(mockKubeCli))
	tests := []struct {
		name           string
		MockCli        *test.MockCli
		instance       string
		component      string
		containerImage string
		containerName  string
	}{
		{
			name:           "patch single component",
			MockCli:        test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			instance:       "pet-be-auto",
			component:      "controller",
			containerImage: "foo/bar:2.0.0",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunPatchForSingleComponent(mockCli, testIteration.instance, testIteration.component, testIteration.containerImage, "", nil)
			if err != nil {
				t.Errorf("error in RunPatchForSingleComponent, %v", err)
			}
		})
	}
}

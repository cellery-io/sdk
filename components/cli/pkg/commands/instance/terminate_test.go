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
	"testing"

	"cellery.io/cellery/components/cli/internal/test"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
)

func TestTerminateInstanceSuccess(t *testing.T) {
	cells := kubernetes.Cells{
		Items: []kubernetes.Cell{
			{
				CellMetaData: kubernetes.K8SMetaData{
					Name:              "employee",
					CreationTimestamp: "2019-10-18T11:40:36Z",
				},
			},
			{
				CellMetaData: kubernetes.K8SMetaData{
					Name:              "stock",
					CreationTimestamp: "2019-10-19T11:40:36Z",
				},
			},
		},
	}
	composites := kubernetes.Composites{
		Items: []kubernetes.Composite{
			{
				CompositeMetaData: kubernetes.K8SMetaData{
					Name:              "hr",
					CreationTimestamp: "2019-10-20T11:40:36Z",
				},
			},
		},
	}
	mockKubeCli := test.NewMockKubeCli(test.WithCells(cells), test.WithComposites(composites))
	tests := []struct {
		name         string
		MockCli      *test.MockCli
		instances    []string
		terminateAll bool
	}{
		{
			name:         "terminate single existing cell instance",
			MockCli:      test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			instances:    []string{"employee"},
			terminateAll: false,
		},
		{
			name:         "terminate single existing composite instance",
			MockCli:      test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			instances:    []string{"hr"},
			terminateAll: false,
		},
		{
			name:         "terminate multiple existing instances",
			MockCli:      test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			instances:    []string{"stock", "hr"},
			terminateAll: false,
		},
		{
			name:         "terminate all instances",
			MockCli:      test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			instances:    []string{},
			terminateAll: true,
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunTerminate(testIteration.MockCli, testIteration.instances, testIteration.terminateAll)
			if err != nil {
				t.Errorf("getCellTableData err, %v", err)
			}
		})
	}
}

func TestTerminateInstanceFailure(t *testing.T) {
	cells := kubernetes.Cells{
		Items: []kubernetes.Cell{
			{
				CellMetaData: kubernetes.K8SMetaData{
					Name:              "employee",
					CreationTimestamp: "2019-10-18T11:40:36Z",
				},
			},
			{
				CellMetaData: kubernetes.K8SMetaData{
					Name:              "stock",
					CreationTimestamp: "2019-10-18T11:40:36Z",
				},
			},
		},
	}
	mockKubeCli := test.NewMockKubeCli(test.WithCells(cells))
	tests := []struct {
		name         string
		want         string
		MockCli      *test.MockCli
		instances    []string
		terminateAll bool
	}{
		{
			name:         "terminate non existing instance",
			want:         "foo",
			MockCli:      test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			instances:    []string{"foo"},
			terminateAll: false,
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			actual := RunTerminate(testIteration.MockCli, testIteration.instances, testIteration.terminateAll)
			expected := "error terminating cell instances, instance: foo does not exist"
			if actual.Error() != expected {
				t.Errorf("getCellTableData err, %v", actual.Error())
			}
		})
	}
}

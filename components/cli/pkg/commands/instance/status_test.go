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

	"github.com/google/go-cmp/cmp"

	"github.com/cellery-io/sdk/components/cli/internal/test"
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
)

func TestRunStatus(t *testing.T) {
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
	composites := kubernetes.Composites{
		Items: []kubernetes.Composite{
			{
				CompositeMetaData: kubernetes.K8SMetaData{
					Name:              "hr",
					CreationTimestamp: "2019-10-18T11:40:36Z",
				},
			},
			{
				CompositeMetaData: kubernetes.K8SMetaData{
					Name:              "job",
					CreationTimestamp: "2019-10-18T11:40:36Z",
				},
			},
		},
	}
	mockCli := test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli(test.WithCells(cells), test.WithComposites(composites))))

	tests := []struct {
		name     string
		instance string
	}{
		{
			name:     "status of cell instance",
			instance: "employee",
		},
		{
			name:     "status of composite instance",
			instance: "hr",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunStatus(mockCli, tst.instance)
			if err != nil {
				t.Errorf("error in RunStatus, %v", err)
			}
		})
	}
}

func TestGetCellSummary(t *testing.T) {
	cells := kubernetes.Cells{
		Items: []kubernetes.Cell{
			{
				CellMetaData: kubernetes.K8SMetaData{
					Name:              "employee",
					CreationTimestamp: "2019-10-18T11:40:36Z",
				},
				CellStatus: kubernetes.CellStatus{
					Status: "Ready",
				},
			},
			{
				CellMetaData: kubernetes.K8SMetaData{
					Name:              "stock",
					CreationTimestamp: "2019-10-19T11:40:36Z",
				},
				CellStatus: kubernetes.CellStatus{
					Status: "Not Ready",
				},
			},
		},
	}
	mockKubeCli := test.NewMockKubeCli(test.WithCells(cells))
	tests := []struct {
		name              string
		MockCli           *test.MockCli
		instance          string
		expectedStatus    string
		expectedTimeStamp string
	}{
		{
			name:              "status of ready cell instance",
			MockCli:           test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			instance:          "employee",
			expectedStatus:    "Ready",
			expectedTimeStamp: "2019-10-18T11:40:36Z",
		},
		{
			name:              "status of not ready composite instance",
			MockCli:           test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			instance:          "stock",
			expectedStatus:    "Not Ready",
			expectedTimeStamp: "2019-10-20T11:40:36Z",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			_, status, err := getCellSummary(testIteration.MockCli, testIteration.instance)
			if err != nil {
				t.Errorf("getCellSummary err, %v", err)
			}
			if diff := cmp.Diff(testIteration.expectedStatus, status); diff != "" {
				t.Errorf("getCellSummary: status (-want, +got)\n%v", diff)
			}
		})
	}
}

func TestGetCompositeSummary(t *testing.T) {
	composites := kubernetes.Composites{
		Items: []kubernetes.Composite{
			{
				CompositeMetaData: kubernetes.K8SMetaData{
					Name:              "hr",
					CreationTimestamp: "2019-10-20T11:40:36Z",
				},
				CompositeStatus: kubernetes.CompositeStatus{
					Status: "Ready",
				},
			},
			{
				CompositeMetaData: kubernetes.K8SMetaData{
					Name:              "job",
					CreationTimestamp: "2019-10-20T11:40:36Z",
				},
				CompositeStatus: kubernetes.CompositeStatus{
					Status: "Not Ready",
				},
			},
		},
	}
	mockKubeCli := test.NewMockKubeCli(test.WithComposites(composites))
	tests := []struct {
		name              string
		MockCli           *test.MockCli
		instance          string
		expectedStatus    string
		expectedTimeStamp string
	}{
		{
			name:              "status of ready composite instance",
			MockCli:           test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			instance:          "hr",
			expectedStatus:    "Ready",
			expectedTimeStamp: "2019-10-18T11:40:36Z",
		},
		{
			name:              "status of not ready composite instance",
			MockCli:           test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			instance:          "job",
			expectedStatus:    "Not Ready",
			expectedTimeStamp: "2019-10-20T11:40:36Z",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			_, status, err := getCompositeSummary(testIteration.MockCli, testIteration.instance)
			if err != nil {
				t.Errorf("getCompositeSummary err, %v", err)
			}
			if diff := cmp.Diff(testIteration.expectedStatus, status); diff != "" {
				t.Errorf("getCompositeSummary: status (-want, +got)\n%v", diff)
			}
		})
	}
}

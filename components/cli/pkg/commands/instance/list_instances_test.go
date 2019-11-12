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

func TestRunListInstances(t *testing.T) {
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
		name string
	}{
		{
			name: "list instances",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunListInstances(mockCli)
			if err != nil {
				t.Errorf("error in RunListInstances, %v", err)
			}
		})
	}
}

func TestGetCellTableData(t *testing.T) {
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
	tests := []struct {
		name        string
		want        string
		MockKubeCli *test.MockKubeCli
	}{
		{
			name:        "list instances with single cell instance",
			want:        "employee",
			MockKubeCli: test.NewMockKubeCli(test.WithCells(cells)),
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			tableData, err := getCellTableData(tst.MockKubeCli)
			if err != nil {
				t.Errorf("error in getCellTableData")
			}
			got := tableData[0][0]
			if diff := cmp.Diff(tst.want, got); diff != "" {
				t.Errorf("getCellTableData (-want, +got)\n%v", diff)
			}
		})
	}
}

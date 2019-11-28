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

	"cellery.io/cellery/components/cli/internal/test"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
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
	tests := []struct {
		name             string
		instancesRunning bool
		expected         string
		mockCli          *test.MockCli
	}{
		{
			name:             "list instances with cells and composites instances running",
			instancesRunning: true,
			expected:         "",
			mockCli:          test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli(test.WithCells(cells), test.WithComposites(composites)))),
		},
		{
			name:             "list instances without instances running",
			instancesRunning: false,
			expected:         "No running instances.\n",
			mockCli:          test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli())),
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunListInstances(tst.mockCli)
			if tst.instancesRunning {
				if err != nil {
					t.Errorf("error in RunListInstances, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expected, tst.mockCli.OutBuffer().String()); diff != "" {
					t.Errorf("RunListInstances: unexpected output (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

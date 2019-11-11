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

	"github.com/cellery-io/sdk/components/cli/internal/test"
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
)

func TestRunListDependencies(t *testing.T) {
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
			{
				CellMetaData: kubernetes.K8SMetaData{
					Name:              "hr",
					CreationTimestamp: "2019-10-19T11:40:36Z",
					Annotations: kubernetes.CellAnnotations{
						Dependencies: "[{\"org\":\"myorg\",\"name\":\"employee\",\"version\":\"1.0.0\",\"instance\":\"employee\",\"kind\":\"Cell\"},{\"org\":\"myorg\",\"name\":\"stock\",\"version\":\"1.0.0\",\"instance\":\"stock\",\"kind\":\"Cell\"}]",
					},
				},
			},
		},
	}
	mockKubeCli := test.NewMockKubeCli(test.WithCells(cells))
	mockCli := test.NewMockCli(test.SetKubeCli(mockKubeCli))

	tests := []struct {
		name     string
		instance string
	}{
		{
			name:     "list dependencies",
			instance: "hr",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunListDependencies(mockCli, testIteration.instance)
			if err != nil {
				t.Errorf("error in RunListDependencies, %v", err)
			}
		})
	}
}

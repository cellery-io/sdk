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
					Name:              "hr",
					CreationTimestamp: "2019-10-19T11:40:36Z",
					Annotations: kubernetes.CellAnnotations{
						Dependencies: "[{\"org\":\"myorg\",\"name\":\"employee\",\"version\":\"1.0.0\",\"instance\":\"employee\",\"kind\":\"Cell\"},{\"org\":\"myorg\",\"name\":\"stock\",\"version\":\"1.0.0\",\"instance\":\"stock\",\"kind\":\"Cell\"}]",
					},
				},
			},
		},
	}
	composites := kubernetes.Composites{
		Items: []kubernetes.Composite{
			{
				CompositeMetaData: kubernetes.K8SMetaData{
					Name:              "stock",
					CreationTimestamp: "2019-10-18T11:40:36Z",
				},
			},
			{
				CompositeMetaData: kubernetes.K8SMetaData{
					Name:              "foo",
					CreationTimestamp: "2019-10-19T11:40:36Z",
					Annotations: kubernetes.CellAnnotations{
						Dependencies: "[{\"org\":\"myorg\",\"bar\":\"employee\",\"version\":\"1.0.0\",\"instance\":\"bar\",\"kind\":\"Composite\"},{\"org\":\"myorg\",\"name\":\"zoo\",\"version\":\"1.0.0\",\"instance\":\"zoo\",\"kind\":\"Composite\"}]",
					},
				},
			},
		},
	}
	mockKubeCli := test.NewMockKubeCli(test.WithCells(cells), test.WithComposites(composites))
	mockCli := test.NewMockCli(test.SetKubeCli(mockKubeCli))

	tests := []struct {
		name             string
		instance         string
		expectedToPass   bool
		expectedErrorMsg string
	}{
		{
			name:             "list dependencies of cell instance with dependencies",
			instance:         "hr",
			expectedToPass:   true,
			expectedErrorMsg: "",
		},
		{
			name:             "list dependencies of cell instance without dependencies",
			instance:         "employee",
			expectedToPass:   false,
			expectedErrorMsg: "no dependencies found in instance employee",
		},
		{
			name:             "list dependencies of composite instance with dependencies",
			instance:         "foo",
			expectedToPass:   true,
			expectedErrorMsg: "",
		},
		{
			name:             "list dependencies of composite instance without dependencies",
			instance:         "stock",
			expectedToPass:   false,
			expectedErrorMsg: "no dependencies found in instance stock",
		},
		{
			name:             "list dependencies of non-existing instance",
			instance:         "hello",
			expectedToPass:   false,
			expectedErrorMsg: "failed to retrieve dependencies of hello, instance not available in the runtime",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunListDependencies(mockCli, tst.instance)
			if tst.expectedToPass {
				if err != nil {
					t.Errorf("error in RunListDependencies, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expectedErrorMsg, err.Error()); diff != "" {
					t.Errorf("invalid error message (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

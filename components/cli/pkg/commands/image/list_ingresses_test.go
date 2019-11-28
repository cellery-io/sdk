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

package image

import (
	"path/filepath"
	"testing"

	"cellery.io/cellery/components/cli/internal/test"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
)

func TestRunListIngresses(t *testing.T) {
	mockRepo := filepath.Join("testdata", "repo")
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo))
	cells := kubernetes.Cells{
		Items: []kubernetes.Cell{
			{
				CellMetaData: kubernetes.K8SMetaData{
					Name:              "employee",
					CreationTimestamp: "2019-10-18T11:40:36Z",
				},
				CellSpec: kubernetes.CellSpec{
					GateWayTemplate: kubernetes.Gateway{
						GatewaySpec: kubernetes.GatewaySpec{
							Ingress: kubernetes.Ingress{
								HttpApis: []kubernetes.GatewayHttpApi{
									{
										Context: "foo",
										Version: "1.0.0",
										Definitions: []kubernetes.APIDefinition{
											{
												Path:   "/bar",
												Method: "POST",
											},
										},
										Global: true,
									},
								},
							},
						},
					},
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
				CompositeSpec: kubernetes.CompositeSpec{
					ComponentTemplates: []kubernetes.ComponentTemplate{
						{
							Spec: kubernetes.ComponentTemplateSpec{
								Ports: []kubernetes.Port{
									{
										Name: "my-port",
										Port: 9999,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	tests := []struct {
		name    string
		arg     string
		mockCli *test.MockCli
	}{
		{
			name:    "list ingresses of cell instance",
			arg:     "employee",
			mockCli: test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli(test.WithCells(cells))), test.SetFileSystem(mockFileSystem)),
		},
		{
			name:    "list ingresses of composite instance",
			arg:     "hr",
			mockCli: test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli(test.WithComposites(composites))), test.SetFileSystem(mockFileSystem)),
		},
		{
			name:    "list ingresses of cell image",
			arg:     "myorg/hello:1.0.0",
			mockCli: test.NewMockCli(test.SetFileSystem(mockFileSystem)),
		},
		{
			name:    "list ingresses of composite image",
			arg:     "myorg/stock-comp:1.0.0",
			mockCli: test.NewMockCli(test.SetFileSystem(mockFileSystem)),
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunListIngresses(tst.mockCli, tst.arg)
			if err != nil {
				t.Errorf("error in RunListIngresses, %v", err)
			}
		})
	}
}

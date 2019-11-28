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

func TestRunListComponentsForInstance(t *testing.T) {
	employeeServises := kubernetes.Services{
		Items: []kubernetes.Service{
			{
				Metadata: kubernetes.ServiceMetaData{
					Name: "employeeservice",
				},
			},
		},
	}
	servicesMap := make(map[string]kubernetes.Services)
	servicesMap["employee"] = employeeServises
	mockKubeCli := test.NewMockKubeCli(test.WithServices(servicesMap))
	mockCli := test.NewMockCli(test.SetKubeCli(mockKubeCli))
	tests := []struct {
		name     string
		want     string
		instance string
	}{
		{
			name:     "list components with single cell instance",
			want:     "employee",
			instance: "employee",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunListComponents(mockCli, testIteration.instance)
			if err != nil {
				t.Errorf("error in RunListComponents, %v", err)
			}
		})
	}
}

func TestRunListComponentsForImage(t *testing.T) {
	mockRepo := filepath.Join("testdata", "repo")
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))

	tests := []struct {
		name  string
		want  string
		image string
	}{
		{
			name:  "list components with single cell image",
			want:  "employee",
			image: "myorg/hello:1.0.0",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunListComponents(mockCli, testIteration.image)
			if err != nil {
				t.Errorf("error in RunListComponents, %v", err)
			}
		})
	}
}

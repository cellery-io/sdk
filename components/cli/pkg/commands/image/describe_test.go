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

	"github.com/cellery-io/sdk/components/cli/internal/test"
	"github.com/cellery-io/sdk/components/cli/kubernetes"
)

func TestRunDescribeInstance(t *testing.T) {
	cells := kubernetes.Cells{
		Items: []kubernetes.Cell{
			{
				CellMetaData: kubernetes.K8SMetaData{
					Name:              "employee",
					CreationTimestamp: "2019-10-18T11:40:36Z",
				},
			},
		},
	}
	tests := []struct {
		name     string
		instance string
		want     string
		MockCli  *test.MockCli
	}{
		{
			name:     "describe cell instance",
			want:     "employee",
			instance: "employee",
			MockCli:  test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli(test.WithCells(cells)))),
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunDescribe(testIteration.MockCli, testIteration.instance)
			if err != nil {
				t.Errorf("error in RunDescribe instance")
			}
		})
	}
}

func TestRunDescribeImage(t *testing.T) {
	mockRepo := filepath.Join("testdata", "repo")
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))
	tests := []struct {
		name    string
		image   string
		want    string
		MockCli *test.MockCli
	}{
		{
			name:  "describe cell instance",
			want:  "employee",
			image: "myorg/hello:1.0.0",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunDescribe(mockCli, testIteration.image)
			if err != nil {
				t.Errorf("error in RunDescribe image")
			}
		})
	}
}

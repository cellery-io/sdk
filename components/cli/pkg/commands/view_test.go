/*
 * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http:www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package commands

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cellery-io/sdk/components/cli/internal/test"
)

func TestRunView(t *testing.T) {
	mockRepo := filepath.Join("testdata", "repo")
	mockCelleryInstallationDir := filepath.Join("testdata", "cellery")
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo), test.SetCelleryInstallationDir(mockCelleryInstallationDir))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))

	tests := []struct {
		name  string
		image string
	}{
		{
			name:  "view cell image",
			image: "myorg/hello:1.0.0",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunView(mockCli, testIteration.image)
			if err != nil {
				t.Errorf("error in RunView, %v", err)
			}
		})
	}
}

func TestRunViewError(t *testing.T) {
	mockRepo := filepath.Join("testdata", "repo")
	mockCelleryInstallationDir := filepath.Join("testdata", "cellery")
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo), test.SetCelleryInstallationDir(mockCelleryInstallationDir))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))

	tests := []struct {
		name  string
		image string
	}{
		{
			name:  "view non-existing cell image",
			image: "myorg/foo:1.0.0",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunView(mockCli, testIteration.image)
			if diff := cmp.Diff("error occurred while unpacking Cell Image, open testdata/repo/myorg/foo/1.0.0/foo.zip: no such file or directory", err.Error()); diff != "" {
				t.Errorf("RunView: error (-want, +got)\n%v", diff)
			}
		})
	}
}

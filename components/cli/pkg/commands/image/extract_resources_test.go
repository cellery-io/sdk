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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cellery-io/sdk/components/cli/internal/test"
)

func TestRunExtractResources(t *testing.T) {
	currentDir, err := ioutil.TempDir("", "current-dir")
	if err != nil {
		t.Errorf("error setting mock current dir")
	}
	mockRepo := filepath.Join("testdata", "repo")
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo), test.SetCurrentDir(currentDir))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))
	extractDir, err := ioutil.TempDir("", "extract-dir")
	if err != nil {
		t.Errorf("failed create dir to extract to")
	}

	tests := []struct {
		name    string
		image   string
		outPath string
	}{
		{
			name:    "extract resources of existing cell image with output path",
			image:   "myorg/hello:1.0.0",
			outPath: extractDir,
		},
		{
			name:    "extract resources of existing cell image without output path",
			image:   "myorg/hello:1.0.0",
			outPath: "",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunExtractResources(mockCli, testIteration.image, testIteration.outPath)
			if err != nil {
				t.Errorf("error in Resources, %v", err)
			}
		})
	}
}

func TestRunExtractResourcesError(t *testing.T) {
	mockRepo := filepath.Join("testdata", "repo")
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))
	extractDir, err := ioutil.TempDir("", "extract-dir")
	if err != nil {
		t.Errorf("failed create dir to extract to")
	}

	tests := []struct {
		name  string
		image string
	}{
		{
			name:  "extract resources of non-existing cell image",
			image: "myorg/foo:1.0.0",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunExtractResources(mockCli, testIteration.image, extractDir)
			if diff := cmp.Diff("failed to extract resources for image myorg/foo:1.0.0, image not Found", err.Error()); diff != "" {
				t.Errorf("RunExtractResources: error (-want, +got)\n%v", diff)
			}
		})
	}
}

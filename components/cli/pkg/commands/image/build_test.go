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
	"os"
	"path/filepath"
	"testing"

	"github.com/cellery-io/sdk/components/cli/internal/test"
)

func TestRunBuild(t *testing.T) {
	currentDir, err := ioutil.TempDir("", "current-dir")
	if err != nil {
		t.Errorf("failed to create current dir")
	}
	defer func() {
		if err := os.RemoveAll(currentDir); err != nil {
			t.Errorf("failed to remove current dir")
		}
	}()
	balFileContent, err := ioutil.ReadFile(filepath.Join("testdata", "project", "foo.bal"))
	if err != nil {
		t.Errorf("Failed to read foo.bal: %v", err)
	}
	balFile, err := os.Create(filepath.Join(currentDir, "foo.bal"))
	if err != nil {
		t.Errorf("Failed to create bal file: %v", err)
	}
	err = ioutil.WriteFile(balFile.Name(), []byte(balFileContent), os.ModePerm)
	if err != nil {
		t.Errorf("Failed to write to bal file: %v", err)
	}
	mockRepo, err := ioutil.TempDir("", "repo-dir")
	if err != nil {
		t.Errorf("failed to create mock repo dir")
	}
	mockFileSystem := test.NewMockFileSystem(test.SetCurrentDir(currentDir), test.SetRepository(mockRepo))
	yamlContent, err := ioutil.ReadFile(filepath.Join("testdata", "project", "target", "cellery", "foo.yaml"))
	if err != nil {
		t.Errorf("Failed to read foo.yaml: %v", err)
	}
	metadataJson, err := ioutil.ReadFile(filepath.Join("testdata", "project", "target", "cellery", "metadata.json"))
	if err != nil {
		t.Errorf("Failed to read metadata.json: %v", err)
	}
	mockBalExecutor := test.NewMockBalExecutor(test.SetBalCurrentDir(currentDir),
		test.SetYamlName("foo.yaml"),
		test.SetYamlContent(yamlContent),
		test.SetMetadataJsonContent(metadataJson))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem), test.SetBalExecutor(mockBalExecutor))

	tests := []struct {
		name  string
		image string
	}{
		{
			name:  "build image",
			image: "myorg/foo:1.0.0",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunBuild(mockCli, tst.image, balFile.Name())
			if err != nil {
				t.Errorf("error in RunBuild, %v", err)
			}
		})
	}
}

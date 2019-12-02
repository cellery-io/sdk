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

	"cellery.io/cellery/components/cli/internal/test"
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
	tempRepo, err := ioutil.TempDir("", "repo")
	if err != nil {
		t.Errorf("error creating temp repo, %v", err)
	}
	mockRepo := filepath.Join("testdata", "repo")
	if copyDir(mockRepo, tempRepo); err != nil {
		t.Errorf("error copying mock repo to temp repo, %v", err)
	}
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(tempRepo), test.SetCurrentDir(currentDir))

	// Test data for building foo.bal
	fooBal, err := copyFile(filepath.Join("testdata", "project", "foo.bal"), filepath.Join(currentDir, "foo.bal"))
	if err != nil {
		t.Errorf("failed to copy foo.bal to mock location")
	}
	fooYamlContent, err := ioutil.ReadFile(filepath.Join("testdata", "project", "build_artifacts", "foo.yaml"))
	if err != nil {
		t.Errorf("Failed to read foo.yaml: %v", err)
	}
	fooMetadataJson, err := ioutil.ReadFile(filepath.Join("testdata", "project", "build_artifacts", "foo_metadata.json"))
	if err != nil {
		t.Errorf("Failed to read metadata.json: %v", err)
	}
	fooReferenceJson, err := ioutil.ReadFile(filepath.Join("testdata", "project", "build_artifacts", "foo_reference.json"))
	if err != nil {
		t.Errorf("Failed to read reference.json: %v", err)
	}
	// Test data for building hr.bal
	hrBal, err := copyFile(filepath.Join("testdata", "project", "hr.bal"), filepath.Join(currentDir, "hr.bal"))
	if err != nil {
		t.Errorf("failed to copy hr.bal to mock location")
	}
	hrYamlContent, err := ioutil.ReadFile(filepath.Join("testdata", "project", "build_artifacts", "hr.yaml"))
	if err != nil {
		t.Errorf("Failed to read hr.yaml: %v", err)
	}
	hrMetadataJson, err := ioutil.ReadFile(filepath.Join("testdata", "project", "build_artifacts", "hr_metadata.json"))
	if err != nil {
		t.Errorf("Failed to read metadata.json: %v", err)
	}
	hrReferenceJson, err := ioutil.ReadFile(filepath.Join("testdata", "project", "build_artifacts", "hr_reference.json"))
	if err != nil {
		t.Errorf("Failed to read reference.json: %v", err)
	}
	tests := []struct {
		name          string
		file          *os.File
		image         string
		yamlName      string
		yaml          []byte
		metadataJson  []byte
		referenceJson []byte
	}{
		{
			name:          "build image",
			image:         "myorg/foo:1.0.0",
			file:          fooBal,
			yamlName:      "foo.yaml",
			yaml:          fooYamlContent,
			metadataJson:  fooMetadataJson,
			referenceJson: fooReferenceJson,
		},
		{
			name:          "build image with dependencies",
			image:         "myorg/hr:1.0.0",
			file:          hrBal,
			yamlName:      "hr.yaml",
			yaml:          hrYamlContent,
			metadataJson:  hrMetadataJson,
			referenceJson: hrReferenceJson,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			mockBalExecutor := test.NewMockBalExecutor(test.SetBalCurrentDir(currentDir),
				test.SetYamlName(tst.yamlName),
				test.SetYamlContent(tst.yaml),
				test.SetMetadataJsonContent(tst.metadataJson),
				test.SetReferenceJsonContent(tst.referenceJson))
			err := RunBuild(test.NewMockCli(test.SetFileSystem(mockFileSystem), test.SetBalExecutor(mockBalExecutor)), tst.image, tst.file.Name())
			if err != nil {
				t.Errorf("error in RunBuild, %v", err)
			}
		})
	}
}

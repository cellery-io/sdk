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
	"github.com/cellery-io/sdk/components/cli/pkg/image"
)

func TestRunRun(t *testing.T) {
	currentDir, err := ioutil.TempDir("", "current-dir")
	if err != nil {
		t.Errorf("failed to create current dir")
	}
	defer func() {
		if err := os.RemoveAll(currentDir); err != nil {
			t.Errorf("failed to remove current dir")
		}
	}()
	mockFileSystem := test.NewMockFileSystem(test.SetCurrentDir(currentDir), test.SetRepository(filepath.Join(
		"testdata", "repo")))
	mockBalExecutor := test.NewMockBalExecutor(test.SetBalCurrentDir(currentDir))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem), test.SetBalExecutor(mockBalExecutor))

	tests := []struct {
		name              string
		image             string
		instance          string
		startDependencies bool
		shareDependencies bool
		dependencyLinks   []string
		envVars           []string
	}{
		{
			name:              "run image",
			image:             "myorg/hello:1.0.0",
			instance:          "hello",
			startDependencies: false,
			shareDependencies: false,
			dependencyLinks:   nil,
			envVars:           nil,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunRun(mockCli, tst.image, tst.instance, tst.startDependencies, tst.shareDependencies,
				tst.dependencyLinks, tst.envVars)
			if err != nil {
				t.Errorf("error in RunRun, %v", err)
			}
		})
	}
}

func TestStartCellInstance(t *testing.T) {
	mockBalExecutor := test.NewMockBalExecutor()
	mockCli := test.NewMockCli(test.SetBalExecutor(mockBalExecutor))
	imageDir, err := ioutil.TempDir("", "temp")
	if err != nil {
		t.Errorf("Failed to create image dir: %v", err)
	}
	defer func() { os.RemoveAll(imageDir) }()
	err = os.MkdirAll(filepath.Join(imageDir, "src"), os.ModePerm)
	if err != nil {
		t.Errorf("Failed to create src directory: %v", err)
	}
	_, err = os.Create(filepath.Join(imageDir, "src", "hello.bal"))
	if err != nil {
		t.Errorf("Failed to create bal file: %v", err)
	}
	cellImage := &image.CellImageName{
		Organization: "myorg",
		Name:         "hello",
		Version:      "1.0.0",
	}
	cellImageMetadata := &image.MetaData{
		CellImageName: *cellImage,
	}
	mainNode := &dependencyTreeNode{
		Instance:  "hello",
		IsRunning: false,
		IsShared:  false,
		MetaData:  cellImageMetadata,
	}
	var instanceEnvVars []*environmentVariable
	rootNodeDependencies := map[string]*dependencyInfo{}
	err = startCellInstance(mockCli, imageDir, "hello", mainNode, instanceEnvVars, false,
		rootNodeDependencies, false)
	if err != nil {
		t.Errorf("startCellInstance failed: %v", err)
	}
}

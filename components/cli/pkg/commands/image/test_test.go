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

// Tests the RunTest function when an image with tests is given
func TestRunTest(t *testing.T) {
	currentDir, err := ioutil.TempDir("", "current-dir")
	if err != nil {
		t.Errorf("failed to create current dir")
	}
	defer func() {
		if err := os.RemoveAll(currentDir); err != nil {
			t.Errorf("failed to remove current dir")
		}
	}()
	mockFileSystem := test.NewMockFileSystem(test.SetCelleryInstallationDir(filepath.Join("testdata", "cellery")),
		test.SetCurrentDir(currentDir), test.SetRepository(filepath.Join("testdata", "repo")))
	mockBalExecutor := test.NewMockBalExecutor(test.SetBalCurrentDir(currentDir))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem),
		test.SetBalExecutor(mockBalExecutor),
		test.SetRuntime(test.NewMockRuntime()), test.SetKubeCli(test.NewMockKubeCli()))
	tests := []struct {
		name                string
		image               string
		instance            string
		startDependencies   bool
		shareDependencies   bool
		dependencyLinks     []string
		envVars             []string
		assumeYes           bool
		isDebug             bool
		verbose             bool
		disableTelepresence bool
		incell              bool
		projectLocation     string
	}{
		{
			name:                "test image",
			image:               "myorg/hello-with-tests:1.0.0",
			instance:            "hello",
			startDependencies:   false,
			shareDependencies:   false,
			dependencyLinks:     nil,
			envVars:             nil,
			assumeYes:           false,
			isDebug:             false,
			verbose:             false,
			disableTelepresence: false,
			incell:              false,
			projectLocation:     "",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunTest(mockCli, tst.image, tst.instance, tst.startDependencies, tst.shareDependencies,
				tst.dependencyLinks, tst.envVars, tst.assumeYes, tst.isDebug, tst.verbose,
				tst.disableTelepresence, tst.incell, tst.projectLocation)
			if err != nil {
				t.Errorf("error in RunTest, %v", err)
			}
		})
	}
}

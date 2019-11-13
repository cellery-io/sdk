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

package project

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cellery-io/sdk/components/cli/internal/test"
)

func TestRunInit(t *testing.T) {
	currentDir, err := ioutil.TempDir("", "current-dir")
	if err != nil {
		t.Errorf("failed to create current dir")
	}
	defer func() {
		if err := os.RemoveAll(currentDir); err != nil {
			t.Errorf("failed to remove current dir")
		}
	}()
	mockFileSystem := test.NewMockFileSystem(test.SetCurrentDir(currentDir))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))
	expectedContent, err := ioutil.ReadFile(filepath.Join("testdata", "expected", "foo.bal"))

	tests := []struct {
		name    string
		project string
	}{
		{
			name:    "init project",
			project: "foo",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunInit(mockCli, tst.project, "")
			if err != nil {
				t.Errorf("error in RunInspect, %v", err)
			}
			actualContent, err := ioutil.ReadFile(filepath.Join(currentDir, tst.project, tst.project+".bal"))
			if err != nil {
				t.Errorf("error reading created bal file, %v", err)
			}
			if diff := cmp.Diff(expectedContent, actualContent); diff != "" {
				t.Errorf("Write (-want, +got)\n%v", diff)
			}
		})
	}
}

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

	"github.com/google/go-cmp/cmp"

	"cellery.io/cellery/components/cli/internal/test"
	"cellery.io/cellery/components/cli/pkg/image"
)

func TestRunPush(t *testing.T) {
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
	tests := []struct {
		name             string
		image            string
		username         string
		password         string
		expectedToPass   bool
		expectedErrorMsg string
		credManager      *test.MockCredManager
	}{
		{
			name:             "push valid image with credentials",
			image:            "myorg/hello:1.0.0",
			username:         "alice",
			password:         "alice123",
			expectedToPass:   true,
			expectedErrorMsg: "",
			credManager:      test.NewMockCredManager(test.SetCredentials("myhub.cellery.io", "aclice", "alice123")),
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			mockCli := test.NewMockCli(test.SetCredManager(tst.credManager),
				test.SetFileSystem(mockFileSystem),
				test.SetRegistry(test.NewMockRegistry()),
				test.SetDockerCli(test.NewMockDockerCli()),
			)
			err := RunPush(mockCli, tst.image, tst.username, tst.password)
			if tst.expectedToPass {
				if err != nil {
					t.Errorf("error in RunPush, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expectedErrorMsg, err.Error()); diff != "" {
					t.Errorf("invalid error message (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

func TestPushImage(t *testing.T) {
	parsedCellImage := &image.CellImage{
		Registry:     "cellery-hub",
		Organization: "myorg",
		ImageName:    "hello",
		ImageVersion: "1.0.0",
	}
	mockRepo, err := ioutil.TempDir("", "mock-repo")
	if err != nil {
		t.Errorf("failed to create mock repository, %v", err)
	}
	if err := os.MkdirAll(filepath.Join(mockRepo, "myorg", "hello", "1.0.0"), os.ModePerm); err != nil {
		t.Errorf("failed to create mock cell image directory, %v", err)
	}
	if _, err = os.Create(filepath.Join(mockRepo, "myorg", "hello", "1.0.0", "hello.zip")); err != nil {
		t.Errorf("failed to create mock cell image zip file, %v", err)
	}
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo))
	err = pushImage(test.NewMockCli(test.SetRegistry(test.NewMockRegistry()), test.SetFileSystem(mockFileSystem)),
		parsedCellImage, "alice", "alice123")
	if err != nil {
		t.Errorf("pullImage err, %v", err)
	}
}

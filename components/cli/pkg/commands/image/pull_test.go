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
	"github.com/cellery-io/sdk/components/cli/pkg/image"
)

func TestRunPull(t *testing.T) {
	mockRepo, err := ioutil.TempDir("", "mock-repo")
	if err != nil {
		t.Errorf("failed to create mock repository, %v", err)
	}
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo))
	mockCredManager := test.NewMockCredManager(test.SetCredentials("myhub.cellery.io", "aclice", "alice123"))

	sampleImage, err := ioutil.ReadFile(filepath.Join("testdata", "repo", "myorg", "hello", "1.0.0", "hello.zip"))
	if err != nil {
		t.Error("error reading sample image file")
	}
	imagesMap := make(map[string][]byte)
	imagesMap["myorg/hello:1.0.0"] = sampleImage
	mockRegistry := test.NewMockRegistry(test.SetImages(imagesMap))
	mockCli := test.NewMockCli(
		test.SetFileSystem(mockFileSystem),
		test.SetCredManager(mockCredManager),
		test.SetRegistry(mockRegistry),
	)
	tests := []struct {
		name             string
		image            string
		silent           bool
		username         string
		password         string
		expectedToPass   bool
		expectedErrorMsg string
	}{
		{
			name:             "pull valid image",
			image:            "myorg/hello:1.0.0",
			silent:           true,
			username:         "alice",
			password:         "alice123",
			expectedToPass:   true,
			expectedErrorMsg: "",
		},
		{
			name:             "pull invalid image",
			image:            "myorg/foo:1.0.0",
			silent:           true,
			username:         "alice",
			password:         "alice123",
			expectedToPass:   false,
			expectedErrorMsg: "invalid cell image, zip: not a valid zip file",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunPull(mockCli, tst.image, tst.silent, tst.username, tst.password)
			if tst.expectedToPass {
				if err != nil {
					t.Errorf("error in RunPull, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expectedErrorMsg, err.Error()); diff != "" {
					t.Errorf("RunLogout: not logged in (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

func TestPullImage(t *testing.T) {
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
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo))
	err = pullImage(test.NewMockCli(test.SetRegistry(test.NewMockRegistry()), test.SetFileSystem(mockFileSystem)),
		parsedCellImage, "alice", "alice123")
	if err != nil {
		t.Errorf("pullImage err, %v", err)
	}
}

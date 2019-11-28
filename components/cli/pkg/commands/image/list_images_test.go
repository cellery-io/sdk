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

	"github.com/google/go-cmp/cmp"

	"cellery.io/cellery/components/cli/internal/test"
)

func TestRunListImages(t *testing.T) {
	mockRepo := filepath.Join("testdata", "repo")
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))

	tests := []struct {
		name string
	}{
		{
			name: "list images",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunListImages(mockCli)
			if err != nil {
				t.Errorf("error in RunListImages, %v", err)
			}
		})
	}
}

func TestGetImagesArray(t *testing.T) {
	mockRepo := filepath.Join("testdata", "repo")
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))

	tests := []struct {
		name     string
		wantName string
		wantKind string
	}{
		{
			name:     "list images with single image",
			wantName: "myorg/employee:1.0.0",
			wantKind: "Cell",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			imagesArray, err := getImagesArray(mockCli)
			if err != nil {
				t.Errorf("error in getImagesArray, %v", err)
			} else {
				if diff := cmp.Diff(testIteration.wantName, imagesArray[0].name); diff != "" {
					t.Errorf("getImagesArray: name (-want, +got)\n%v", diff)
				}
				if diff := cmp.Diff(testIteration.wantKind, imagesArray[0].kind); diff != "" {
					t.Errorf("getImagesArray: kind (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

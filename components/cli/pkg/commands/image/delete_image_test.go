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

package image

import (
	"path/filepath"
	"testing"

	"github.com/cellery-io/sdk/components/cli/internal/test"
)

func TestDeleteImageSuccess(t *testing.T) {
	mockRepo := filepath.Join("testdata", "repo")
	mockFileSystem := test.NewMockFileSystem(test.SetRepository(mockRepo))
	mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))

	tests := []struct {
		name      string
		images    []string
		deleteAll bool
		regex     string
	}{
		{
			name:      "delete single image",
			images:    []string{"employee"},
			deleteAll: false,
			regex:     "",
		},
		{
			name:      "delete multiple images",
			images:    []string{"employee", "stock"},
			deleteAll: false,
			regex:     "",
		},
		{
			name:      "delete all images",
			images:    []string{},
			deleteAll: true,
			regex:     "",
		},
		{
			name:      "delete images with regex",
			images:    []string{"employee", "stock"},
			deleteAll: false,
			regex:     ".*/employee:.*",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunDeleteImage(mockCli, testIteration.images, testIteration.regex, testIteration.deleteAll)
			if err != nil {
				t.Errorf("error in RunDeleteImage, %v", err)
			}
		})
	}
}

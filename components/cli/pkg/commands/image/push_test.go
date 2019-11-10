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

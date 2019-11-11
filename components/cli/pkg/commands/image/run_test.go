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
	"github.com/cellery-io/sdk/components/cli/pkg/ballerina"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
)

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
	var instanceEnvVars []*ballerina.EnvironmentVariable
	rootNodeDependencies := map[string]*dependencyInfo{}
	err = startCellInstance(mockCli, imageDir, "hello", mainNode, instanceEnvVars, false,
		rootNodeDependencies, false)
	if err != nil {
		t.Errorf("startCellInstance failed: %v", err)
	}
}

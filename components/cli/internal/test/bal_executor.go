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

package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/ballerina"
)

type MockBalExecutor struct {
	currentDir          string
	version             string
	executablePath      string
	yamlName            string
	yamlContent         []byte
	metadataJsonContent []byte
}

// NewMockBalExecutor returns a MockBalExecutor instance.
func NewMockBalExecutor(opts ...func(*MockBalExecutor)) *MockBalExecutor {
	balExecutor := &MockBalExecutor{
		executablePath: "",
	}
	for _, opt := range opts {
		opt(balExecutor)
	}
	return balExecutor
}

func SetBalVersion(version string) func(*MockBalExecutor) {
	return func(balExecutor *MockBalExecutor) {
		balExecutor.version = version
	}
}
func SetBalCurrentDir(currentDir string) func(*MockBalExecutor) {
	return func(balExecutor *MockBalExecutor) {
		balExecutor.currentDir = currentDir
	}
}

func SetYamlName(name string) func(*MockBalExecutor) {
	return func(balExecutor *MockBalExecutor) {
		balExecutor.yamlName = name
	}
}

func SetYamlContent(content []byte) func(*MockBalExecutor) {
	return func(balExecutor *MockBalExecutor) {
		balExecutor.yamlContent = content
	}
}

func SetMetadataJsonContent(content []byte) func(*MockBalExecutor) {
	return func(balExecutor *MockBalExecutor) {
		balExecutor.metadataJsonContent = content
	}
}

// Build mocks execution of ballerina build on an executable bal file.
func (balExecutor *MockBalExecutor) Build(fileName string, args []string) error {
	var err error
	var metadataJson, yaml *os.File
	if err := os.MkdirAll(filepath.Join(balExecutor.currentDir, "target", "cellery"), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create target/cellery dir, %v", err)
	}
	if metadataJson, err = os.Create(filepath.Join(balExecutor.currentDir, "target", "cellery", "metadata.json")); err != nil {
		return fmt.Errorf("failed to create metadata.json, %v", err)
	}
	if err := ioutil.WriteFile(metadataJson.Name(), []byte(balExecutor.metadataJsonContent), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write to metadata.json, %v", err)
	}
	if yaml, err = os.Create(filepath.Join(balExecutor.currentDir, "target", "cellery", balExecutor.yamlName)); err != nil {
		return fmt.Errorf("failed to create %s, %v", balExecutor.yamlName, err)
	}
	if err := ioutil.WriteFile(yaml.Name(), []byte(balExecutor.yamlContent), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write to yaml, %v", err)
	}
	return nil
}

// Build mocks execution of ballerina run on an executable bal file.
func (balExecutor *MockBalExecutor) Run(fileName string, args []string, envVars []*ballerina.EnvironmentVariable) error {
	return nil
}

// Build mocks execution of ballerina run for tests on an executable bal file.
func (balExecutor *MockBalExecutor) Test(fileName string, args []string, envVars []*ballerina.EnvironmentVariable) error {
	return nil
}

// Build mocks execution of ballerina run for cellery init on an executable bal file.
func (balExecutor *MockBalExecutor) Init(projectName string) error {
	return nil
}

// Version returns the mock ballerina version.
func (balExecutor *MockBalExecutor) Version() (string, error) {
	return balExecutor.version, nil
}

// ExecutablePath returns mock ballerina executable path.
func (balExecutor *MockBalExecutor) ExecutablePath() (string, error) {
	return balExecutor.executablePath, nil
}

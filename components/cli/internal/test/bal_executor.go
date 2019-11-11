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
	"os"
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/ballerina"
)

type MockBalExecutor struct {
	CurrentDir     string
	version        string
	executablePath string
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

// Build mocks execution of ballerina build on an executable bal file.
func (balExecutor *MockBalExecutor) Build(fileName string, iName []byte) error {
	_, err := os.Create(filepath.Join(balExecutor.CurrentDir, "metadata.json"))
	return err
}

// Build mocks execution of ballerina run on an executable bal file.
func (balExecutor *MockBalExecutor) Run(imageDir string, instanceName string, envVars []*ballerina.EnvironmentVariable, tempRunFileName string, args []string) error {
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

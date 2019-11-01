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

package test

import (
	"os"
	"path/filepath"
)

type MockBalExecutor struct {
	CurrentDir string
}

// Build mocks execution of ballerina build on an executable bal file.
func (baleExecutor *MockBalExecutor) Build(fileName string, iName []byte) error {
	_, err := os.Create(filepath.Join(baleExecutor.CurrentDir, "metadata.json"))
	return err
}

// Version returns the mock ballerina version.
func (baleExecutor *MockBalExecutor) Version() (string, error) {
	return "", nil
}

// ExecutablePath returns mock ballerina executable path.
func (baleExecutor *MockBalExecutor) ExecutablePath() (string, error) {
	return "", nil
}

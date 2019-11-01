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
	"bytes"
	"io"

	"github.com/cellery-io/sdk/components/cli/ballerina"
	"github.com/cellery-io/sdk/components/cli/cli"
)

type MockCli struct {
	out               io.Writer
	outBuffer         *bytes.Buffer
	BallerinaExecutor ballerina.BalExecutor
}

// NewMockCli returns a mock cli for the cli.Cli interface.
func NewMockCli() *MockCli {
	outBuffer := new(bytes.Buffer)
	mockCli := &MockCli{
		out:       outBuffer,
		outBuffer: outBuffer,
	}
	return mockCli
}

// Out returns the output stream (stdout) the cli should write on.
func (cli *MockCli) Out() io.Writer {
	return cli.out
}

// OutBuffer returns the stdout buffer.
func (cli *MockCli) OutBuffer() *bytes.Buffer {
	return cli.outBuffer
}

// ExecuteTask mocks function execution.
func (cli *MockCli) ExecuteTask(startMessage, errorMessage, successMessage string, function func() error) error {
	err := function()
	if err != nil {
		return err
	}
	return nil
}

// FileSystem returns a mock FileSystemManager instance.
func (cli *MockCli) FileSystem() cli.FileSystemManager {
	return NewMockFileSystem()
}

// BalExecutor returns a mock ballerina.BalExecutor instance.
func (cli *MockCli) BalExecutor() ballerina.BalExecutor {
	return cli.BallerinaExecutor
}

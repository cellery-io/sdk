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
)

type MockCli struct {
	out       io.Writer
	outBuffer *bytes.Buffer
}

// NewMockCli returns a mock cli for the cli.Cli interface.
func NewMockCli() *MockCli {
	outBuffer := new(bytes.Buffer)
	cli := &MockCli{
		out:       outBuffer,
		outBuffer: outBuffer,
	}
	return cli
}

// Out returns the output stream (stdout) the cli should write on.
func (cli *MockCli) Out() io.Writer {
	return cli.out
}

// OutBuffer returns the stdout buffer.
func (cli *MockCli) OutBuffer() *bytes.Buffer {
	return cli.outBuffer
}

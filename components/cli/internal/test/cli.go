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
	"bytes"
	"io"
	"time"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/ballerina"
	"cellery.io/cellery/components/cli/pkg/docker"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/registry"
	"cellery.io/cellery/components/cli/pkg/registry/credentials"
	"cellery.io/cellery/components/cli/pkg/runtime"
)

type MockCli struct {
	out               io.Writer
	outBuffer         *bytes.Buffer
	ballerinaExecutor ballerina.BalExecutor
	kubeCli           kubernetes.KubeCli
	registry          registry.Registry
	manager           cli.FileSystemManager
	docker            docker.Docker
	credManager       credentials.CredManager
	credReader        credentials.CredReader
	runtime           runtime.Runtime
	actionItem        int
	selection         cli.Selection
}

// NewMockCli returns a mock cli for the cli.Cli interface.
func NewMockCli(opts ...func(*MockCli)) *MockCli {
	outBuffer := new(bytes.Buffer)
	mockCli := &MockCli{
		out:       outBuffer,
		outBuffer: outBuffer,
	}
	for _, opt := range opts {
		opt(mockCli)
	}
	return mockCli
}

func SetKubeCli(kubecli kubernetes.KubeCli) func(*MockCli) {
	return func(cli *MockCli) {
		cli.kubeCli = kubecli
	}
}

func SetRegistry(registry registry.Registry) func(*MockCli) {
	return func(cli *MockCli) {
		cli.registry = registry
	}
}

func SetFileSystem(manager cli.FileSystemManager) func(*MockCli) {
	return func(cli *MockCli) {
		cli.manager = manager
	}
}

func SetDockerCli(docker docker.Docker) func(*MockCli) {
	return func(cli *MockCli) {
		cli.docker = docker
	}
}

func SetBalExecutor(balExecutor ballerina.BalExecutor) func(*MockCli) {
	return func(cli *MockCli) {
		cli.ballerinaExecutor = balExecutor
	}
}

func SetCredManager(manager credentials.CredManager) func(*MockCli) {
	return func(cli *MockCli) {
		cli.credManager = manager
	}
}

func SetCredReader(reader credentials.CredReader) func(*MockCli) {
	return func(cli *MockCli) {
		cli.credReader = reader
	}
}

func SetActionItem(actionItem int) func(*MockCli) {
	return func(cli *MockCli) {
		cli.actionItem = actionItem
	}
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
	return cli.manager
}

// BalExecutor returns a mock ballerina.BalExecutor instance.
func (cli *MockCli) BalExecutor() ballerina.BalExecutor {
	return cli.ballerinaExecutor
}

// KubeCli returns mock kubernetes.KubeCli instance.
func (cli *MockCli) KubeCli() kubernetes.KubeCli {
	return cli.kubeCli
}

// KubeCli returns kubernetes.KubeCli instance.
func (cli *MockCli) Registry() registry.Registry {
	return cli.registry
}

// OpenBrowser mocks opening up of the provided URL in a browser
func (cli *MockCli) OpenBrowser(url string) error {
	return nil
}

// FileSystem returns mock DockerCli instance.
func (cli *MockCli) DockerCli() docker.Docker {
	return cli.docker
}

// CredManager returns a mock CredManager instance.
func (cli *MockCli) CredManager() credentials.CredManager {
	return cli.credManager
}

// CredReader returns a CredReader instance.
func (cli *MockCli) CredReader() credentials.CredReader {
	return cli.credReader
}

// Runtime returns a Runtime instance.
func (cli *MockCli) Runtime() runtime.Runtime {
	return cli.runtime
}

func SetRuntime(runtime runtime.Runtime) func(*MockCli) {
	return func(cli *MockCli) {
		cli.runtime = runtime
	}
}

func (cli *MockCli) Sleep(seconds time.Duration) {
}

func (cli *MockCli) ExecuteUserSelection(prompt string, options []cli.Selection) error {
	for _, opt := range options {
		if cli.actionItem == opt.Number {
			cli.selection = opt
		}
	}
	return nil
}

func (cli *MockCli) Selection() cli.Selection {
	return cli.selection
}

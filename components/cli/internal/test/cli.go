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
	"github.com/cellery-io/sdk/components/cli/docker"
	"github.com/cellery-io/sdk/components/cli/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/registry"
)

type MockCli struct {
	out               io.Writer
	outBuffer         *bytes.Buffer
	ballerinaExecutor ballerina.BalExecutor
	kubeCli           kubernetes.KubeCli
	registry          registry.Registry
	manager           cli.FileSystemManager
	docker            docker.Docker
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

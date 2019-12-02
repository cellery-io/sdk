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

package cli

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"cellery.io/cellery/components/cli/pkg/ballerina"
	"cellery.io/cellery/components/cli/pkg/docker"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/registry"
	"cellery.io/cellery/components/cli/pkg/registry/credentials"
	cliRuntime "cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
)

// Cli represents the cellery command line client.
type Cli interface {
	Out() io.Writer
	ExecuteTask(startMessage, errorMessage, successMessage string, function func() error) error
	FileSystem() FileSystemManager
	BalExecutor() ballerina.BalExecutor
	KubeCli() kubernetes.KubeCli
	Registry() registry.Registry
	OpenBrowser(url string) error
	DockerCli() docker.Docker
	CredManager() credentials.CredManager
	CredReader() credentials.CredReader
	Runtime() cliRuntime.Runtime
}

// CelleryCli is an instance of the cellery command line client.
// Instances of the client can be returned from NewCelleryCli.
type CelleryCli struct {
	fileSystemManager FileSystemManager
	kubecli           kubernetes.KubeCli
	ballerinaExecutor ballerina.BalExecutor
	registry          registry.Registry
	docker            docker.Docker
	credManager       credentials.CredManager
	credReader        credentials.CredReader
	runtime           cliRuntime.Runtime
}

// NewCelleryCli returns a CelleryCli instance.
func NewCelleryCli(opts ...func(*CelleryCli)) *CelleryCli {
	cli := &CelleryCli{
		kubecli: kubernetes.NewCelleryKubeCli(),
		docker:  docker.NewCelleryDockerCli(),
	}
	for _, opt := range opts {
		opt(cli)
	}
	return cli
}

func SetRegistry(registry registry.Registry) func(*CelleryCli) {
	return func(cli *CelleryCli) {
		cli.registry = registry
	}
}

func SetFileSystem(manager FileSystemManager) func(*CelleryCli) {
	return func(cli *CelleryCli) {
		cli.fileSystemManager = manager
	}
}

func SetBallerinaExecutor(balExecutor ballerina.BalExecutor) func(*CelleryCli) {
	return func(cli *CelleryCli) {
		cli.ballerinaExecutor = balExecutor
	}
}

func SetCredManager(credManager credentials.CredManager) func(*CelleryCli) {
	return func(cli *CelleryCli) {
		cli.credManager = credManager
	}
}

func SetCredReader(credReader credentials.CredReader) func(*CelleryCli) {
	return func(cli *CelleryCli) {
		cli.credReader = credReader
	}
}

// Out returns the writer used for the stdout.
func (cli *CelleryCli) Out() io.Writer {
	return os.Stdout
}

// ExecuteTask executes a function.
// It starts a spinner upon starting function execution.
// Spinner exits with a success message (optional) if the function execution was successful.
// Spinner exists with an error message (optional) if the function execution failed.
func (cli *CelleryCli) ExecuteTask(startMessage, errorMessage, successMessage string, function func() error) error {
	spinner := util.StartNewSpinner(startMessage)
	err := function()
	if err != nil {
		spinner.Stop(false)
		if errorMessage != "" {
			fmt.Fprintln(cli.Out(), errorMessage)
		}
		return err
	}
	spinner.Stop(true)
	if successMessage != "" {
		fmt.Fprintln(cli.Out(), successMessage)
	}
	return nil
}

// FileSystem returns a FileSystemManager instance.
func (cli *CelleryCli) FileSystem() FileSystemManager {
	return cli.fileSystemManager
}

// BalExecutor returns a BalExecutor instance.
func (cli *CelleryCli) BalExecutor() ballerina.BalExecutor {
	return cli.ballerinaExecutor
}

// KubeCli returns a KubeCli instance.
func (cli *CelleryCli) KubeCli() kubernetes.KubeCli {
	return cli.kubecli
}

// Registry returns a registry instance.
func (cli *CelleryCli) Registry() registry.Registry {
	return cli.registry
}

// FileSystem returns FileSystemManager instance.
func (cli *CelleryCli) DockerCli() docker.Docker {
	return cli.docker
}

// CredManager returns a CredManager instance.
func (cli *CelleryCli) CredManager() credentials.CredManager {
	return cli.credManager
}

// CredReader returns a CredReader instance.
func (cli *CelleryCli) CredReader() credentials.CredReader {
	return cli.credReader
}

// OpenBrowser opens up the provided URL in a browser
func (cli *CelleryCli) OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "openbsd":
		fallthrough
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		r := strings.NewReplacer("&", "^&")
		cmd = exec.Command("cmd", "/c", "start", r.Replace(url))
	}
	if cmd != nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			log.Printf("Failed to open browser due to error %v", err)
			return fmt.Errorf("Failed to open browser: " + err.Error())
		}
		err = cmd.Wait()
		if err != nil {
			log.Printf("Failed to wait for open browser command to finish due to error %v", err)
			return fmt.Errorf("Failed to wait for open browser command to finish: " + err.Error())
		}
		return nil
	} else {
		return errors.New(fmt.Sprintf("unsupported platform: %v", runtime.GOOS))
	}
}

// Runtime returns a Runtime instance.
func (cli *CelleryCli) Runtime() cliRuntime.Runtime {
	return cli.runtime
}

func SetRuntime(runtime cliRuntime.Runtime) func(*CelleryCli) {
	return func(cli *CelleryCli) {
		cli.runtime = runtime
	}
}

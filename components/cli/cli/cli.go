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

package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/cellery-io/sdk/components/cli/ballerina"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// Cli represents the cellery command line client.
type Cli interface {
	Out() io.Writer
	ExecuteTask(startMessage, errorMessage, successMessage string, function func() error) error
	FileSystem() FileSystemManager
	BalExecutor() ballerina.BalExecutor
}

// CelleryCli is an instance of the cellery command line client.
// Instances of the client can be returned from NewCelleryCli.
type CelleryCli struct {
	fileSystemManager FileSystemManager
	BallerinaExecutor ballerina.BalExecutor
}

// NewCelleryCli returns a CelleryCli instance.
func NewCelleryCli() *CelleryCli {
	cli := &CelleryCli{
		fileSystemManager: NewCelleryFileSystem(),
	}
	return cli
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
			fmt.Println(errorMessage)
		}
		return err
	}
	spinner.Stop(true)
	if successMessage != "" {
		fmt.Println(successMessage)
	}
	return nil
}

// FileSystem returns FileSystemManager instance.
func (cli *CelleryCli) FileSystem() FileSystemManager {
	return cli.fileSystemManager
}

// BalExecutor returns ballerina.BalExecutor instance.
func (cli *CelleryCli) BalExecutor() ballerina.BalExecutor {
	return cli.BallerinaExecutor
}

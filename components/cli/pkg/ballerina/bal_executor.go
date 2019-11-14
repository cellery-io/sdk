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

package ballerina

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

const ballerina = "ballerina"
const celleryImageDirEnvVar = "CELLERY_IMAGE_DIR"

type BalExecutor interface {
	Build(fileName string, args []string) error
	Run(fileName string, args []string, envVars []*EnvironmentVariable) error
	Test(fileName string, args []string, envVars []*EnvironmentVariable) error
	Init(projectName string) error
	Version() (string, error)
	ExecutablePath() (string, error)
}

type LocalBalExecutor struct {
}

// NewLocalBalExecutor returns a LocalBalExecutor instance.
func NewLocalBalExecutor() *LocalBalExecutor {
	balExecutor := &LocalBalExecutor{}
	return balExecutor
}

// Build executes ballerina build on an executable bal file.
func (balExecutor *LocalBalExecutor) Build(fileName string, args []string) error {
	exePath, err := balExecutor.ExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path, %v", err)
	}
	cmd := exec.Command(exePath, "run", fileName, "build")
	cmd.Args = append(cmd.Args, args...)
	cmd.Args = append(cmd.Args, "{}", "false", "false")
	var stderr bytes.Buffer
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Printf("\r\x1b[2K\033[36m%s\033[m\n", stdoutScanner.Text())
		}
	}()
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			fmt.Printf("\r\x1b[2K\033[36m%s\033[m\n", stderrScanner.Text())
			fmt.Fprintf(&stderr, stderrScanner.Text())
		}
	}()
	err = cmd.Start()
	if err != nil {
		errStr := string(stderr.Bytes())
		return fmt.Errorf("error occurred while starting to build image, %v", errStr)
	}
	err = cmd.Wait()
	if err != nil {
		errStr := string(stderr.Bytes())
		return fmt.Errorf("error occurred while waiting to build image, %v", errStr)
	}
	return nil
}

// Run executes ballerina run on an executable bal file.
func (balExecutor *LocalBalExecutor) Run(fileName string, args []string,
	envVars []*EnvironmentVariable) error {
	cmd := &exec.Cmd{}
	exePath, err := balExecutor.ExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path, %v", err)
	}
	cmd = exec.Command(exePath, "run", fileName, "run")
	cmd.Args = append(cmd.Args, args...)
	cmd.Env = os.Environ()
	//cmd.Env = append(cmd.Env, celleryImageDirEnvVar+"="+imageDir)
	// Export environment variables defined by user
	for _, envVar := range envVars {
		cmd.Env = append(cmd.Env, envVar.Key+"="+envVar.Value)
	}
	var stderr bytes.Buffer
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Printf("\r\x1b[2K\033[36m%s\033[m\n", stdoutScanner.Text())
		}
	}()
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			fmt.Printf("\r\x1b[2K\033[36m%s\033[m\n", stderrScanner.Text())
			fmt.Fprintf(&stderr, stderrScanner.Text())
		}
	}()
	err = cmd.Start()
	if err != nil {
		errStr := string(stderr.Bytes())
		return fmt.Errorf("failed starting to execute run method %v", errStr)
	}
	err = cmd.Wait()
	if err != nil {
		errStr := string(stderr.Bytes())
		return fmt.Errorf("failed waiting to execute run method %v", errStr)
	}
	return nil
}

// Init initializes a ballerina project in the current working directory
func (balExecutor *LocalBalExecutor) Init(projectDir string) error {
	balProjectName := filepath.Base(projectDir) + constants.BalProjExt
	cmd := exec.Command(ballerina, "new", balProjectName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error occurred while initializing ballerina project for tests %v", err)
	}
	return nil
}

// Version returns the ballerina version.
func (balExecutor *LocalBalExecutor) Version() (string, error) {
	version := ""
	cmd := exec.Command(ballerina, "version")
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			version += stdoutScanner.Text()
		}
	}()
	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("error starting to get ballerina version, %v", err)
	}
	err = cmd.Wait()
	if err != nil {
		return "", fmt.Errorf("error waiting to get ballerina version, %v", err)
	}
	return version, nil
}

// ExecutablePath returns the ballerina executable path.
func (balExecutor *LocalBalExecutor) ExecutablePath() (string, error) {
	var err error
	var ballerinaVersion string
	if ballerinaVersion, err = balExecutor.Version(); err != nil {
		return ballerinaInstallationPath()
	}
	if strings.Contains(ballerinaVersion, "Ballerina") {
		if len(strings.Split(ballerinaVersion, " ")) > 0 {
			if strings.Split(ballerinaVersion, " ")[1] == constants.BallerinaVersion {
				// If existing ballerina version is as the expected version, execute ballerina run without executable path
				return ballerina, nil
			}
		}
	}
	return ballerinaInstallationPath()
}

// ballerinaInstallationPath checks if the expected ballerina version exists.
// If so return its installation path.
func ballerinaInstallationPath() (string, error) {
	exePath := util.BallerinaInstallationDir() + constants.BallerinaExecutablePath
	if _, err := os.Stat(exePath); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		} else {
			return "", err
		}
	}
	return exePath + ballerina, nil
}

func (balExecutor *LocalBalExecutor) Test(fileName string, args []string, envVars []*EnvironmentVariable) error {
	cmd := &exec.Cmd{}
	exePath, err := balExecutor.ExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path, %v", err)
	}
	cmd.Env = os.Environ()
	// Export environment variables defined by user
	for _, envVar := range envVars {
		cmd.Env = append(cmd.Env, envVar.Key+"="+envVar.Value)
	}
	var debug bool
	for _, envVar := range envVars {
		if envVar.Key == "DEBUG" {
			debug, err = strconv.ParseBool(envVar.Value)
			if err != nil {
				return err
			}
		}
	}
	if !debug {
		balArgs := []string{exePath, "test", "--all"}
		args = append(args, "--run")
		args = append(args, balArgs...)
	}
	telepresenceExecPath := filepath.Join(util.CelleryInstallationDir(), constants.TelepresenceExecPath, "/telepresence")
	cmd.Path = telepresenceExecPath
	cmd.Args = args
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error occurred while running tests, %v", err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error occurred while waiting for tests to complete, %v", err)
	}
	return nil
}

// EnvironmentVariable is used to store the environment variables to be passed to the instances
type EnvironmentVariable struct {
	Key   string
	Value string
}

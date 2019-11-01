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

package ballerina

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

const homeCellery = "/home/cellery"

type DockerBalExecutor struct {
}

// NewDockerBalExecutor returns a DockerBalExecutor instance.
func NewDockerBalExecutor() *DockerBalExecutor {
	balExecutor := &DockerBalExecutor{}
	return balExecutor
}

// Build executes ballerina build when ballerina is not installed.
func (baleExecutor *DockerBalExecutor) Build(fileName string, iName []byte) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error in determining working directory, %v", err)
	}
	// Retrieve the cellery cli docker instance status.
	cmdDockerPs := exec.Command("docker", "ps", "--filter",
		"label=ballerina-runtime="+version.BuildVersion(),
		"--filter", "label=currentDir="+currentDir, "--filter", "status=running", "--format", "{{.ID}}")
	containerId, err := cmdDockerPs.Output()
	if err != nil {
		return fmt.Errorf("error in retrieving cellery cli docker instance status, %v", err)
	}

	if string(containerId) == "" {
		cmdDockerRun := exec.Command("docker", "run", "-d",
			"-l", "ballerina-runtime="+version.BuildVersion(),
			"-l", "current.dir="+currentDir,
			"--mount", "type=bind,source="+currentDir+",target="+homeCellery+"/src",
			"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".ballerina,target="+homeCellery+"/.ballerina",
			"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".cellery,target="+homeCellery+"/.cellery",
			"wso2cellery/ballerina-runtime:"+version.BuildVersion(), "sleep", "600",
		)
		stderrReader, err := cmdDockerRun.StderrPipe()
		if err != nil {
			return fmt.Errorf("error while building stderr pipe, %v", err)
		}
		stdoutReader, _ := cmdDockerRun.StdoutPipe()
		if err != nil {
			return fmt.Errorf("error while building stdout pipe, %v", err)
		}
		stderrScanner := bufio.NewScanner(stderrReader)
		stdoutScanner := bufio.NewScanner(stdoutReader)
		err = cmdDockerRun.Start()
		if err != nil {
			return fmt.Errorf("error while starting docker process, %v", err)
		}
		go func() {
			for {
				if stderrScanner.Scan() && strings.HasPrefix(stderrScanner.Text(), "Unable to find image") {
					break
				}
			}
		}()
		go func() {
			for {
				if stdoutScanner.Scan() {
					containerId = []byte(stdoutScanner.Text())
				}
			}
		}()

		err = cmdDockerRun.Wait()
		if err != nil {
			return fmt.Errorf("error while running ballerina-runtime docker image, %v", err)
		}
		time.Sleep(5 * time.Second)
	}

	cliUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("error while retrieving the current user, %v", err)
	}
	if cliUser.Uid != constants.CELLERY_DOCKER_CLI_USER_ID {
		cmdUserExist := exec.Command("docker", "exec", strings.TrimSpace(string(containerId)),
			"id", "-u", cliUser.Username)
		_, errUserExist := cmdUserExist.Output()
		if errUserExist != nil {
			cmdUserAdd := exec.Command("docker", "exec", strings.TrimSpace(string(containerId)), "useradd", "-m",
				"-d", homeCellery, "--uid", cliUser.Uid, cliUser.Username)

			_, errUserAdd := cmdUserAdd.Output()
			if errUserAdd != nil {
				return fmt.Errorf("error in adding Cellery execution user, %v", err)
			}
		}
	}
	re := regexp.MustCompile("^" + currentDir + "/")
	balFilePath := re.ReplaceAllString(fileName, "")
	cmd := exec.Command("docker", "exec", "-w", homeCellery+"/src", "-u", cliUser.Uid,
		strings.TrimSpace(string(containerId)), constants.DOCKER_CLI_BALLERINA_EXECUTABLE_PATH, "run", balFilePath, "build", string(iName), "{}")
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error occurred while starting to build image, %v", err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error occurred while waiting to build image, %v", err)
	}
	return nil
}

// Version returns the ballerina version.
func (baleExecutor *DockerBalExecutor) Version() (string, error) {
	return "", nil
}

// ExecutablePath returns ballerina executable path.
func (baleExecutor *DockerBalExecutor) ExecutablePath() (string, error) {
	return "", nil
}

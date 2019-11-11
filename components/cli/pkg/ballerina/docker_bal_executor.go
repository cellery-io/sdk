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
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

const homeCellery = "/home/cellery"
const dockerCliUserId = "1000"
const dockerCliBallerinaExecutablePath = "/usr/lib/ballerina/ballerina-1.0.3/bin/ballerina"
const dockerCliCellImageDir = "/home/cellery/.cellery/tmp/cellery-cell-image"

type DockerBalExecutor struct {
}

// NewDockerBalExecutor returns a DockerBalExecutor instance.
func NewDockerBalExecutor() *DockerBalExecutor {
	balExecutor := &DockerBalExecutor{}
	return balExecutor
}

// Build executes ballerina build when ballerina is not installed.
func (balExecutor *DockerBalExecutor) Build(fileName string, iName []byte) error {
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
	if cliUser.Uid != dockerCliUserId {
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
		strings.TrimSpace(string(containerId)), dockerCliBallerinaExecutablePath, "run",
		filepath.Join(homeCellery, "src", balFilePath), "build", string(iName), "{}", "false", "false")
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

// Run executes ballerina run when ballerina is not installed.
func (balExecutor *DockerBalExecutor) Run(imageDir string, instanceName string,
	envVars []*EnvironmentVariable, tempRunFileName string, args []string) error {

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error in determining working directory, %v", err)
	}
	//Retrieve the cellery cli docker instance status.
	cmdDockerPs := exec.Command("docker", "ps", "--filter", "label=ballerina-runtime="+version.BuildVersion(),
		"--filter", "label=currentDir="+currentDir, "--filter", "status=running", "--format", "{{.ID}}")

	containerId, err := cmdDockerPs.Output()
	if err != nil {
		return fmt.Errorf("docker Run Error, %v", err)
	}

	if string(containerId) == "" {
		cmdDockerRun := exec.Command("docker", "run", "-d", "-l", "ballerina-runtime="+version.BuildVersion(),
			"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".ballerina,target=/home/cellery/.ballerina",
			"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".cellery,target=/home/cellery/.cellery",
			"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".kube,target=/home/cellery/.kube",
			"wso2cellery/ballerina-runtime:"+version.BuildVersion(), "sleep", "600",
		)
		containerId, err = cmdDockerRun.Output()
		if err != nil {
			return fmt.Errorf("docker Run Error, %v", err)
		}
		time.Sleep(5 * time.Second)
	}
	cliUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("error while retrieving the current user, %v", err)
	}
	exeUid := dockerCliUserId

	if cliUser.Uid != dockerCliUserId && runtime.GOOS == "linux" {
		cmdUserExist := exec.Command("docker", "exec", strings.TrimSpace(string(containerId)),
			"id", "-u", cliUser.Username)
		_, errUserExist := cmdUserExist.Output()
		if errUserExist != nil {
			cmdUserAdd := exec.Command("docker", "exec", strings.TrimSpace(string(containerId)), "useradd", "-m",
				"-d", "/home/cellery", "--uid", cliUser.Uid, cliUser.Username)

			_, errUserAdd := cmdUserAdd.Output()
			if errUserAdd != nil {
				return fmt.Errorf("error in adding Cellery execution user, %v", err)
			}
		}
		exeUid = cliUser.Uid
	}
	var cmdArgs []string
	cmdArgs = append(cmdArgs, "-e", celleryImageDirEnvVar+"="+imageDir)

	re := regexp.MustCompile(`^.*cellery-cell-image`)
	tempRunFileName = re.ReplaceAllString(tempRunFileName, dockerCliCellImageDir)
	dockerImageDir := re.ReplaceAllString(imageDir, dockerCliCellImageDir)

	cmd := exec.Command("docker", "exec", "-e", celleryImageDirEnvVar+"="+dockerImageDir)
	shellEnvs := os.Environ()
	// check if any env var prepended with `CELLERY` exists. If so, set them to docker exec command.
	if len(shellEnvs) != 0 {
		for _, shellEnv := range shellEnvs {
			if strings.HasPrefix(shellEnv, "CELLERY") {
				cmd.Args = append(cmd.Args, "-e", shellEnv)
			}
		}
	}
	// set any explicitly passed env vars in cellery run command to the docker exec.
	// This will override any env vars with identical names (prefixed with 'CELLERY') set previously.
	if len(envVars) != 0 {
		for _, envVar := range envVars {
			cmd.Args = append(cmd.Args, "-e", envVar.Key+"="+envVar.Value)
		}
	}
	cmd.Args = append(cmd.Args, "-w", "/home/cellery", "-u", exeUid,
		strings.TrimSpace(string(containerId)), dockerCliBallerinaExecutablePath)
	cmd.Args = append(cmd.Args, "run", tempRunFileName, "run")
	cmd.Args = append(cmd.Args, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, celleryImageDirEnvVar+"="+imageDir)
	// Export environment variables defined by user
	// Todo: check if setting env vars to cmd.Env is redundant
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

// Version returns the ballerina version.
func (balExecutor *DockerBalExecutor) Version() (string, error) {
	return "", nil
}

// ExecutablePath returns ballerina executable path.
func (balExecutor *DockerBalExecutor) ExecutablePath() (string, error) {
	return "", nil
}

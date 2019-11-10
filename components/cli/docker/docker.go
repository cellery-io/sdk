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

package docker

import (
	"fmt"
	"os/exec"
	"strings"
)

const docker = "docker"

type Docker interface {
	ServerVersion() (string, error)
	ClientVersion() (string, error)
}

type CelleryDockerCli struct {
}

func NewCelleryDockerCli() *CelleryDockerCli {
	cli := &CelleryDockerCli{}
	return cli
}

func (cli *CelleryDockerCli) ServerVersion() (string, error) {
	cmd := exec.Command(docker,
		"version",
		"--format",
		"{{.Server.Version}}",
	)
	dockerServerResult, err := cmd.CombinedOutput()
	if err != nil {
		return string(dockerServerResult), fmt.Errorf("error while getting Docker Server version, %v", strings.TrimSpace(string(dockerServerResult)))
	}
	return string(dockerServerResult), nil
}

func (cli *CelleryDockerCli) ClientVersion() (string, error) {
	cmd := exec.Command(docker,
		"version",
		"--format",
		"{{.Client.Version}}",
	)
	dockerClientResult, err := cmd.CombinedOutput()
	if err != nil {
		return string(dockerClientResult), fmt.Errorf("error while getting Docker Client version, %v", strings.TrimSpace(string(dockerClientResult)))
	}
	return string(dockerClientResult), nil
}

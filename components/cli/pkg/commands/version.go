/*
 * Copyright (c) 2018 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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

package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/version"

	"github.com/fatih/color"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunVersion() {

	type K8SStruct struct {
		ServerVersion struct {
			GitVersion string
		}
	}

	// Creating functions for printing version details
	white := color.New(color.FgWhite)
	boldWhite := white.Add(color.Bold)

	// Printing Cellery version information
	_, _ = boldWhite.Println("Cellery:")
	fmt.Printf(" CLI Version:\t\t%s\n", version.BuildVersion())
	fmt.Printf(" OS/Arch:\t\t%s/%s\n", runtime.GOOS, runtime.GOARCH)
	//fmt.Println(" Experimental:\t\ttrue") // TODO

	// Printing Ballerina version information
	_, _ = boldWhite.Println("\nBallerina:")
	moduleMgr := &util.BLangManager{}
	exePath, err := moduleMgr.GetExecutablePath()
	if err != nil {
		util.ExitWithErrorMessage("Failed to get executable path", err)
	}
	balVersionCmd := exec.Command(exePath+"ballerina", "version")
	balResult, err := balVersionCmd.Output()
	if err != nil {
		// Having b7a locally is optional since the cellery build and run can be invoked via a docker container.
		fmt.Println(" Ballerina not found locally")
	} else {
		balVersion := string(balResult)
		fmt.Println(" Version:\t\t" + strings.TrimSpace(strings.Split(balVersion, " ")[1]))
	}

	// Printing Kubernetes version information
	_, _ = boldWhite.Println("\nKubernetes")
	k8sCmd := exec.Command("kubectl", "version", "-o", "json")
	var k8sStdout, k8sStderr bytes.Buffer
	k8sCmd.Stdout = &k8sStdout
	k8sCmd.Stderr = &k8sStderr
	k8sExecErr := k8sCmd.Run()
	if k8sExecErr != nil {
		util.ExitWithErrorMessage("\x1b[31;1m Error while getting Kubernetes version, \x1b[0m%v\n",
			errors.New(strings.TrimSpace(k8sStderr.String())))
	} else {
		k8sVersion := k8sStdout.String()
		k8sVersionJson := K8SStruct{}
		jsonErr := json.Unmarshal([]byte(k8sVersion), &k8sVersionJson)
		if jsonErr != nil {
			fmt.Printf("\x1b[31;1m Cannot connect to Kubernetes Cluster \x1b[0m %v \n", k8sExecErr)
		}
		fmt.Println(" Version:\t\t" + k8sVersionJson.ServerVersion.GitVersion)
		fmt.Println(" CRD:\t\t\ttrue") // TODO
	}

	// Printing Docker version information
	_, _ = boldWhite.Println("\nDocker:")
	dockerServerCmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	dockerServerResult, err := dockerServerCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("\x1b[31;1m Error while getting Docker Server version, \x1b[0m%v\n",
			strings.TrimSpace(string(dockerServerResult)))
		return
	} else {
		dockerServerVersion := string(dockerServerResult)
		fmt.Println(" Server Version:\t" + strings.TrimSpace(dockerServerVersion))
	}
	dockerClientCmd := exec.Command("docker", "version", "--format", "{{.Client.Version}}")
	dockerClientResult, err := dockerClientCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("\x1b[31;1m Error while getting Docker Client version, \x1b[0m%v\n",
			strings.TrimSpace(string(dockerServerResult)))
	} else {
		dockerClientVersion := string(dockerClientResult)
		fmt.Println(" Client Version:\t" + strings.TrimSpace(dockerClientVersion))
	}
}

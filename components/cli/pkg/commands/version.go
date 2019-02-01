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
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

func RunVersion() error {

	type Version struct {
		CelleryTool struct {
			Version          string `json:"version"`
			APIVersion       string `json:"apiVersion"`
			BallerinaVersion string `json:"ballerinaVersion"`
			GitCommit        string `json:"gitCommit"`
			Built            string `json:"built"`
			OsArch           string `json:"osArch"`
			Experimental     bool   `json:"experimental"`
		} `json:"celleryTool"`
		CelleryRepository struct {
			Server        string `json:"server"`
			APIVersion    string `json:"apiVersion"`
			Authenticated bool   `json:"authenticated"`
		} `json:"celleryRepository"`
		Kubernetes struct {
			Version string `json:"version"`
			Crd     string `json:"crd"`
		} `json:"kubernetes"`
		Docker struct {
			Registry string `json:"registry"`
		} `json:"docker"`
	}

	type K8SStruct struct {
		ServerVersion struct {
			GitVersion string
		}
	}

	//// Call API server to get the version details.
	//resp, err := http.Get(constants.BASE_API_URL + "/version")
	//if err != nil {
	//	fmt.Println("Error connecting to the cellery api-server.")
	//}
	//defer resp.Body.Close()
	//
	//// Read the response body and get to an bute array.
	//body, err := ioutil.ReadAll(resp.Body)
	//response := Version{}
	//
	//// set the resopose byte array to the version struct.
	//json.Unmarshal(body, &response)

	balVersionCmd := exec.Command("ballerina", "version")
	balResult, err := balVersionCmd.Output()
	if err != nil {
		fmt.Printf("\x1b[31;1m Ballerina not found. You must have ballerina installed. \x1b[0m\n")
	}
	balVersion := string(balResult)

	//Print the version details.
	white := color.New(color.FgWhite)
	boldWhite := white.Add(color.Bold)

	boldWhite.Println("Cellery:")
	fmt.Println(" Version:\t\t0.0.1")                    //TODO
	fmt.Println(" API version:\t\t0.0.1")                //TODO
	fmt.Println(" Git commit:\t\t78a6bdb")               //TODO
	fmt.Println(" Built:\t\t\tTue Oct 23 22:41:53 2018") //TODO
	fmt.Println(" OS/Arch:\t\t" + runtime.GOOS + "/" + runtime.GOARCH)
	fmt.Println(" Experimental:\t\ttrue") //TODO

	boldWhite.Println("\nBallerina:")
	fmt.Println(" Version:\t\t" + strings.TrimSpace(strings.Split(balVersion, " ")[1]))

	boldWhite.Println("\nCellery Repository:")
	fmt.Println(" Server:\t\thttp://central.cellery.io/api") //TODO
	fmt.Println(" API version:\t\t0.0.1")                    //TODO
	fmt.Println(" Login:\t\t\tLogged in")                    //TODO

	boldWhite.Println("\nKubernetes")
	k8sCmd := exec.Command("kubectl", "version", "-o", "json")
	var k8sStdout, k8sStderr bytes.Buffer
	k8sCmd.Stdout = &k8sStdout
	k8sCmd.Stderr = &k8sStderr
	k8sExecErr := k8sCmd.Run()
	if k8sExecErr != nil {
		fmt.Printf("\x1b[31;1m Error while getting Kubernetes version, \x1b[0m%v\n",
			strings.TrimSpace(k8sStderr.String()))
	} else {
		k8sVersion := k8sStdout.String()
		k8sVersionJson := K8SStruct{}
		jsonErr := json.Unmarshal([]byte(k8sVersion), &k8sVersionJson)
		if jsonErr != nil {
			fmt.Printf("\x1b[31;1m Cannot connect to Kubernetes Cluster \x1b[0m %v \n", k8sExecErr)
		}
		fmt.Println(" Version:\t\t" + k8sVersionJson.ServerVersion.GitVersion)
		fmt.Println(" CRD:\t\t\ttrue") //TODO
	}

	boldWhite.Println("\nDocker:")
	dockerServerCmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	dockerServerResult, err := dockerServerCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("\x1b[31;1m Error while getting Docker Server version, \x1b[0m%v\n",
			strings.TrimSpace(string(dockerServerResult)))
		return nil
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

	return nil
}

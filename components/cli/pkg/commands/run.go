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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunRun(cellImageTag string) error {
	parsedCellImage, err := util.ParseImage(cellImageTag)
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while parsing cell image: \x1b[0m %v \n", err)
		os.Exit(1)
	}

	repoLocation := filepath.Join(util.UserHomeDir(), ".cellery", "repos", parsedCellImage.RegistryHost,
		parsedCellImage.Organization, parsedCellImage.ImageName, parsedCellImage.ImageVersion)
	fmt.Printf("Running cell image: %s ...\n", util.Bold(cellImageTag))
	zipLocation := filepath.Join(repoLocation, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)

	if _, err := os.Stat(zipLocation); os.IsNotExist(err) {
		fmt.Printf("\nUnable to find image %s locally.", cellImageTag)
		fmt.Printf("\nPulling image: %s ...", cellImageTag)
		_ = RunPull(cellImageTag, true)
	}

	//create tmp directory
	currentTIme := time.Now()
	timstamp := currentTIme.Format("20060102150405")
	tmpPath := filepath.Join(util.UserHomeDir(), ".cellery", "tmp", timstamp)
	err = util.CreateDir(tmpPath)
	if err != nil {
		panic(err)
	}

	err = util.Unzip(zipLocation, tmpPath)
	if err != nil {
		panic(err)
	}

	kubeYamlDir := filepath.Join(tmpPath, "artifacts", "cellery")

	cmd := exec.Command("kubectl", "apply", "-f", kubeYamlDir)
	execError := ""
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Printf("\033[36m%s\033[m\n", stdoutScanner.Text())
		}
	}()
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			execError += stderrScanner.Text()
		}
	}()
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Error in executing cell run: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()

	_ = os.RemoveAll(tmpPath)

	if err != nil {
		fmt.Printf("\x1b[31;1m\n Error occurred while running cell image:\x1b[0m %v \n", execError)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf(util.GreenBold("\U00002714")+" Successfully deployed cell image: %s\n", util.Bold(cellImageTag))
	util.PrintWhatsNextMessage("cellery ps")

	return nil
}

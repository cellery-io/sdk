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

package internal

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/celleryio/sdk/components/cli/pkg/constants"
	"github.com/celleryio/sdk/components/cli/pkg/util"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func newRunCommand() *cobra.Command {
	var cellImage string
	cmd := &cobra.Command{
		Use:   "run [OPTIONS]",
		Short: "Use a cell image to create a running instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
				return nil
			}
			cellImage = args[0]
			err := run(cellImage)
			if err != nil {
				cmd.Help()
				return err
			}
			return nil
		},
		Example: "  cellery run my-project:1.0.0 -n myproject-v1.0.0",
	}
	return cmd
}

func RunRun(cellImageTag string) error {
	if cellImageTag == "" {
		return fmt.Errorf("please specify the cell image")
	}

	registryHost := constants.CENTRAL_REGISTRY_HOST
	organization := ""
	imageName := ""
	imageVersion := ""

	strArr := strings.Split(cellImageTag, "/")
	if len(strArr) == 3 {
		registryHost = strArr[0]
		organization = strArr[1]
		imageTag := strings.Split(strArr[2], ":")
		if len(imageTag) != 2 {
			util.ExitWithImageFormatError()
		}
		imageName = imageTag[0]
		imageVersion = imageTag[1]
	} else if len(strArr) == 2 {
		organization = strArr[0]
		imageTag := strings.Split(strArr[1], ":")
		if len(imageTag) != 2 {
			util.ExitWithImageFormatError()
		}
		imageName = imageTag[0]
		imageVersion = imageTag[1]
	} else {
		util.ExitWithImageFormatError()
	}

	repoLocation := filepath.Join(util.UserHomeDir(), ".cellery", "repos", registryHost, organization, imageName,
		imageVersion)
	fmt.Printf("Running cell image: %s ...\n", util.Bold(cellImageTag))
	zipLocation := filepath.Join(repoLocation, imageName+constants.CELL_IMAGE_EXT)

	if _, err := os.Stat(zipLocation); os.IsNotExist(err) {
		fmt.Printf("\nUnable to find image %s locally.", cellImageTag)
		fmt.Printf("\nPulling image: %s ...", cellImageTag)
		_ = RunPull(cellImageTag, true)
	}

	//create tmp directory
	currentTIme := time.Now()
	timstamp := currentTIme.Format("20060102150405")
	tmpPath := filepath.Join(util.UserHomeDir(), ".cellery", "tmp", timstamp)
	err := util.CreateDir(tmpPath)
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

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

package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tj/go-spin"
	"github.com/wso2/cellery/components/cli/pkg/constants"
	"github.com/wso2/cellery/components/cli/pkg/util"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func newPullCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull [CELL IMAGE]",
		Short: "pull cell image from the remote repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
				return nil
			}
			cellImage = args[0]
			err := runPull(cellImage, false)
			if err != nil {
				cmd.Help()
				return err
			}
			return nil
		},
		Example: "  cellery pull mycellery.org/hello:v1",
	}
	return cmd
}

/**
Spinner
*/
func pullSpinner(tag string) {
	s := spin.New()
	for {
		if isSpinning {
			fmt.Printf("\r\033[36m%s\033[m Pulling %s %s", s.Next(), "image", util.Bold(tag))
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func runPull(cellImage string, silent bool) error {

	if cellImage == "" {
		return fmt.Errorf("no cell image specified")
	}

	registryHost := constants.CENTRAL_REGISTRY_HOST
	organization := ""
	imageName := ""
	imageVersion := ""

	strArr := strings.Split(cellImage, "/")
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

	if !silent {
		go pullSpinner(cellImage)
	}

	var url = "http://" + registryHost + constants.REGISTRY_BASE_PATH + "/images/" + organization + "/" +
		imageName + "/" + imageVersion

	repoLocation := filepath.Join(util.UserHomeDir(), ".cellery", "repos", registryHost, organization, imageName,
		imageVersion)

	repoCreateErr := util.CreateDir(repoLocation)
	if repoCreateErr != nil {
		fmt.Println("Error while creating image location: " + repoCreateErr.Error())
		os.Exit(1)
	}

	zipLocation := filepath.Join(repoLocation, imageName+constants.CELL_IMAGE_EXT)
	response, downloadError := util.DownloadFile(zipLocation, url)

	if downloadError != nil {
		fmt.Printf("\x1b[31;1m Error occurred while pulling the cell image: \x1b[0m %v \n", downloadError)
		os.Exit(1)
	}

	if response.StatusCode == 200 {
		fmt.Println()
		fmt.Printf(util.GreenBold("\n\U00002714")+" Successfully pulled cell image: %s\n", util.Bold(cellImage))
		if !silent {
			util.PrintWhatsNextMessage("cellery run " + cellImage)
		}
	}
	if response.StatusCode == 404 {
		fmt.Println()
		fmt.Printf("\x1b[31;1mError occurred while pulling cell image:\x1b[0m %v. "+
			"Repository does not exist or may require authorization\n", cellImage)
		os.Exit(1)
	}
	return nil
}

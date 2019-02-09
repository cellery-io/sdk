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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tj/go-spin"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

/**
Spinner
*/
func pullSpinner(tag string) {
	s := spin.New()
	for {
		fmt.Printf("\r\033[36m%s\033[m Pulling %s %s", s.Next(), "image", util.Bold(tag))
		time.Sleep(100 * time.Millisecond)
	}
}

func RunPull(cellImage string, silent bool) error {
	parsedCellImage, err := util.ParseImage(cellImage)
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while parsing cell image: \x1b[0m %v \n", err)
		os.Exit(1)
	}

	if !silent {
		go pullSpinner(cellImage)
	}

	var url = "http://" + parsedCellImage.RegistryHost + constants.REGISTRY_BASE_PATH + "/images/" +
		parsedCellImage.Organization + "/" + parsedCellImage.ImageName + "/" + parsedCellImage.ImageVersion

	repoLocation := filepath.Join(util.UserHomeDir(), ".cellery", "repos", parsedCellImage.RegistryHost,
		parsedCellImage.Organization, parsedCellImage.ImageName, parsedCellImage.ImageVersion)

	repoCreateErr := util.CreateDir(repoLocation)
	if repoCreateErr != nil {
		fmt.Println("Error while creating image location: " + repoCreateErr.Error())
		os.Exit(1)
	}

	zipLocation := filepath.Join(repoLocation, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)
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

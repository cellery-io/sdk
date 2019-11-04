/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/image"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunExtractResources extracts the cell image zip file and copies the resources folder to the provided path
func RunExtractResources(cellImage string, outputPath string) {
	parsedCellImage, err := image.ParseImageTag(cellImage)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
	}

	repoLocation := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo", parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion)
	imageLocation := filepath.Join(repoLocation, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)

	// Checking if the image is present in the local repo
	isImagePresent, _ := util.FileExists(imageLocation)
	if !isImagePresent {
		util.ExitWithErrorMessage(fmt.Sprintf("Failed to extract resources for image %s", util.Bold(cellImage)),
			errors.New("Image not Found"))
	}

	// Create temp directory
	currentTIme := time.Now()
	timestamp := currentTIme.Format("27065102350415")
	tempPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "tmp", timestamp)
	err = util.CreateDir(tempPath)
	if err != nil {
		util.ExitWithErrorMessage("Error while extracting resources from cell image", err)
	}
	defer func() {
		err = os.RemoveAll(tempPath)
		if err != nil {
			util.ExitWithErrorMessage("Error while cleaning up", err)
		}
	}()

	err = util.Unzip(imageLocation, tempPath)
	if err != nil {
		util.ExitWithErrorMessage("Error extracting zip file", err)
	}

	// Copying the image resources to the provided output directory
	resourcesDir, err := filepath.Abs(filepath.Join(tempPath, artifacts, "resources"))
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while extracting the image resources", err)
	}

	resourcesExists, _ := util.FileExists(resourcesDir)
	if resourcesExists {
		if outputPath == "" {
			currentPath, err := util.GetCurrentPath()
			if err != nil {
				util.ExitWithErrorMessage("Error occurred while extracting the image resources", err)
			}
			outputPath = currentPath
		}
		err = util.CopyDir(resourcesDir, outputPath)
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while extracting the image resources", err)
		}

		absOutputPath, _ := filepath.Abs(outputPath)
		fmt.Printf("\nExtracted Resources: %s", util.Bold(absOutputPath))
		util.PrintSuccessMessage(fmt.Sprintf("Successfully extracted cell image resources: %s", util.Bold(cellImage)))
	} else {
		fmt.Printf("\n%s No resources available in %s\n", util.CyanBold("\U00002139"), util.Bold(cellImage))
	}
}

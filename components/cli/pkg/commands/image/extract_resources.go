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

package image

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunExtractResources extracts the cell image zip file and copies the resources folder to the provided path
func RunExtractResources(cli cli.Cli, cellImage string, outputPath string) error {
	parsedCellImage, err := image.ParseImageTag(cellImage)
	if err != nil {
		return fmt.Errorf("error occurred while parsing cell image, %v", err)
	}

	repoLocation := filepath.Join(cli.FileSystem().Repository(), parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion)
	imageLocation := filepath.Join(repoLocation, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)

	// Checking if the image is present in the local repo
	isImagePresent, _ := util.FileExists(imageLocation)
	if !isImagePresent {
		return fmt.Errorf(fmt.Sprintf("failed to extract resources for image %s, %v", util.Bold(cellImage),
			errors.New("image not Found")))
	}

	// Create temp directory
	currentTIme := time.Now()
	timestamp := currentTIme.Format("27065102350415")
	tempPath := filepath.Join(cli.FileSystem().TempDir(), timestamp)
	err = util.CreateDir(tempPath)
	if err != nil {
		return fmt.Errorf("error while extracting resources from cell image, %v", err)
	}
	defer func() error {
		if err = os.RemoveAll(tempPath); err != nil {
			return fmt.Errorf("error while cleaning up, %v", err)
		}
		return nil
	}()

	if err = util.Unzip(imageLocation, tempPath); err != nil {
		return fmt.Errorf("error extracting zip file, %v", err)
	}

	// Copying the image resources to the provided output directory
	resourcesDir, err := filepath.Abs(filepath.Join(tempPath, artifacts, "resources"))
	if err != nil {
		return fmt.Errorf("error occurred while extracting the image resources, %v", err)
	}

	resourcesExists, _ := util.FileExists(resourcesDir)
	if resourcesExists {
		if outputPath == "" {
			outputPath = cli.FileSystem().CurrentDir()
		}
		if err = util.CopyDir(resourcesDir, outputPath); err != nil {
			return fmt.Errorf("error occurred while extracting the image resources, %v", err)
		}
		absOutputPath, _ := filepath.Abs(outputPath)
		fmt.Fprintf(cli.Out(), fmt.Sprintf("\nExtracted Resources: %s", util.Bold(absOutputPath)))
		util.PrintSuccessMessage(fmt.Sprintf("Successfully extracted cell image resources: %s", util.Bold(cellImage)))
	} else {
		fmt.Fprintf(cli.Out(), fmt.Sprintf("\n%s No resources available in %s\n", util.CyanBold("\U00002139"),
			util.Bold(cellImage)))
	}
	return nil
}

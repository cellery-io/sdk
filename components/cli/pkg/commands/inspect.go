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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/image"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunInspect extracts the cell image and lists the files in the cell image
func RunInspect(cellImage string) {
	parsedCellImage, err := image.ParseImageTag(cellImage)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
	}
	cellImageFile := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo", parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)

	// Checking if the image is present in the local repo
	isImagePresent, _ := util.FileExists(cellImageFile)
	if !isImagePresent {
		util.ExitWithErrorMessage(fmt.Sprintf("Failed to list files for image %s", util.Bold(cellImage)),
			errors.New("Image not Found"))
	}

	// Create temp directory
	currentTime := time.Now()
	timestamp := currentTime.Format("27065102350415")
	tempPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "tmp", timestamp)
	err = util.CreateDir(tempPath)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while reading the cell image", err)
	}
	defer func() {
		err = os.RemoveAll(tempPath)
		if err != nil {
			util.ExitWithErrorMessage("Error while cleaning up", err)
		}
	}()

	// Unzipping Cellery Image
	err = util.Unzip(cellImageFile, tempPath)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while reading the cell image", err)
	}

	// Printing the cell image directory structure
	fmt.Printf("\n%s\n  │ \n", util.Bold(fmt.Sprintf("%s/%s:%s", parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion)))
	err = printCellImageDirectory(tempPath, 0, []bool{})
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while printing the cell image files", err)
	}
	fmt.Println()
}

func printCellImageDirectory(dir string, nestingLevel int, ancestorBranchPrintRequirement []bool) error {
	fds, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for i, fd := range fds {
		//fmt.Printf("%d %d\n", nestingLevel, i)
		fileName := fd.Name()

		for j := 0; j < nestingLevel; j++ {
			if ancestorBranchPrintRequirement[j] {
				fmt.Printf("  │ ")
			} else {
				fmt.Printf("    ")
			}
		}

		if i == len(fds)-1 {
			fmt.Print("  └")
		} else {
			fmt.Print("  ├")
		}
		fmt.Printf("──%s\n", fileName)

		if fd.IsDir() {
			err = printCellImageDirectory(path.Join(dir, fileName), nestingLevel+1,
				append(ancestorBranchPrintRequirement, i != len(fds)-1))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

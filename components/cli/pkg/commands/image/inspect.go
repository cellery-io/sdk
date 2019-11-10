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

package image

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunInspect extracts the cell image and lists the files in the cell image
func RunInspect(cli cli.Cli, cellImage string) error {
	parsedCellImage, err := image.ParseImageTag(cellImage)
	if err != nil {
		return fmt.Errorf("error occurred while parsing cell image, %v", err)
	}
	cellImageFile := filepath.Join(cli.FileSystem().Repository(), parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)

	// Checking if the image is present in the local repo
	isImagePresent, _ := util.FileExists(cellImageFile)
	if !isImagePresent {
		return fmt.Errorf(fmt.Sprintf("failed to list files for image %s, %v", util.Bold(cellImage), errors.New("image not Found")))
	}

	// Create temp directory
	currentTime := time.Now()
	timestamp := currentTime.Format("27065102350415")
	tempPath := filepath.Join(cli.FileSystem().TempDir(), timestamp)
	err = util.CreateDir(tempPath)
	if err != nil {
		return fmt.Errorf("error occurred while reading the cell image, %v", err)
	}
	defer func() error {
		err = os.RemoveAll(tempPath)
		if err != nil {
			return fmt.Errorf("error while cleaning up, %v", err)
		}
		return nil
	}()

	// Unzipping Cellery Image
	err = util.Unzip(cellImageFile, tempPath)
	if err != nil {
		return fmt.Errorf("error occurred while reading the cell image, %v", err)
	}

	// Printing the cell image directory structure
	fmt.Fprintf(cli.Out(), "\n%s\n  │ \n", util.Bold(fmt.Sprintf("%s/%s:%s", parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion)))
	err = printCellImageDirectory(cli, tempPath, 0, []bool{})
	if err != nil {
		return fmt.Errorf("error occurred while printing the cell image files, %v", err)
	}
	fmt.Fprintln(cli.Out())
	return nil
}

func printCellImageDirectory(cli cli.Cli, dir string, nestingLevel int, ancestorBranchPrintRequirement []bool) error {
	fds, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for i, fd := range fds {
		fileName := fd.Name()

		for j := 0; j < nestingLevel; j++ {
			if ancestorBranchPrintRequirement[j] {
				fmt.Fprintf(cli.Out(), "  │ ")
			} else {
				fmt.Fprintf(cli.Out(), "    ")
			}
		}
		if i == len(fds)-1 {
			fmt.Fprintf(cli.Out(), "  └")
		} else {
			fmt.Fprintf(cli.Out(), "  ├")
		}
		fmt.Fprintf(cli.Out(), "──%s\n", fileName)

		if fd.IsDir() {
			err = printCellImageDirectory(cli, path.Join(dir, fileName), nestingLevel+1,
				append(ancestorBranchPrintRequirement, i != len(fds)-1))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

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

package commands

import (
	"fmt"
	"path"
	"regexp"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
)

func RunDeleteImage(cli cli.Cli, images []string, regex string, deleteAll bool) error {
	repoLocation := cli.FileSystem().Repository()
	imagesInRepo, err := getImagesArray(cli)
	if err != nil {
		return fmt.Errorf("error getting images array, %v", err)
	}
	for _, imageInRepo := range imagesInRepo {
		parsedCellImage, err := image.ParseImageTag(imageInRepo.name)
		if err != nil {
			return fmt.Errorf("error occurred while parsing cell image, %v", err)
		}
		cellImagePath := path.Join(repoLocation, parsedCellImage.Organization,
			parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)
		if deleteAll {
			return cli.FileSystem().RemoveAll(cellImagePath)
		} else {
			if regex != "" {
				// Check if image name matches regex pattern
				regexMatches, err := regexp.MatchString(regex, imageInRepo.name)
				if err != nil {
					return fmt.Errorf("error checking if pattern matches with image name, %v", err)
				}
				if regexMatches {
					return cli.FileSystem().RemoveAll(cellImagePath)
					continue
				}
			}
			if len(images) > 0 {
				for _, imageToBeDeleted := range images {
					if imageInRepo.name == imageToBeDeleted {
						return cli.FileSystem().RemoveAll(cellImagePath)
						break
					}
				}
			}
		}
	}
	return nil
}

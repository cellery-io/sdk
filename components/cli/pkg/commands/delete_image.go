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
	"os"
	"path"
	"regexp"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunDeleteImage(images []string, regex string, deleteAll bool) {
	imagesInRepo := getImagesArray()
	for _, imageInRepo := range imagesInRepo {
		parsedCellImage, err := util.ParseImageTag(imageInRepo.name)
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
		}
		cellImagePath := path.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo", parsedCellImage.Organization,
			parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)
		if deleteAll {
			_ = os.RemoveAll(cellImagePath)
		} else {
			if regex != "" {
				// Check if image name matches regex pattern
				regexMatches, err := regexp.MatchString(regex, imageInRepo.name)
				if err != nil {
					util.ExitWithErrorMessage("Error checking if pattern matches with image name", err)
				}
				if regexMatches {
					_ = os.RemoveAll(cellImagePath)
					continue
				}
			}
			if len(images) > 0 {
				for _, imageToBeDeleted := range images {
					if imageInRepo.name == imageToBeDeleted {
						_ = os.RemoveAll(cellImagePath)
						break
					}
				}
			}
		}
	}
}

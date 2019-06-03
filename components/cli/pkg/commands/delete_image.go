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
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunDeleteImage(images string) {
	imagesToBeDeleted := strings.Split(images, ",")
	// Get the list of images in repo
	existingImages := getImagesArray()
	for _, imageToBeDeleted := range imagesToBeDeleted {
		imageExists := false
		for _, existingImage := range existingImages {
			var err error
			imageExists, err = regexp.MatchString(imageToBeDeleted, existingImage.name)
			if err != nil {
				util.ExitWithErrorMessage("Error checking if image exists", err)
			}
			if imageExists {
				parsedCellImage, err := util.ParseImageTag(existingImage.name)
				if err != nil {
					util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
				}
				cellImagePath := path.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo", parsedCellImage.Organization,
					parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)
				_ = os.RemoveAll(cellImagePath)
			}
		}
	}
}

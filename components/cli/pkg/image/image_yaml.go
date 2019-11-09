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
 *
 */

package image

import (
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func ReadCellImageYaml(repo, cellImage string) []byte {
	parsedCellImage, err := ParseImageTag(cellImage)
	cellImageZip := path.Join(repo, parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)
	// Create tmp directory
	tmpPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "tmp", "imageExtracted")
	err = util.CleanOrCreateDir(tmpPath)
	if err != nil {
		panic(err)
	}
	err = util.Unzip(cellImageZip, tmpPath)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while extracting cell image", err)
	}
	cellYamlContent, err := ioutil.ReadFile(filepath.Join(tmpPath, constants.ZIP_ARTIFACTS, "cellery",
		parsedCellImage.ImageName+".yaml"))
	if err != nil {
		util.ExitWithErrorMessage("Error while reading cell image content", err)
	}
	// Delete tmp directory
	err = util.CleanOrCreateDir(tmpPath)
	if err != nil {
		util.ExitWithErrorMessage("Error while deleting temporary file", err)
	}
	return cellYamlContent
}

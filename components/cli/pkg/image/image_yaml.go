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
 *
 */

package image

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/util"
)

func ReadCellImageYaml(repo, cellImage string) ([]byte, error) {
	parsedCellImage, err := ParseImageTag(cellImage)
	cellImageZip := path.Join(repo, parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+constants.CellImageExt)
	// Create tmp directory
	tmpPath := filepath.Join(util.UserHomeDir(), constants.CelleryHome, "tmp", "imageExtracted")
	err = util.CleanOrCreateDir(tmpPath)
	if err != nil {
		panic(err)
	}
	err = util.Unzip(cellImageZip, tmpPath)
	if err != nil {
		return nil, fmt.Errorf("error occurred while extracting cell image, %v", err)
	}
	cellYamlContent, err := ioutil.ReadFile(filepath.Join(tmpPath, constants.ZipArtifacts, "cellery",
		parsedCellImage.ImageName+".yaml"))
	if err != nil {
		return nil, fmt.Errorf("error while reading cell image content, %v", err)
	}
	// Delete tmp directory
	err = util.CleanOrCreateDir(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("error while deleting temporary file, %v", err)
	}
	return cellYamlContent, nil
}

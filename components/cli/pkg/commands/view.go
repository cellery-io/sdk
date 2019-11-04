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

// RunView opens the View for a particular Cell Image
func RunView(cellImage string) {
	celleryHomeDocsViewDir := path.Join(util.CelleryInstallationDir(), constants.CELLERY_HOME_DOCS_VIEW_DIR)
	errorMessage := "Error occurred while generating Docs View"

	// Making a copy of the Docs Viewer
	docsViewDir, err := ioutil.TempDir("", "cellery-docs-view")
	if err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}
	err = util.CopyDir(celleryHomeDocsViewDir, docsViewDir)
	if err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}

	// Finding Cell Image location
	parsedCellImage, err := image.ParseImageTag(cellImage)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
	}
	cellImageFile := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo", parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)

	// Create temp directory
	currentTime := time.Now()
	timestamp := currentTime.Format("27065102350415")
	tempPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "tmp", timestamp)
	err = util.CreateDir(tempPath)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while unpacking Cell Image", err)
	}
	defer func() {
		err = os.RemoveAll(tempPath)
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while cleaning up", err)
		}
	}()

	// Unzipping Cellery Image
	err = util.Unzip(cellImageFile, tempPath)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while unpacking Cell Image", err)
	}

	metadataFileContent, err := ioutil.ReadFile(filepath.Join(tempPath, artifacts, "cellery", "metadata.json"))
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while reading Cell metadata", err)
	}

	docsViewData := "window.__CELL_METADATA__ = " + string(metadataFileContent) + ";"
	err = ioutil.WriteFile(filepath.Join(docsViewDir, "data", "cell.js"), []byte(docsViewData), 0666)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while creating Cell view", err)
	}

	// Opening browser
	docsViewIndexFile := path.Join(docsViewDir, "index.html")
	fmt.Printf("Cell Image Viewer: file://%s\n\n", docsViewIndexFile)
	err = util.OpenBrowser(docsViewIndexFile)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while opening the browser", err)
	}
}

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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/image"
	"cellery.io/cellery/components/cli/pkg/util"
)

const celleryHomeDocsViewDir = "docs-view"

// RunView opens the View for a particular Cell Image
func RunView(cli cli.Cli, cellImage string) error {
	celleryHomeDocsViewDir := path.Join(cli.FileSystem().CelleryInstallationDir(), celleryHomeDocsViewDir)

	// Making a copy of the Docs Viewer
	docsViewDir, err := ioutil.TempDir("", "cellery-docs-view")
	if err != nil {
		return fmt.Errorf("error creating temp dir, %v", err)
	}
	if err = util.CopyDir(celleryHomeDocsViewDir, docsViewDir); err != nil {
		return fmt.Errorf("error copying docs view dir, %v", err)
	}
	// Finding Cell Image location
	parsedCellImage, err := image.ParseImageTag(cellImage)
	if err != nil {
		return fmt.Errorf("error occurred while parsing cell image, %v", err)
	}
	cellImageFile := filepath.Join(cli.FileSystem().Repository(), parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+cellImageExt)

	// Create temp directory
	currentTime := time.Now()
	timestamp := currentTime.Format("27065102350415")
	tempPath := filepath.Join(cli.FileSystem().TempDir(), timestamp)
	if err = util.CreateDir(tempPath); err != nil {
		return fmt.Errorf("error occurred while unpacking Cell Image, %v", err)
	}
	defer func() error {
		if err = os.RemoveAll(tempPath); err != nil {
			return fmt.Errorf("error occurred while cleaning up, %v", err)
		}
		return nil
	}()
	// Unzipping Cellery Image
	if err = util.Unzip(cellImageFile, tempPath); err != nil {
		return fmt.Errorf("error occurred while unpacking Cell Image, %v", err)
	}
	metadataFileContent, err := ioutil.ReadFile(filepath.Join(tempPath, artifacts, "cellery", "metadata.json"))
	if err != nil {
		return fmt.Errorf("error occurred while reading Cell metadata, %v", err)
	}

	docsViewData := "window.__CELL_METADATA__ = " + string(metadataFileContent) + ";"
	if err = ioutil.WriteFile(filepath.Join(docsViewDir, "data", "cell.js"), []byte(docsViewData), 0666); err != nil {
		return fmt.Errorf("error occurred while creating Cell view, %v", err)
	}
	// Opening browser
	docsViewIndexFile := path.Join(docsViewDir, "index.html")
	fmt.Fprintf(cli.Out(), "Cell Image Viewer: file://%s\n\n", docsViewIndexFile)
	if err = cli.OpenBrowser(docsViewIndexFile); err != nil {
		return fmt.Errorf("error occurred while opening the browser, %v", err)
	}
	return nil
}

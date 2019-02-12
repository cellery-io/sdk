/*
 * Copyright (c) 2018 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nokia/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunPull connects to the Cellery Registry and pulls the cell image and saves it in the local repository.
// This also adds the relevant ballerina files to the ballerina repo directory.
func RunPull(cellImage string, silent bool) error {
	err := pullImage(cellImage, "", "", silent)
	if err != nil {
		fmt.Println()
		username, password, err := util.RequestCredentials()
		if err != nil {
			util.ExitWithErrorMessage("Failed to acquire credentials", err)
		}
		fmt.Println()

		err = pullImage(cellImage, username, password, silent)
		if err != nil {
			util.ExitWithErrorMessage("Failed to pull image", err)
		}
	}
	return nil
}

func pullImage(cellImage string, username string, password string, silent bool) error {
	parsedCellImage, err := util.ParseImageTag(cellImage)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
	}
	repository := parsedCellImage.Organization + "/" + parsedCellImage.ImageName

	var spinner *util.Spinner = nil
	if !silent {
		imageName := fmt.Sprintf("%s/%s:%s", parsedCellImage.Organization, parsedCellImage.ImageName,
			parsedCellImage.ImageVersion)
		spinner = util.StartNewSpinner(fmt.Sprintf("Pulling image %s from %s", util.Bold(imageName),
			util.Bold(parsedCellImage.Registry)))
		defer func() {
			spinner.IsSpinning = false
		}()
	}

	// Initiating a connection to Cellery Registry
	hub, err := registry.New("https://"+parsedCellImage.Registry, username, password)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while initializing connection to the Cellery Registry", err)
	}

	// Fetching the Docker Image Manifest
	cellImageManifest, err := hub.Manifest(repository, "0.1.0")
	if err != nil {
		return err
	}

	var cellImageDigest digest.Digest
	if len(cellImageManifest.References()) == 1 {
		cellImageReference := cellImageManifest.References()[0]
		cellImageDigest = cellImageReference.Digest

		// Downloading the Cell Image from the repository
		reader, err := hub.DownloadBlob(repository, cellImageReference.Digest)
		if err != nil {
			return err
		}
		if reader != nil {
			defer func() {
				err = reader.Close()
				if err != nil {
					util.ExitWithErrorMessage("Error occurred while cleaning up", err)
				}
			}()
		}
		bytes, err := ioutil.ReadAll(reader)
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while pulling cell image", err)
		}

		repoLocation := filepath.Join(util.UserHomeDir(), ".cellery", "repo", parsedCellImage.Organization,
			parsedCellImage.ImageName, parsedCellImage.ImageVersion)

		// Cleaning up the old image if it already exists
		hasOldImage, err := util.FileExists(repoLocation)
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while removing the old cell image", err)
		}
		if hasOldImage {
			err = os.RemoveAll(repoLocation)
			if err != nil {
				util.ExitWithErrorMessage("Error while cleaning up", err)
			}
		}

		// Creating the Repo location
		err = util.CreateDir(repoLocation)
		if err != nil {
			util.ExitWithErrorMessage("Error while saving cell image to local repo", err)
		}

		// Writing the Cell Image to local file
		cellImageFile := filepath.Join(repoLocation, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)
		err = ioutil.WriteFile(cellImageFile, bytes, 0644)
		if err != nil {
			util.ExitWithErrorMessage("Error while saving cell image to local repo", err)
		}

		util.AddImageToBalPath(parsedCellImage)
	} else {
		util.ExitWithErrorMessage("Invalid cell image",
			errors.New(fmt.Sprintf("expected exactly 1 File Layer, but found %d",
				len(cellImageManifest.References()))))
	}

	spinner.IsSpinning = false
	fmt.Print("\n\nImage Digest : " + util.Bold(cellImageDigest))
	util.PrintSuccessMessage(fmt.Sprintf("Successfully pulled cell image: %s", util.Bold(cellImage)))
	if !silent {
		util.PrintWhatsNextMessage("run the image", "cellery run "+cellImage)
	}

	return nil
}

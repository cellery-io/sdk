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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nokia/docker-registry-client/registry"

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
			fmt.Printf("\x1b[31;1m Failed to acquire credentials: \x1b[0m %v \n", err)
			os.Exit(1)
		}
		fmt.Println()

		err = pullImage(cellImage, username, password, silent)
		if err != nil {
			fmt.Printf("\x1b[31;1m Failed to pull image: \x1b[0m %v \n", err)
			os.Exit(1)
		}
	}
	return nil
}

func pullImage(cellImage string, username string, password string, silent bool) error {
	parsedCellImage, err := util.ParseImage(cellImage)
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while parsing cell image: \x1b[0m %v \n", err)
		os.Exit(1)
	}
	repository := parsedCellImage.Organization + "/" + parsedCellImage.ImageName

	var spinner *util.Spinner = nil
	if !silent {
		spinner = util.StartNewSpinner("Pulling image " + util.Bold(cellImage))
		defer func() {
			spinner.IsSpinning = false
		}()
	}

	// Initiating a connection to Cellery Registry
	hub, err := registry.New("https://"+parsedCellImage.Registry, username, password)
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while initializing connection to the Cellery Registry: "+
			"\x1b[0m %v \n", err)
		os.Exit(1)
	}

	// Fetching the Docker Image Manifest
	cellImageManifest, err := hub.Manifest(repository, "0.1.0")
	if err != nil {
		return err
	}

	// Fetching the Docker Image Digest
	cellImageDigest, err := hub.ManifestDigest(repository, "0.1.0")
	if err != nil {
		return err
	}

	if len(cellImageManifest.References()) == 1 {
		cellImageReference := cellImageManifest.References()[0]

		// Downloading the Cell Image from the repository
		reader, err := hub.DownloadBlob(repository, cellImageReference.Digest)
		if err != nil {
			return err
		}
		if reader != nil {
			defer func() {
				err = reader.Close()
				if err != nil {
					fmt.Printf("\x1b[31;1m Error occurred while pulling cell image: \x1b[0m %v \n", err)
					os.Exit(1)
				}
			}()
		}
		bytes, err := ioutil.ReadAll(reader)
		if err != nil {
			fmt.Printf("\x1b[31;1m Error occurred while pulling cell image: \x1b[0m %v \n", err)
			os.Exit(1)
		}

		repoLocation := filepath.Join(util.UserHomeDir(), ".cellery", "repos", parsedCellImage.Registry,
			parsedCellImage.Organization, parsedCellImage.ImageName, parsedCellImage.ImageVersion)

		// Cleaning up the old image if it already exists
		hasOldImage, err := util.FileExists(repoLocation)
		if err != nil {
			fmt.Printf("\x1b[31;1m Error occurred while removing the old cell image: \x1b[0m %v \n", err)
			os.Exit(1)
		}
		if hasOldImage {
			err = os.RemoveAll(repoLocation)
			if err != nil {
				fmt.Printf("\x1b[31;1m Error while cleaning up: \x1b[0m %v \n", err)
				os.Exit(1)
			}
		}

		// Creating the Repo location
		err = util.CreateDir(repoLocation)
		if err != nil {
			fmt.Printf("\x1b[31;1m Error while saving cell image to local repo: \x1b[0m %v \n", err)
			os.Exit(1)
		}

		// Writing the Cell Image to Local File
		cellImageFile := filepath.Join(repoLocation, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)
		err = ioutil.WriteFile(cellImageFile, bytes, 0644)
		if err != nil {
			fmt.Printf("\x1b[31;1m Error while saving cell image to local repo: \x1b[0m %v \n", err)
			os.Exit(1)
		}

		util.AddImageToBalPath(parsedCellImage)
	} else {
		fmt.Printf("\x1b[31;1m Invalid cell image: \x1b[0m Expected exactly 1 File Layer, but found %d \n",
			len(cellImageManifest.References()))
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("\nImage Digest : " + util.Bold(cellImageDigest))
	fmt.Printf(util.GreenBold("\U00002714")+" Successfully pulled cell image: %s\n", util.Bold(cellImage))
	if !silent {
		util.PrintWhatsNextMessage("run the image", "cellery run "+cellImage)
	}

	return nil
}

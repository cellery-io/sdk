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
	"strings"

	"github.com/nokia/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/registry/credentials"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

// RunPull connects to the Cellery Registry and pulls the cell image and saves it in the local repository.
// This also adds the relevant ballerina files to the ballerina repo directory.
func RunPull(cellImage string, isSilent bool, username string, password string) {
	parsedCellImage, err := util.ParseImageTag(cellImage)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
	}

	var registryCredentials = &credentials.RegistryCredentials{
		Registry: parsedCellImage.Registry,
		Username: username,
		Password: password,
	}
	if username != "" && password == "" {
		username, password, err = credentials.FromTerminal(username)
	}
	isCredentialsPresent := err == nil && registryCredentials.Username != "" &&
		registryCredentials.Password != ""

	var credManager credentials.CredManager
	if !isCredentialsPresent {
		credManager, err = credentials.NewCredManager()
		if err != nil {
			util.ExitWithErrorMessage("Unable to use a Credentials Manager, "+
				"please use inline flags instead", err)
		}
		savedCredentials, err := credManager.GetCredentials(parsedCellImage.Registry)
		if err == nil && savedCredentials.Username != "" && savedCredentials.Password != "" {
			registryCredentials = savedCredentials
			isCredentialsPresent = true
		} else {
			isCredentialsPresent = false
		}
	}

	if isCredentialsPresent {
		// Pulling the image using the saved credentials
		err = pullImage(parsedCellImage, registryCredentials.Username, registryCredentials.Password)
		if err != nil {
			util.ExitWithErrorMessage("Failed to pull image", err)
		}
	} else {
		// Pulling image without credentials
		err = pullImage(parsedCellImage, "", "")
		if err != nil {
			if strings.Contains(err.Error(), "401") {
				// Requesting the credentials since server responded with an Unauthorized status code
				var isAuthorized chan bool
				var done chan bool
				finalizeChannelCalls := func(isAuthSuccessful bool) {
					if isAuthorized != nil {
						isAuthorized <- isAuthSuccessful
					}
					if done != nil {
						<-done
					}
				}
				if strings.HasSuffix(parsedCellImage.Registry, constants.CENTRAL_REGISTRY_HOST) {
					isAuthorized = make(chan bool)
					done = make(chan bool)
					registryCredentials.Username, registryCredentials.Password, err = credentials.FromBrowser(username,
						isAuthorized, done)
				} else {
					registryCredentials.Username, registryCredentials.Password, err = credentials.FromTerminal(username)
				}
				if err != nil {
					finalizeChannelCalls(false)
					util.ExitWithErrorMessage("Failed to acquire credentials", err)
				}
				finalizeChannelCalls(true)
				fmt.Println()

				// Trying to pull the image again with the provided credentials
				err = pullImage(parsedCellImage, registryCredentials.Username, registryCredentials.Password)
				if err != nil {
					util.ExitWithErrorMessage("Failed to pull image", err)
				}

				if credManager != nil {
					err = credManager.StoreCredentials(registryCredentials)
					if err == nil {
						fmt.Printf("\n%s Saved Credentials for %s Registry", util.GreenBold("\U00002714"),
							util.Bold(parsedCellImage.Registry))
					} else {
						fmt.Printf("\n\n%s %s", util.YellowBold("\U000026A0"),
							"Error occurred while saving credentials")
					}
				}
			} else {
				util.ExitWithErrorMessage("Failed to pull image", err)
			}
		}
	}
	if !isSilent {
		util.PrintSuccessMessage(fmt.Sprintf("Successfully pulled cell image: %s", util.Bold(cellImage)))
		util.PrintWhatsNextMessage("run the image", "cellery run "+cellImage)
	}
}

func pullImage(parsedCellImage *util.CellImage, username string, password string) error {
	repository := parsedCellImage.Organization + "/" + parsedCellImage.ImageName

	spinner := util.StartNewSpinner("Connecting to " + util.Bold(parsedCellImage.Registry))
	defer func() {
		spinner.Stop(true)
	}()

	// Initiating a connection to Cellery Registry
	hub, err := registry.New("https://"+parsedCellImage.Registry, username, password)
	if err != nil {
		spinner.Stop(false)
		return fmt.Errorf("failed to initialize connection to Cellery Registry %v", err)
	}

	// Fetching the Docker Image Manifest
	spinner.SetNewAction("Fetching metadata")
	cellImageManifest, err := hub.Manifest(repository, parsedCellImage.ImageVersion)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	var cellImageDigest digest.Digest
	if len(cellImageManifest.References()) == 1 {
		cellImageReference := cellImageManifest.References()[0]
		cellImageDigest = cellImageReference.Digest

		imageName := fmt.Sprintf("%s/%s:%s", parsedCellImage.Organization, parsedCellImage.ImageName,
			parsedCellImage.ImageVersion)
		spinner.SetNewAction("Pulling image " + util.Bold(imageName))

		// Downloading the Cell Image from the repository
		reader, err := hub.DownloadBlob(repository, cellImageReference.Digest)
		if err != nil {
			spinner.Stop(false)
			return err
		}
		if reader != nil {
			defer func() {
				err = reader.Close()
				if err != nil {
					spinner.Stop(false)
					util.ExitWithErrorMessage("Error occurred while cleaning up", err)
				}
			}()
		}
		bytes, err := ioutil.ReadAll(reader)
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error occurred while pulling cell image", err)
		}

		repoLocation := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo", parsedCellImage.Organization,
			parsedCellImage.ImageName, parsedCellImage.ImageVersion)

		// Cleaning up the old image if it already exists
		hasOldImage, err := util.FileExists(repoLocation)
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error occurred while removing the old cell image", err)
		}
		if hasOldImage {
			spinner.SetNewAction("Removing old Image")
			err = os.RemoveAll(repoLocation)
			if err != nil {
				spinner.Stop(false)
				util.ExitWithErrorMessage("Error while cleaning up", err)
			}
		}

		spinner.SetNewAction("Saving new Image to the Local Repository")

		// Creating the Repo location
		err = util.CreateDir(repoLocation)
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error occurred while saving cell image to local repo", err)
		}

		// Writing the Cell Image to local file
		cellImageFile := filepath.Join(repoLocation, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)
		err = ioutil.WriteFile(cellImageFile, bytes, 0644)
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error occurred while saving cell image to local repo", err)
		}

		// Validating image compatibility with Cellery installation
		metadata, err := image.ReadMetaData(parsedCellImage.Organization, parsedCellImage.ImageName,
			parsedCellImage.ImageVersion)
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Invalid cell image", err)
		}
		// TODO : Add a proper validation based on major, minor, patch, version before stable release
		if metadata.BuildCelleryVersion != "" && metadata.BuildCelleryVersion != version.BuildVersion() {
			fmt.Printf("\r\x1b[2K%s Pulled image version (%s) and Cellery installation version (%s) "+
				"do not match. The image %s/%s:%s may not work properly with this installation.\n",
				util.YellowBold("\U000026A0"), util.Bold(metadata.BuildCelleryVersion), version.BuildVersion(),
				parsedCellImage.Organization, parsedCellImage.ImageName, parsedCellImage.ImageVersion)
		}
	} else {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Invalid cell image",
			errors.New(fmt.Sprintf("expected exactly 1 File Layer, but found %d",
				len(cellImageManifest.References()))))
	}

	spinner.Stop(true)
	fmt.Printf("\nImage Digest : %s\n", util.Bold(cellImageDigest))
	return nil
}

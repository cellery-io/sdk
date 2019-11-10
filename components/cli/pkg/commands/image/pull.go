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

package image

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/registry/credentials"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

// RunPull connects to the Cellery Registry and pulls the cell image and saves it in the local repository.
// This also adds the relevant ballerina files to the ballerina repo directory.
func RunPull(cli cli.Cli, cellImage string, isSilent bool, username string, password string) error {
	parsedCellImage, err := image.ParseImageTag(cellImage)
	if err != nil {
		return fmt.Errorf("error occurred while parsing cell image, %v", err)
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
			return fmt.Errorf("unable to use a Credentials Manager, please use inline flags instead, %v", err)
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
		err = pullImage(cli, parsedCellImage, registryCredentials.Username, registryCredentials.Password)
	} else {
		// Pulling image without credentials
		err = pullImage(cli, parsedCellImage, "", "")
	}
	if err != nil {
		// Need to check 404 since docker auth does not validates the image tag
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "404") {
			return fmt.Errorf(fmt.Sprintf("image %s/%s:%s not found in Registry %s",
				parsedCellImage.Organization, parsedCellImage.ImageName, parsedCellImage.ImageVersion,
				parsedCellImage.Registry), err)
		} else {
			return fmt.Errorf("failed to pull image, %v", err)
		}
	}
	// Validating image compatibility with Cellery installation
	repoLocation := cli.FileSystem().Repository()
	metadata, err := image.ReadMetaData(repoLocation, parsedCellImage.Organization, parsedCellImage.ImageName,
		parsedCellImage.ImageVersion)
	if err != nil {
		return fmt.Errorf("invalid cell image, %v", err)
	}
	// TODO : Add a proper validation based on major, minor, patch, version before stable release
	if metadata.BuildCelleryVersion != "" && metadata.BuildCelleryVersion != version.BuildVersion() {
		fmt.Fprint(cli.Out(), fmt.Sprintf("\r\x1b[2K%s Pulled cell image's build version (%s) and Cellery "+
			"installation version (%s) do not match. The image %s/%s:%s may not work properly with this installation.\n",
			util.YellowBold("\U000026A0"), util.Bold(metadata.BuildCelleryVersion), version.BuildVersion(),
			parsedCellImage.Organization, parsedCellImage.ImageName, parsedCellImage.ImageVersion))
	}
	if !isSilent {
		util.PrintSuccessMessage(fmt.Sprintf("Successfully pulled cell image: %s", util.Bold(cellImage)))
		util.PrintWhatsNextMessage("run the image", "cellery run "+cellImage)
	}
	return nil
}

func pullImage(cli cli.Cli, parsedCellImage *image.CellImage, username string, password string) error {
	var err error
	var cellImage []byte
	if err := cli.ExecuteTask("Pulling cell image", "Failed to pull image",
		"", func() error {
			cellImage, err = cli.Registry().Pull(parsedCellImage, username, password)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
		return fmt.Errorf("error pulling image, %v", err)
	}
	repoLocation := filepath.Join(cli.FileSystem().Repository(), parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion)
	// Cleaning up the old image if it already exists
	hasOldImage, err := util.FileExists(repoLocation)
	if err != nil {
		return fmt.Errorf("error occurred while removing the old cell image, %v", err)
	}
	if hasOldImage {
		err = os.RemoveAll(repoLocation)
		if err != nil {
			return fmt.Errorf("error while cleaning up, %v", err)
		}
	}
	fmt.Fprintln(cli.Out(), "Saving new Image to the Local Repository")
	// Creating the Repo location
	err = util.CreateDir(repoLocation)
	if err != nil {
		return fmt.Errorf("error occurred while saving cell image to local repo, %v", err)
	}
	// Writing the Cell Image to local file
	cellImageFile := filepath.Join(repoLocation, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)
	err = ioutil.WriteFile(cellImageFile, cellImage, 0644)
	if err != nil {
		return fmt.Errorf("error occurred while saving cell image to local repo, %v", err)
	}
	return nil
}

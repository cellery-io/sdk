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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/registry/credentials"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunPush parses the cell image name to recognize the Cellery Registry (A Docker Registry), Organization and version
// and pushes to the Cellery Registry
func RunPush(cli cli.Cli, cellImage string, username string, password string) error {
	parsedCellImage, err := image.ParseImageTag(cellImage)
	//Read docker images from metadata.json
	imageDir, err := ExtractImage(cli, parsedCellImage, false)
	if err != nil {
		return fmt.Errorf("error occurred while extracting image, %v", err)
	}
	metadataFileContent, err := ioutil.ReadFile(filepath.Join(imageDir, artifacts, "cellery",
		"metadata.json"))
	if err != nil {
		return fmt.Errorf("error occurred while reading Cell Image metadata, %v", err)
	}
	cellImageMetadata := &image.MetaData{}
	err = json.Unmarshal(metadataFileContent, cellImageMetadata)
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
	var dockerImagesToBePushed []string
	for _, componentMetadata := range cellImageMetadata.Components {
		if componentMetadata.IsDockerPushRequired {
			dockerImagesToBePushed = append(dockerImagesToBePushed, componentMetadata.DockerImage)
		}
	}
	if isCredentialsPresent {
		// Pushing the image using the saved credentials
		err = pushImage(cli, parsedCellImage, registryCredentials.Username, registryCredentials.Password)
		if err != nil {
			return fmt.Errorf("failed to push image, %v", err)
		}
		pushDockerImages(dockerImagesToBePushed)
	} else {
		// Pushing image without credentials
		err = pushImage(cli, parsedCellImage, "", "")
		if err != nil {
			if strings.Contains(err.Error(), "401") {
				log.Printf("Unauthorized to push Cell image. Trying to login")
				// Requesting the credentials since server responded with an Unauthorized status code
				var isAuthorized chan bool
				var done chan bool
				finalizeChannelCalls := func(isAuthSuccessful bool) {
					if isAuthorized != nil {
						log.Printf("Writing authorized status: %v to channel from main goroutine", isAuthorized)
						isAuthorized <- isAuthSuccessful
						log.Printf("Finished writing authorized status: %v to channel from main goroutine", isAuthorized)
					}
					if done != nil {
						log.Printf("Reading done channel from main goroutine to wait for server task finish")
						<-done
						log.Printf("Finished reading done channel from main goroutine")
					}
				}
				regex, err := regexp.Compile(constants.CentralRegistryHostRegx)
				if err != nil {
					return fmt.Errorf("error occurred while compiling the registry regex, %v", err)
				}
				if regex.MatchString(parsedCellImage.Registry) {
					isAuthorized = make(chan bool)
					done = make(chan bool)
					registryCredentials.Username, registryCredentials.Password, err = credentials.FromBrowser(username,
						isAuthorized, done)
				} else {
					registryCredentials.Username, registryCredentials.Password, err = credentials.FromTerminal(username)
				}
				if err != nil {
					finalizeChannelCalls(false)
					return fmt.Errorf("failed to acquire credentials, %v", err)
				}
				finalizeChannelCalls(true)
				fmt.Println()

				// Trying to push the image again with the provided credentials
				err = pushImage(cli, parsedCellImage, registryCredentials.Username, registryCredentials.Password)
				if err != nil {
					return fmt.Errorf("failed to push image, %v", err)
				}
				pushDockerImages(dockerImagesToBePushed)

				if credManager != nil {
					log.Printf("Storing credentials in Credentials Manager")
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
				return fmt.Errorf("failed to pull image, %v", err)
			}
		} else {
			pushDockerImages(dockerImagesToBePushed)
		}
	}
	util.PrintSuccessMessage(fmt.Sprintf("Successfully pushed cell image: %s", util.Bold(cellImage)))
	util.PrintWhatsNextMessage("pull the image", "cellery pull "+cellImage)
	return nil
}

func pushDockerImages(dockerImages []string) error {
	log.Printf("Pushing docker images [%s]", strings.Join(dockerImages, ", "))
	for _, elem := range dockerImages {
		cmd := exec.Command("docker", "push", elem)
		execError := ""
		stderrReader, _ := cmd.StderrPipe()
		stderrScanner := bufio.NewScanner(stderrReader)
		go func() {
			for stderrScanner.Scan() {
				execError += stderrScanner.Text()
			}
		}()
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Start()
		if err != nil {
			errStr := string(stderr.Bytes())
			fmt.Printf("%s\n", errStr)
			return fmt.Errorf("error occurred while pushing Docker image, %v", err)
		}
		err = cmd.Wait()
		if err != nil {
			fmt.Println()
			fmt.Printf("\x1b[31;1m\nPush Failed.\x1b[0m %v \n", execError)
			fmt.Println("\x1b[31;1m======================\x1b[0m")
			errStr := string(stderr.Bytes())
			fmt.Printf("\x1b[31;1m%s\x1b[0m", errStr)
			return fmt.Errorf("error occurred while pushing cell image, %v", err)
		}
	}
	return nil
}

func pushImage(cli cli.Cli, parsedCellImage *image.CellImage, username string, password string) error {
	log.Printf("Pushing image %s/%s:%s to registry %s", parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.Registry)
	repository := parsedCellImage.Organization + "/" + parsedCellImage.ImageName
	cellImage := parsedCellImage.Registry + "/" + repository + ":" + parsedCellImage.ImageVersion

	imageName := fmt.Sprintf("%s/%s:%s", parsedCellImage.Organization, parsedCellImage.ImageName,
		parsedCellImage.ImageVersion)
	fmt.Println(fmt.Sprintf("\nReading image %s from the Local Repository", util.Bold(imageName)))
	// Reading the cell image
	cellImageFilePath := filepath.Join(cli.FileSystem().Repository(), parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+cellImageExt)

	// Checking if the image is present in the local repo
	isImagePresent, _ := util.FileExists(cellImageFilePath)
	if !isImagePresent {
		return fmt.Errorf(fmt.Sprintf("Failed to push image %s", util.Bold(cellImage)),
			errors.New("image not Found"))
	}

	cellImageFile, err := os.Open(cellImageFilePath)
	if err != nil {
		return fmt.Errorf("error occurred while reading the cell image, %v", err)
	}
	if cellImageFile != nil {
		defer func() error {
			err := cellImageFile.Close()
			if err != nil {
				return fmt.Errorf("error occurred while opening the cell image, %v", err)
			}
			return nil
		}()
	}
	cellImageFileBytes, err := ioutil.ReadAll(cellImageFile)
	if err != nil {
		return fmt.Errorf("error occurred while reading the cell image, %v", err)
	}
	if err := cli.ExecuteTask("Pushing cell image", "Failed to push image",
		"", func() error {
			err = cli.Registry().Push(parsedCellImage, cellImageFileBytes, username, password)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
		return fmt.Errorf("error pushing image, %v", err)
	}
	return nil
}

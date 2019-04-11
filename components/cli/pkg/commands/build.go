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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunBuild executes the cell's build life cycle method and saves the generated cell image to the local repo.
// This also copies the relevant ballerina files to the ballerina repo directory.
func RunBuild(tag string, fileName string) {
	fileExist, err := util.FileExists(fileName)
	if !fileExist {
		util.ExitWithErrorMessage("Unable to build image",
			errors.New(fmt.Sprintf("file '%s' does not exist", util.Bold(fileName))))
	}

	parsedCellImage, err := util.ParseImageTag(tag)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
	}
	repoLocation := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo", parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion)

	spinner := util.StartNewSpinner("Building image " + util.Bold(tag))
	defer func() {
		spinner.Stop(true)
	}()

	// First clean target directory if exists
	projectDir, err := os.Getwd()
	targetDir := filepath.Join(projectDir, "target")
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error in getting current directory location", err)
	}
	_ = os.RemoveAll(targetDir)

	var imageName = &util.CellImageName{
		Organization: parsedCellImage.Organization,
		Name:         parsedCellImage.ImageName,
		Version:      parsedCellImage.ImageVersion,
	}
	iName, err := json.Marshal(imageName)
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error in generating cellery:CellImageName construct", err)
	}
	// Executing the build method in the cell file
	moduleMgr := &util.BLangManager{}
	exePath, err := moduleMgr.GetExecutablePath()
	if err != nil {
		util.ExitWithErrorMessage("Failed to get executable path", err)
	}
	cmd := exec.Command(exePath+"ballerina", "run", constants.BALLERINA_PRINT_RETURN_FLAG, fileName+":build", string(iName))
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
	err = cmd.Start()
	if err != nil {
		spinner.Stop(false)
		errStr := string(stderr.Bytes())
		fmt.Printf("%s\n", errStr)
		util.ExitWithErrorMessage("Error occurred while building cell image", err)
	}
	err = cmd.Wait()
	if err != nil {
		spinner.Stop(false)
		fmt.Println()
		fmt.Printf("\x1b[31;1m\nBuild Failed.\x1b[0m %v \n", execError)
		fmt.Println("\x1b[31;1m======================\x1b[0m")
		errStr := string(stderr.Bytes())
		fmt.Printf("\x1b[31;1m%s\x1b[0m", errStr)
		util.ExitWithErrorMessage("Error occurred while building cell image", err)
	}

	outStr := string(stdout.Bytes())
	fmt.Printf("\r\x1b[2K\033[36m%s\033[m\n", outStr)

	generateMetaData(parsedCellImage, targetDir, spinner)

	folderCopyError := util.CopyDir(targetDir, filepath.Join(projectDir, constants.ZIP_ARTIFACTS))
	if folderCopyError != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred creating cell image", err)
	}
	err = util.CleanOrCreateDir(filepath.Join(projectDir, constants.ZIP_BALLERINA_SOURCE))
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred while creating the cell image", err)
	}
	fileCopyError := util.CopyFile(fileName, filepath.Join(projectDir, constants.ZIP_BALLERINA_SOURCE, filepath.Base(fileName)))
	if fileCopyError != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred creating cell image", err)
	}
	folders := []string{constants.ZIP_ARTIFACTS, constants.ZIP_BALLERINA_SOURCE}
	output := parsedCellImage.ImageName + ".zip"
	err = util.RecursiveZip(nil, folders, output)
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred while creating the cell image", err)
	}

	_ = os.RemoveAll(filepath.Join(projectDir, constants.ZIP_ARTIFACTS))
	_ = os.RemoveAll(filepath.Join(projectDir, constants.ZIP_BALLERINA_SOURCE))

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
			util.ExitWithErrorMessage("Error occurred while cleaning up", err)
		}
	}

	spinner.SetNewAction("Saving new Image to the Local Repository")
	repoCreateErr := util.CreateDir(repoLocation)
	if repoCreateErr != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred while creating image location", err)
	}

	// Cleaning up temp directory generated by Cellery language
	_ = os.Remove(path.Join(targetDir, "tmp"))

	zipSrc := filepath.Join(projectDir, output)
	zipDst := filepath.Join(repoLocation, output)
	zipCopyError := util.CopyFile(zipSrc, zipDst)
	if zipCopyError != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred while saving cell image to local repo", err)
	}

	_ = os.Remove(zipSrc)
	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully built cell image: %s", util.Bold(tag)))
	util.PrintWhatsNextMessage("run the image", "cellery run "+tag)
}

// generateMetaData generates the metadata file for cellery
func generateMetaData(cellImage *util.CellImage, targetDir string, spinner *util.Spinner) {
	errorMessage := "Error occurred while generating metadata"
	dependenciesFile := path.Join(targetDir, "tmp", "dependencies.properties")
	dependenciesMap := map[string]*util.CellImageMetaData{}

	// Generating the alias to dependency Cell Image map
	dependenciesFileExists, err := util.FileExists(dependenciesFile)
	if err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}
	if dependenciesFileExists {
		dependenciesFileBytes, err := ioutil.ReadFile(dependenciesFile)
		if err != nil {
			util.ExitWithErrorMessage(errorMessage, err)
		}

		dependenciesFileLines := strings.Split(string(dependenciesFileBytes), "\n")
		for _, line := range dependenciesFileLines {
			if !strings.Contains(line, "=") { // Handling empty lines and other problematic conditions
				continue
			}

			lineSplit := strings.Split(line, "=")
			alias := strings.TrimSpace(lineSplit[0])
			dependency := strings.TrimSpace(lineSplit[1])

			dependencyCellImage, err := util.ParseImageTag(dependency)
			if err != nil {
				util.ExitWithErrorMessage(errorMessage, err)
			}
			cellImageZip := path.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo",
				dependencyCellImage.Organization, dependencyCellImage.ImageName, dependencyCellImage.ImageVersion,
				dependencyCellImage.ImageName+constants.CELL_IMAGE_EXT)

			// Pulling the dependency if not exist (This will not be executed most of the time)
			dependencyExists, err := util.FileExists(cellImageZip)
			if !dependencyExists {
				spinner.Pause()
				RunPull(dependency, true, "", "")
				fmt.Println()
				spinner.Resume()
			}

			// Create temp directory
			currentTime := time.Now()
			timestamp := currentTime.Format("27065102350415")
			tempPath := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "tmp", timestamp)
			err = util.CreateDir(tempPath)
			if err != nil {
				util.ExitWithErrorMessage(errorMessage, err)
			}

			// Unzipping Cellery Image
			err = util.Unzip(cellImageZip, tempPath)
			if err != nil {
				util.ExitWithErrorMessage(errorMessage, err)
			}

			// Reading the dependency's metadata
			metadataJsonContent, err := ioutil.ReadFile(
				filepath.Join(tempPath, "artifacts", "cellery", "metadata.json"))
			if err != nil {
				fmt.Println(dependenciesFileExists)
				util.ExitWithErrorMessage(errorMessage+". metadata.json file not found for dependency: "+dependency,
					err)
			}
			dependencyMetadata := &util.CellImageMetaData{}
			err = json.Unmarshal(metadataJsonContent, dependencyMetadata)
			if err != nil {
				util.ExitWithErrorMessage(errorMessage, err)
			}

			dependenciesMap[alias] = dependencyMetadata

			// Cleaning up
			err = os.RemoveAll(tempPath)
			if err != nil {
				util.ExitWithErrorMessage("Error occurred while cleaning up", err)
			}
		}
	}

	cellYamlContent, err := ioutil.ReadFile(filepath.Join(targetDir, "cellery", cellImage.ImageName+".yaml"))
	if err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}
	k8sCell := &util.Cell{}
	err = yaml.Unmarshal(cellYamlContent, k8sCell)
	if err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}

	// Getting the components of the Cell Image being built
	var components []string
	for _, component := range k8sCell.CellSpec.ComponentTemplates {
		components = append(components, component.Metadata.Name)
	}

	// Writing the metadata file
	outputFileMetadata := &util.CellImageMetaData{
		CellImageName: util.CellImageName{
			Organization: cellImage.Organization,
			Name:         cellImage.ImageName,
			Version:      cellImage.ImageVersion,
		},
		BuildTimestamp: time.Now().Unix(),
		Components:     components,
		Dependencies:   dependenciesMap,
	}
	metadataFileContent, err := json.Marshal(outputFileMetadata)
	if err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}
	metadataFile := filepath.Join(targetDir, "cellery", "metadata.json")
	err = ioutil.WriteFile(metadataFile, metadataFileContent, 0666)
	if err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}
}

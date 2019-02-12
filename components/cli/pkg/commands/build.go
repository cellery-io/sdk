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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunBuild executes the cell's build life cycle method and saves the generated cell image to the local repo.
// This also copies the relevant ballerina files to the ballerina repo directory.
func RunBuild(tag string, fileName string) error {
	fileExist, err := util.FileExists(fileName)
	if !fileExist {
		util.ExitWithErrorMessage("Unable to build image",
			errors.New(fmt.Sprintf("file '%s' does not exist", util.Bold(fileName))))
	}

	parsedCellImage, err := util.ParseImageTag(tag)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
	}

	repoLocation := filepath.Join(util.UserHomeDir(), ".cellery", "repo", parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion)

	spinner := util.StartNewSpinner("Building image " + util.Bold(tag))
	defer func() {
		spinner.IsSpinning = false
	}()

	// First clean target directory if exists
	projectDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		util.ExitWithErrorMessage("Error in getting current directory location", err)
	}
	_ = os.RemoveAll(filepath.Join(projectDir, "target"))

	// Executing the build method in the cell file
	cmd := exec.Command("ballerina", "run", fileName+":build", parsedCellImage.ImageName,
		parsedCellImage.ImageVersion)
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
		errStr := string(stderr.Bytes())
		fmt.Printf("%s\n", errStr)
		util.ExitWithErrorMessage("Error occurred while building cell image", err)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println()
		fmt.Printf("\x1b[31;1m\nBuild Failed.\x1b[0m %v \n", execError)
		fmt.Println("\x1b[31;1m======================\x1b[0m")
		errStr := string(stderr.Bytes())
		fmt.Printf("\x1b[31;1m%s\x1b[0m", errStr)
		util.ExitWithErrorMessage("Error occurred while building cell image", err)
	}

	outStr := string(stdout.Bytes())
	fmt.Printf("\n\033[36m%s\033[m\n", outStr)

	// Creating additional Ballerina.toml file for ballerina reference project
	tomlTemplate := "[project]\n" +
		"org-name = \"" + parsedCellImage.Organization + "\"\n" +
		"version = \"" + parsedCellImage.ImageVersion + "\"\n"
	tomlFile, err := os.Create(filepath.Join(projectDir, "target", "bal", "Ballerina.toml"))
	if err != nil {
		util.ExitWithErrorMessage("Error in creating Toml File", err)
	}
	defer func() {
		err = tomlFile.Close()
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while cleaning up", err)
		}
	}()
	writer := bufio.NewWriter(tomlFile)
	_, err = writer.WriteString(tomlTemplate)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while creating cell reference", err)
	}
	err = writer.Flush()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred creating cell reference", err)
	}

	folderCopyError := util.CopyDir(filepath.Join(projectDir, "target"), filepath.Join(projectDir, "artifacts"))
	if folderCopyError != nil {
		util.ExitWithErrorMessage("Error occurred creating cell image", err)
	}
	folders := []string{"artifacts"}
	files := []string{fileName}
	output := parsedCellImage.ImageName + ".zip"
	err = util.RecursiveZip(files, folders, output)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while creating the cell image", err)
	}

	_ = os.RemoveAll(filepath.Join(projectDir, "artifacts"))

	// Cleaning up the old image if it already exists
	hasOldImage, err := util.FileExists(repoLocation)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while removing the old cell image", err)
	}
	if hasOldImage {
		err = os.RemoveAll(repoLocation)
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while cleaning up", err)
		}
	}

	repoCreateErr := util.CreateDir(repoLocation)
	if repoCreateErr != nil {
		util.ExitWithErrorMessage("Error occurred while creating image location", err)
	}

	zipSrc := filepath.Join(projectDir, output)
	zipDst := filepath.Join(repoLocation, output)
	zipCopyError := util.CopyFile(zipSrc, zipDst)
	if zipCopyError != nil {
		util.ExitWithErrorMessage("Error occurred while saving cell image to local repo", err)
	}

	_ = os.Remove(zipSrc)

	util.AddImageToBalPath(parsedCellImage)

	spinner.IsSpinning = false
	util.PrintSuccessMessage(fmt.Sprintf("Successfully built cell image: %s", util.Bold(tag)))
	util.PrintWhatsNextMessage("run the image", "cellery run "+tag)
	return nil
}

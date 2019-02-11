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
		fmt.Printf("\x1b[31;1m Please check the filename. File '%s' does not exist.\x1b[0m\n", fileName)
		os.Exit(1)
	}

	parsedCellImage, err := util.ParseImageTag(tag)
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while parsing cell image: \x1b[0m %v \n", err)
		os.Exit(1)
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
		fmt.Println("\x1b[31;1m Error in getting current directory location: \x1b[0m %v \n" + err.Error())
		os.Exit(1)
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
		fmt.Printf("\x1b[31;1m Error in executing cell build: \x1b[0m %v \n", err)
		errStr := string(stderr.Bytes())
		fmt.Printf("%s\n", errStr)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println()
		fmt.Printf("\x1b[31;1m\nBuild Failed.\x1b[0m %v \n", execError)
		fmt.Println("\x1b[31;1m======================\x1b[0m")
		errStr := string(stderr.Bytes())
		fmt.Printf("\x1b[31;1m%s\x1b[0m", errStr)
		os.Exit(1)
	}

	outStr := string(stdout.Bytes())
	fmt.Printf("\n\033[36m%s\033[m\n", outStr)

	// Creating additional Ballerina.toml file for ballerina reference project
	tomlTemplate := "[project]\n" +
		"org-name = \"" + parsedCellImage.Organization + "\"\n" +
		"version = \"" + parsedCellImage.ImageVersion + "\"\n"
	tomlFile, err := os.Create(filepath.Join(projectDir, "target", "bal", "Ballerina.toml"))
	if err != nil {
		fmt.Println("Error in creating Toml File: " + err.Error())
		os.Exit(1)
	}
	defer func() {
		err = tomlFile.Close()
		if err != nil {
			fmt.Printf("\x1b[31;1m Error occurred while cleaning up: \x1b[0m %v \n", err)
			os.Exit(1)
		}
	}()
	writer := bufio.NewWriter(tomlFile)
	_, err = writer.WriteString(tomlTemplate)
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while creating cell reference: \x1b[0m %v \n", err)
		os.Exit(1)
	}
	err = writer.Flush()
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred creating cell reference: \x1b[0m %v \n", err)
		os.Exit(1)
	}

	folderCopyError := util.CopyDir(filepath.Join(projectDir, "target"), filepath.Join(projectDir, "artifacts"))
	if folderCopyError != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	folders := []string{"artifacts"}
	files := []string{fileName}
	output := parsedCellImage.ImageName + ".zip"
	err = util.RecursiveZip(files, folders, output)
	if err != nil {
		fmt.Printf("\x1b[31;1m Cell build finished with error: \x1b[0m %v \n", err)
		os.Exit(1)
	}

	_ = os.RemoveAll(filepath.Join(projectDir, "artifacts"))

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

	repoCreateErr := util.CreateDir(repoLocation)
	if repoCreateErr != nil {
		fmt.Println("\x1b[31;1m Error while creating image location: \x1b[0m %v \n" + repoCreateErr.Error())
		os.Exit(1)
	}

	zipSrc := filepath.Join(projectDir, output)
	zipDst := filepath.Join(repoLocation, output)
	zipCopyError := util.CopyFile(zipSrc, zipDst)
	if zipCopyError != nil {
		fmt.Println("\x1b[31;1m Error while saving cell image to local repo: \x1b[0m %v \n" + zipCopyError.Error())
		os.Exit(1)
	}

	_ = os.Remove(zipSrc)

	util.AddImageToBalPath(parsedCellImage)

	spinner.IsSpinning = false
	fmt.Println()
	fmt.Printf("\n%s Successfully built cell image: %s\n", util.GreenBold("\U00002714"), util.Bold(tag))
	util.PrintWhatsNextMessage("run the image", "cellery run "+tag)
	return nil
}

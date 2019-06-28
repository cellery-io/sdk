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
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
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

	tempBuildFileName, err := util.CreateTempExecutableBalFile(fileName, "build")
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error executing ballerina file", err)
	}

	cmd := &exec.Cmd{}

	if exePath != "" {
		cmd = exec.Command(exePath+"ballerina", "run", tempBuildFileName, "build", string(iName), "{}")
	} else {
		currentDir, err := os.Getwd()
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error in determining working directory", err)
		}
		//Retrieve the cellery cli docker instance status.
		cmdDockerPs := exec.Command("docker", "ps", "--filter",
			"label=ballerina-runtime="+version.BuildVersion(),
			"--filter", "label=currentDir="+currentDir, "--filter", "status=running", "--format", "{{.ID}}")
		containerId, err := cmdDockerPs.Output()
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error in retrieving cellery cli docker instance status", err)
		}

		if string(containerId) == "" {
			cmdDockerRun := exec.Command("docker", "run", "-d",
				"-l", "ballerina-runtime="+version.BuildVersion(),
				"-l", "current.dir="+currentDir,
				"--mount", "type=bind,source="+currentDir+",target=/home/cellery/src",
				"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".ballerina,target=/home/cellery/.ballerina",
				"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".cellery,target=/home/cellery/.cellery",
				"wso2cellery/ballerina-runtime:"+version.BuildVersion(), "sleep", "600",
			)
			stderrReader, err := cmdDockerRun.StderrPipe()
			if err != nil {
				spinner.Stop(false)
				util.ExitWithErrorMessage("Error while building stderr pipe ", err)
			}
			stdoutReader, _ := cmdDockerRun.StdoutPipe()
			if err != nil {
				spinner.Stop(false)
				util.ExitWithErrorMessage("Error while building stdout pipe ", err)
			}

			stderrScanner := bufio.NewScanner(stderrReader)
			stdoutScanner := bufio.NewScanner(stdoutReader)

			err = cmdDockerRun.Start()
			if err != nil {
				spinner.Stop(false)
				util.ExitWithErrorMessage("Error while starting docker process ", err)
			}

			go func() {
				for {
					if stderrScanner.Scan() && strings.HasPrefix(stderrScanner.Text(), "Unable to find image") {
						spinner.Pause()
						spinner.Stop(false)
						util.StartNewSpinner(fmt.Sprintf("%s: Cannot find ballerina docker image. Pulling %s", "Building image "+util.Bold(tag), "wso2cellery/ballerina-runtime:"+version.BuildVersion()))
						spinner.Resume()
						break
					}
				}
			}()

			go func() {
				for {
					if stdoutScanner.Scan() {
						containerId = []byte(stdoutScanner.Text())
						break
					}
				}
			}()

			err = cmdDockerRun.Wait()
			if err != nil {
				spinner.Stop(false)
				util.ExitWithErrorMessage("Error while running ballerina-runtime docker image", err)
			}
			time.Sleep(5 * time.Second)
		}

		cliUser, err := user.Current()
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error while retrieving the current user", err)
		}

		if cliUser.Uid != constants.CELLERY_DOCKER_CLI_USER_ID {
			cmdUserExist := exec.Command("docker", "exec", strings.TrimSpace(string(containerId)),
				"id", "-u", cliUser.Username)
			_, errUserExist := cmdUserExist.Output()
			if errUserExist != nil {
				cmdUserAdd := exec.Command("docker", "exec", strings.TrimSpace(string(containerId)), "useradd", "-m",
					"-d", "/home/cellery", "--uid", cliUser.Uid, cliUser.Username)

				_, errUserAdd := cmdUserAdd.Output()
				if errUserAdd != nil {
					spinner.Stop(false)
					util.ExitWithErrorMessage("Error in adding Cellery execution user", errUserAdd)
				}
			}
		}

		re := regexp.MustCompile("^"+currentDir+"/")
		balFilePath := re.ReplaceAllString(tempBuildFileName,"")
		cmd = exec.Command("docker", "exec", "-w", "/home/cellery/src", "-u", cliUser.Uid,
			strings.TrimSpace(string(containerId)), constants.DOCKER_CLI_BALLERINA_EXECUTABLE_PATH, "run", balFilePath, "build", string(iName), "{}")
	}
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
	_ = os.Remove(tempBuildFileName)
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

	metadataFile := filepath.Join(targetDir, "cellery", "metadata.json")
	metadataJSON, err := ioutil.ReadFile(metadataFile)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while reading metadata "+metadataFile, err)
	}

	metadata := &util.CellImageMetaData{
		BuildCelleryVersion: version.BuildVersion(),
		BuildTimestamp:      time.Now().Unix(),
	}
	err = json.Unmarshal(metadataJSON, metadata)
	if err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}

	for alias, dependencyMetadata := range metadata.Dependencies {
		cellImageZip := path.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo",
			dependencyMetadata.Organization, dependencyMetadata.Name, dependencyMetadata.Version,
			dependencyMetadata.Name+constants.CELL_IMAGE_EXT)

		dependencyImage := dependencyMetadata.Organization + "/" + dependencyMetadata.Name +
			":" + dependencyMetadata.Version
		if cellImage.Registry != "" {
			dependencyImage = cellImage.Registry + "/" + dependencyImage
		}

		// Pulling the dependency if not exist (This will not be executed most of the time)
		dependencyExists, err := util.FileExists(cellImageZip)
		if !dependencyExists {
			spinner.Pause()
			RunPull(dependencyImage, true, "", "")
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
			util.ExitWithErrorMessage(errorMessage+". metadata.json file not found for dependency: "+dependencyImage,
				err)
		}
		dependencyMetadata := &util.CellImageMetaData{}
		err = json.Unmarshal(metadataJsonContent, dependencyMetadata)
		if err != nil {
			util.ExitWithErrorMessage(errorMessage, err)
		}

		metadata.Dependencies[alias] = dependencyMetadata

		// Cleaning up
		err = os.RemoveAll(tempPath)
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while cleaning up", err)
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
	metadata.Components = components

	// Getting the Ingress Types
	if k8sCell.CellSpec.GateWayTemplate.GatewaySpec.TcpApis != nil {
		metadata.Ingresses = append(metadata.Ingresses, "TCP")
	}
	if k8sCell.CellSpec.GateWayTemplate.GatewaySpec.HttpApis != nil {
		metadata.Ingresses = append(metadata.Ingresses, "HTTP")
	}
	if k8sCell.CellSpec.GateWayTemplate.GatewaySpec.GrpcApis != nil {
		metadata.Ingresses = append(metadata.Ingresses, "GRPC")
	}

	metadataFileContent, err := json.Marshal(metadata)
	if err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}

	err = ioutil.WriteFile(metadataFile, metadataFileContent, 0666)
	if err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}
}

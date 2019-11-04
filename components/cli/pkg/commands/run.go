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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/cellery-io/sdk/components/cli/ballerina"
	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

const celleryEnvVarPrefix = "cellery_env_"
const celleryImageDirEnvVar = "CELLERY_IMAGE_DIR"

// RunRun starts Cell instance (along with dependency instances if specified by the user)
// This also support linking instances to parts of the dependency tree
// This command also strictly validates whether the requested Cell (and the dependencies are valid)
func RunRun(cli cli.Cli, cellImageTag string, instanceName string, startDependencies bool, shareDependencies bool,
	dependencyLinks []string, envVars []string) error {
	var err error
	var parsedCellImage *image.CellImage
	if parsedCellImage, err = image.ParseImageTag(cellImageTag); err != nil {
		return fmt.Errorf("error occurred while parsing cell image, %v", err)
	}
	var imageDir string
	if err = cli.ExecuteTask("Extracting cell image", "Failed to extract cell image",
		"", func() error {
			imageDir, err = ExtractImage(parsedCellImage, true)
			return err
		}); err != nil {
		return err
	}
	defer func() error {
		if err = os.RemoveAll(imageDir); err != nil {
			return fmt.Errorf("error occurred while removing image, %v", err)
		}
		return nil
	}()
	// Reading Cell Image metadata
	var metadataFileContent []byte
	if metadataFileContent, err = ioutil.ReadFile(filepath.Join(imageDir, artifacts, "cellery",
		"metadata.json")); err != nil {
		return fmt.Errorf("error occurred while reading Image metadata, %v", err)
	}
	cellImageMetadata := &image.MetaData{}
	if err = json.Unmarshal(metadataFileContent, cellImageMetadata); err != nil {
		return fmt.Errorf("error occurred while reading Image metadata, %v", err)
	}
	var parsedDependencyLinks []*dependencyAliasLink
	if len(dependencyLinks) > 0 {
		// Parsing the dependency links list
		for _, link := range dependencyLinks {
			var dependencyLink *dependencyAliasLink
			linkSplit := strings.Split(link, ":")
			if strings.Contains(linkSplit[0], ".") {
				instanceSplit := strings.Split(linkSplit[0], ".")
				dependencyLink = &dependencyAliasLink{
					Instance:           instanceSplit[0],
					DependencyAlias:    instanceSplit[1],
					DependencyInstance: linkSplit[1],
				}
			} else {
				dependencyLink = &dependencyAliasLink{
					DependencyAlias:    linkSplit[0],
					DependencyInstance: linkSplit[1],
				}
			}
			parsedDependencyLinks = append(parsedDependencyLinks, dependencyLink)
		}
	}
	var instanceEnvVars []*environmentVariable
	if len(envVars) > 0 {
		// Parsing environment variables
		for _, envVar := range envVars {
			var targetInstance string
			var envVarKey string
			var envVarValue string

			// Parsing the environment variable
			r := regexp.MustCompile(fmt.Sprintf("^%s$", celleryArgEnvVarPattern))
			matches := r.FindStringSubmatch(envVar)
			if matches != nil {
				for i, name := range r.SubexpNames() {
					if i != 0 && name != "" && matches[i] != "" { // Ignore the whole regexp match and unnamed groups
						switch name {
						case "instance":
							targetInstance = matches[i]
						case "key":
							envVarKey = matches[i]
						case "value":
							envVarValue = matches[i]
						}
					}
				}
			}
			if targetInstance == "" {
				targetInstance = instanceName
			}
			parsedEnvVar := &environmentVariable{
				InstanceName: targetInstance,
				Key:          envVarKey,
				Value:        envVarValue,
			}
			instanceEnvVars = append(instanceEnvVars, parsedEnvVar)
		}
	}
	var mainNode *dependencyTreeNode
	mainNode = &dependencyTreeNode{
		Instance:  instanceName,
		MetaData:  cellImageMetadata,
		IsRunning: false,
		IsShared:  false,
	}
	rootNodeDependencies := map[string]*dependencyInfo{}
	for _, link := range parsedDependencyLinks {
		rootNodeDependencies[link.DependencyAlias] = &dependencyInfo{
			InstanceName: link.DependencyInstance,
		}
	}
	if err = cli.ExecuteTask(fmt.Sprintf("Starting main instance %v", util.Bold(instanceName)), "Failed to extract cell image",
		"", func() error {
			err = startCellInstance(cli, imageDir, instanceName, mainNode, instanceEnvVars, startDependencies,
				rootNodeDependencies, shareDependencies)
			return err
		}); err != nil {
		return err
	}
	util.PrintSuccessMessage(fmt.Sprintf("Successfully deployed cell image: %s", util.Bold(cellImageTag)))
	util.PrintWhatsNextMessage("list running cells", "cellery list instances")
	return nil
}

func startCellInstance(cli cli.Cli, imageDir string, instanceName string, runningNode *dependencyTreeNode,
	envVars []*environmentVariable, startDependencies bool, dependencyLinks map[string]*dependencyInfo,
	shareDependencies bool) error {
	defer os.Remove(imageDir)
	imageTag := fmt.Sprintf("%s/%s:%s", runningNode.MetaData.Organization, runningNode.MetaData.Name,
		runningNode.MetaData.Version)
	balFileName, err := util.GetSourceFileName(filepath.Join(imageDir, src))
	if err != nil {
		return fmt.Errorf("failed to find source file in Image %s due to %v", imageTag, err)
	}
	balFilePath := filepath.Join(imageDir, src, balFileName)
	var tempRunFileName string
	if tempRunFileName, err = util.CreateTempExecutableBalFile(balFilePath, "run"); err != nil {
		return fmt.Errorf("error creating temporarily executable bal file, %v", err)
	}
	var balEnvVars []ballerina.EnvironmentVariable
	balEnvVars = append(balEnvVars, ballerina.EnvironmentVariable{
		Key:   celleryImageDirEnvVar,
		Value: imageDir})
	// Export environment variables defined by user for dependent instances
	for _, envVar := range envVars {
		if !(envVar.InstanceName == "" || envVar.InstanceName == instanceName) {
			balEnvVars = append(balEnvVars, ballerina.EnvironmentVariable{
				Key:   celleryEnvVarPrefix + envVar.InstanceName + "." + envVar.Key,
				Value: envVar.Value,
			})
		}
	}
	var runCommandArgs []string
	if runCommandArgs, err = runCmdArgs(instanceName, tempRunFileName, dependencyLinks, envVars, runningNode, startDependencies, shareDependencies); err != nil {
		return fmt.Errorf("failed to get run command arguements, %v", err)
	}
	cli.BalExecutor().Run(tempRunFileName, imageDir, runCommandArgs, balEnvVars)
	if err = os.Remove(tempRunFileName); err != nil {
		return fmt.Errorf("error removing temp run file %s", tempRunFileName)
	}
	return nil
}

// extractImage extracts the image into a temporary directory and returns the path.
// Cleaning the path after finishing your work is your responsibility.
func ExtractImage(cellImage *image.CellImage, pullIfNotPresent bool) (string, error) {
	var err error
	repoLocation := filepath.Join(util.UserHomeDir(), celleryHome, "repo", cellImage.Organization,
		cellImage.ImageName, cellImage.ImageVersion)
	zipLocation := filepath.Join(repoLocation, cellImage.ImageName+cellImageExt)
	// Pull image if not exist
	var imageExists bool
	if imageExists, err = util.FileExists(zipLocation); err != nil {
		return "", err
	}
	if !imageExists {
		if pullIfNotPresent {
			cellImageTag := cellImage.Registry + "/" + cellImage.Organization + "/" + cellImage.ImageName +
				":" + cellImage.ImageVersion
			RunPull(cellImageTag, true, "", "")
		} else {
			return "", fmt.Errorf("image %s/%s:%s not present on the local repository", cellImage.Organization,
				cellImage.ImageName, cellImage.ImageVersion)
		}
	}
	// Unzipping image to a temporary location
	celleryHomeTmp := path.Join(util.UserHomeDir(), celleryHome, "tmp")
	if _, err := os.Stat(celleryHomeTmp); os.IsNotExist(err) {
		os.Mkdir(celleryHomeTmp, 0755)
	}
	var tempPath string
	if tempPath, err = ioutil.TempDir(celleryHomeTmp, "cellery-cell-image"); err != nil {
		return "", err
	}
	if err = util.Unzip(zipLocation, tempPath); err != nil {
		return "", err
	}
	return tempPath, nil
}

// runCmdArgs returns the run command arguments.
func runCmdArgs(instanceName, tempRunFileName string, dependencyLinks map[string]*dependencyInfo, envVars []*environmentVariable, runningNode *dependencyTreeNode, startDependencies, shareDependencies bool) ([]string, error) {
	var err error
	// Preparing the run command arguments
	cmdArgs := []string{"run"}
	for _, envVar := range envVars {
		// Setting root instance environment variables
		if envVar.InstanceName == "" || envVar.InstanceName == instanceName {
			cmdArgs = append(cmdArgs, "-e", envVar.Key+"="+envVar.Value)
		}
	}
	var imageNameStruct = &dependencyInfo{
		Organization: runningNode.MetaData.Organization,
		Name:         runningNode.MetaData.Name,
		Version:      runningNode.MetaData.Version,
		InstanceName: instanceName,
		IsRoot:       true,
	}
	var iName []byte
	if iName, err = json.Marshal(imageNameStruct); err != nil {
		return cmdArgs, fmt.Errorf("error in generating cellery:CellImageName construct, %v", err)
	}
	// Preparing the dependency instance map
	dependencyLinksJson, err := json.Marshal(dependencyLinks)
	if err != nil {
		return cmdArgs, fmt.Errorf("failed to prepare dependency info map %v", err)
	}
	var startDependenciesFlag = "false"
	if startDependencies {
		startDependenciesFlag = "true"
	}
	var shareDependenciesFlag = "false"
	if shareDependencies {
		shareDependenciesFlag = "true"
	}
	cmdArgs = append(cmdArgs, tempRunFileName, "run", string(iName), string(dependencyLinksJson),
		startDependenciesFlag, shareDependenciesFlag)
	return cmdArgs, nil
}

// dependencyAliasLink is used to store the link information provided by the user
type dependencyAliasLink struct {
	Instance           string
	DependencyAlias    string
	DependencyInstance string
	IsRunning          bool
}

// environmentVariable is used to store the environment variables to be passed to the instances
type environmentVariable struct {
	InstanceName string
	Key          string
	Value        string
}

// dependencyTreeNode is used as a node of the dependency tree
type dependencyTreeNode struct {
	Mux          sync.Mutex
	Instance     string
	MetaData     *image.MetaData
	Dependencies map[string]*dependencyTreeNode
	IsShared     bool
	IsRunning    bool
}

// dependencyInfo is used to pass the dependency information to Ballerina
type dependencyInfo struct {
	Organization string `json:"org"`
	Name         string `json:"name"`
	Version      string `json:"ver"`
	InstanceName string `json:"instanceName"`
	IsRoot       bool   `json:"isRoot"`
}

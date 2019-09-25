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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

// RunRun starts Cell instance (along with dependency instances if specified by the user)
// This also support linking instances to parts of the dependency tree
// This command also strictly validates whether the requested Cell (and the dependencies are valid)

const celleryEnvVar = "cellery_env_"

func RunRun(cellImageTag string, instanceName string, startDependencies bool, shareDependencies bool,
	dependencyLinks []string, envVars []string, assumeYes bool) {
	spinner := util.StartNewSpinner("Extracting Cell Image " + util.Bold(cellImageTag))
	parsedCellImage, err := image.ParseImageTag(cellImageTag)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
	}
	imageDir, err := ExtractImage(parsedCellImage, true, spinner)
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred while extracting image", err)
	}
	defer func() {
		err = os.RemoveAll(imageDir)
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error occurred while cleaning up", err)
		}
	}()

	// Reading Cell Image metadata
	spinner.SetNewAction("Reading Image " + util.Bold(cellImageTag))
	metadataFileContent, err := ioutil.ReadFile(filepath.Join(imageDir, constants.ZIP_ARTIFACTS, "cellery",
		"metadata.json"))
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred while reading Image metadata", err)
	}
	cellImageMetadata := &image.MetaData{}
	err = json.Unmarshal(metadataFileContent, cellImageMetadata)
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred while reading Image metadata", err)
	}
	var parsedDependencyLinks []*dependencyAliasLink
	if len(dependencyLinks) > 0 {
		// Parsing the dependency links list
		spinner.SetNewAction("Parsing dependency links")
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
		spinner.SetNewAction("Parsing environment variables")
		for _, envVar := range envVars {
			var targetInstance string
			var envVarKey string
			var envVarValue string

			// Parsing the environment variable
			r := regexp.MustCompile(fmt.Sprintf("^%s$", constants.CLI_ARG_ENV_VAR_PATTERN))
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

	spinner.SetNewAction("Starting main instance " + util.Bold(instanceName))
	err = startCellInstance(imageDir, instanceName, mainNode, instanceEnvVars, startDependencies,
		rootNodeDependencies, shareDependencies)
	if err != nil {
		util.ExitWithErrorMessage("Failed to start Cell instance "+instanceName, err)
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully deployed cell image: %s", util.Bold(cellImageTag)))
	util.PrintWhatsNextMessage("list running cells", "cellery list instances")
}

func startCellInstance(imageDir string, instanceName string, runningNode *dependencyTreeNode,
	envVars []*environmentVariable, startDependencies bool, dependencyLinks map[string]*dependencyInfo,
	shareDependencies bool) error {
	imageTag := fmt.Sprintf("%s/%s:%s", runningNode.MetaData.Organization, runningNode.MetaData.Name,
		runningNode.MetaData.Version)
	balFileName, err := util.GetSourceFileName(filepath.Join(imageDir, constants.ZIP_BALLERINA_SOURCE))
	if err != nil {
		return fmt.Errorf("failed to find source file in Image %s due to %v", imageTag, err)
	}
	balFilePath := filepath.Join(imageDir, constants.ZIP_BALLERINA_SOURCE, balFileName)

	containsRunFunction, err := util.RunMethodExists(balFilePath)
	if err != nil {
		return fmt.Errorf("failed to check whether run function exists in Image %s due to %v", imageTag, err)
	}
	// run function will be mandatory for all cells
	if containsRunFunction {
		// Preparing the dependency instance map
		dependencyLinksJson, err := json.Marshal(dependencyLinks)
		if err != nil {
			return fmt.Errorf("failed to start the Image %s due to %v", imageTag, err)
		}
		tempRunFileName, err := util.CreateTempExecutableBalFile(balFilePath, "run")
		if err != nil {
			util.ExitWithErrorMessage("Error executing ballerina file", err)
		}
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
		iName, err := json.Marshal(imageNameStruct)
		if err != nil {
			util.ExitWithErrorMessage("Error in generating cellery:CellImageName construct", err)
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

		// Calling the run function
		moduleMgr := &util.BLangManager{}
		exePath, err := moduleMgr.GetExecutablePath()
		if err != nil {
			util.ExitWithErrorMessage("Failed to get executable path", err)
		}

		cmd := &exec.Cmd{}

		if exePath != "" {
			cmd = exec.Command(exePath+"ballerina", cmdArgs...)
		} else {
			currentDir, err := os.Getwd()
			if err != nil {
				util.ExitWithErrorMessage("Error in determining working directory", err)
			}

			//Retrieve the cellery cli docker instance status.
			cmdDockerPs := exec.Command("docker", "ps", "--filter", "label=ballerina-runtime="+version.BuildVersion(),
				"--filter", "label=currentDir="+currentDir, "--filter", "status=running", "--format", "{{.ID}}")

			containerId, err := cmdDockerPs.Output()
			if err != nil {
				util.ExitWithErrorMessage("Docker Run Error", err)
			}

			if string(containerId) == "" {

				cmdDockerRun := exec.Command("docker", "run", "-d", "-l", "ballerina-runtime="+version.BuildVersion(),
					"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".ballerina,target=/home/cellery/.ballerina",
					"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".cellery,target=/home/cellery/.cellery",
					"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".kube,target=/home/cellery/.kube",
					"wso2cellery/ballerina-runtime:"+version.BuildVersion(), "sleep", "600",
				)

				containerId, err = cmdDockerRun.Output()
				if err != nil {
					util.ExitWithErrorMessage("Docker Run Error", err)
				}
				time.Sleep(5 * time.Second)
			}

			cliUser, err := user.Current()
			if err != nil {
				util.ExitWithErrorMessage("Error while retrieving the current user", err)
			}

			exeUid := constants.CELLERY_DOCKER_CLI_USER_ID

			if cliUser.Uid != constants.CELLERY_DOCKER_CLI_USER_ID && runtime.GOOS == "linux" {
				cmdUserExist := exec.Command("docker", "exec", strings.TrimSpace(string(containerId)),
					"id", "-u", cliUser.Username)
				_, errUserExist := cmdUserExist.Output()
				if errUserExist != nil {
					cmdUserAdd := exec.Command("docker", "exec", strings.TrimSpace(string(containerId)), "useradd", "-m",
						"-d", "/home/cellery", "--uid", cliUser.Uid, cliUser.Username)

					_, errUserAdd := cmdUserAdd.Output()
					if errUserAdd != nil {
						util.ExitWithErrorMessage("Error in adding Cellery execution user", errUserAdd)
					}
				}
				exeUid = cliUser.Uid
			}

			cmdArgs = append(cmdArgs, "-e", constants.CELLERY_IMAGE_DIR_ENV_VAR+"="+imageDir)

			re := regexp.MustCompile(`^.*cellery-cell-image`)
			tempRunFileName = re.ReplaceAllString(tempRunFileName, "/home/cellery/.cellery/tmp/cellery-cell-image")
			dockerImageDir := re.ReplaceAllString(imageDir, "/home/cellery/.cellery/tmp/cellery-cell-image")

			cmd = exec.Command("docker", "exec", "-e", constants.CELLERY_IMAGE_DIR_ENV_VAR+"="+dockerImageDir)
			shellEnvs := os.Environ()
			// check if any env var prepended with `CELLERY` exists. If so, set them to docker exec command.
			if len(shellEnvs) != 0 {
				for _, shellEnv := range shellEnvs {
					if strings.HasPrefix(shellEnv, "CELLERY") {
						cmd.Args = append(cmd.Args, "-e", shellEnv)
					}
				}
			}
			// set any explicitly passed env vars in cellery run command to the docker exec.
			// This will override any env vars with identical names (prefixed with 'CELLERY') set previously.
			if len(envVars) != 0 {
				for _, envVar := range envVars {
					if envVar.InstanceName == "" || envVar.InstanceName == instanceName {
						cmd.Args = append(cmd.Args, "-e", envVar.Key+"="+envVar.Value)
					}
				}
			}
			cmd.Args = append(cmd.Args, "-w", "/home/cellery", "-u", exeUid,
				strings.TrimSpace(string(containerId)), constants.DOCKER_CLI_BALLERINA_EXECUTABLE_PATH, "run", tempRunFileName, "run",
				string(iName), string(dependencyLinksJson))
		}
		defer os.Remove(imageDir)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, constants.CELLERY_IMAGE_DIR_ENV_VAR+"="+imageDir)
		// Export environment variables defined by user for dependent instances
		for _, envVar := range envVars {
			if !(envVar.InstanceName == "" || envVar.InstanceName == instanceName) {
				cmd.Env = append(cmd.Env, celleryEnvVar+envVar.InstanceName+"."+envVar.Key+"="+envVar.Value)
			}
		}
		stdoutReader, _ := cmd.StdoutPipe()
		stdoutScanner := bufio.NewScanner(stdoutReader)
		go func() {
			for stdoutScanner.Scan() {
				fmt.Printf("\r\x1b[2K\033[36m%s\033[m\n", stdoutScanner.Text())
			}
		}()
		stderrReader, _ := cmd.StderrPipe()
		stderrScanner := bufio.NewScanner(stderrReader)
		go func() {
			for stderrScanner.Scan() {
				fmt.Printf("\r\x1b[2K\033[36m%s\033[m\n", stderrScanner.Text())
			}
		}()
		err = cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to execute run method in Cell instance %s due to %v", instanceName, err)
		}
		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("failed to execute run method in Cell instance %s due to %v", instanceName, err)
		}
		_ = os.Remove(tempRunFileName)
	} else {
		return fmt.Errorf("run function does not exist in Image %s", imageTag)
	}
	return nil
}

// extractImage extracts the image into a temporary directory and returns the path.
// Cleaning the path after finishing your work is your responsibility.
func ExtractImage(cellImage *image.CellImage, pullIfNotPresent bool, spinner *util.Spinner) (string, error) {
	repoLocation := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo", cellImage.Organization,
		cellImage.ImageName, cellImage.ImageVersion)
	zipLocation := filepath.Join(repoLocation, cellImage.ImageName+constants.CELL_IMAGE_EXT)

	// Pull image if not exist
	imageExists, err := util.FileExists(zipLocation)
	if err != nil {
		return "", err
	}
	if !imageExists {
		if pullIfNotPresent {
			if spinner != nil {
				spinner.Pause()
			}
			cellImageTag := cellImage.Registry + "/" + cellImage.Organization + "/" + cellImage.ImageName +
				":" + cellImage.ImageVersion
			RunPull(cellImageTag, true, "", "")
			fmt.Println()
			if spinner != nil {
				spinner.Resume()
			}
		} else {
			return "", fmt.Errorf("image %s/%s:%s not present on the local repository", cellImage.Organization,
				cellImage.ImageName, cellImage.ImageVersion)
		}
	}

	// Unzipping image to a temporary location
	celleryHomeTmp := path.Join(util.UserHomeDir(), constants.CELLERY_HOME, "tmp")
	if _, err := os.Stat(celleryHomeTmp); os.IsNotExist(err) {
		os.Mkdir(celleryHomeTmp, 0755)
	}

	tempPath, err := ioutil.TempDir(celleryHomeTmp, "cellery-cell-image")
	if err != nil {
		return "", err
	}
	err = util.Unzip(zipLocation, tempPath)
	if err != nil {
		return "", nil
	}
	return tempPath, nil
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

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
	"os/user"
	"runtime"
	"strconv"

	"github.com/cellery-io/sdk/components/cli/pkg/image"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/version"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunRun starts Cell instance (along with dependency instances if specified by the user)\
func RunTest(cellImageTag string, instanceName string, startDependencies bool, shareDependencies bool,
	dependencyLinks []string, envVars []string, assumeYes bool, debug bool, verbose bool, incell bool) {
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
		spinner.SetNewAction("Validating dependency links")
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
			cellInstance, err := kubectl.GetCell(dependencyLink.DependencyInstance)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				spinner.Stop(false)
				util.ExitWithErrorMessage("Error occurred while validating dependency links", err)
			} else {
				dependencyLink.IsRunning = err == nil && cellInstance.CellStatus.Status == "Ready"
				parsedDependencyLinks = append(parsedDependencyLinks, dependencyLink)
				continue
			}
			compositeInstance, err := kubectl.GetComposite(dependencyLink.DependencyInstance)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				spinner.Stop(false)
				util.ExitWithErrorMessage("Error occurred while validating dependency links", err)
			} else {
				dependencyLink.IsRunning = err == nil && compositeInstance.CompositeStatus.Status == "Ready"
				parsedDependencyLinks = append(parsedDependencyLinks, dependencyLink)
			}
		}
	}

	instanceEnvVars := map[string][]*environmentVariable{}
	if len(envVars) > 0 {
		// Parsing environment variables
		spinner.SetNewAction("Validating environment variables")
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
				Key:   envVarKey,
				Value: envVarValue,
			}

			// Validating whether the instance of the environment var is provided as an instance of a link
			if targetInstance != instanceName {
				isInstanceProvided := false
				isInstanceToBeStarted := false
				for _, link := range parsedDependencyLinks {
					if targetInstance == link.DependencyInstance {
						isInstanceProvided = true
						isInstanceToBeStarted = !link.IsRunning
						break
					}
				}
				if !isInstanceProvided {
					spinner.Stop(false)
					util.ExitWithErrorMessage("Invalid environment variable",
						fmt.Errorf("the instance of the environment variables should be provided as a "+
							"dependency link, instance %s of the environment variable %s not found", targetInstance,
							parsedEnvVar.Key))
				} else if !isInstanceToBeStarted {
					spinner.Stop(false)
					util.ExitWithErrorMessage("Invalid environment variable",
						fmt.Errorf("the instance of the environment should be an instance to be "+
							"created, instance %s is already available in the runtime", targetInstance))
				}
			}

			if _, hasKey := instanceEnvVars[targetInstance]; !hasKey {
				instanceEnvVars[targetInstance] = []*environmentVariable{}
			}
			instanceEnvVars[targetInstance] = append(instanceEnvVars[targetInstance], parsedEnvVar)
		}
	}

	var mainNode *dependencyTreeNode
	mainNode = &dependencyTreeNode{
		Instance:  instanceName,
		MetaData:  cellImageMetadata,
		IsRunning: false,
		IsShared:  false,
	}

	//rootNodeDependencies := map[string]*dependencyTreeNode{}
	rootNodeDependencies := map[string]*dependencyInfo{}
	for _, link := range parsedDependencyLinks {
		rootNodeDependencies[link.DependencyAlias] = &dependencyInfo{
			InstanceName: link.DependencyInstance,
		}
	}

	spinner.Stop(true)
	util.PrintSuccessMessage("Starting execution of tests for " + util.Bold(cellImageTag) + "...")
	err = startTestCellInstance(imageDir, instanceName, mainNode, instanceEnvVars[instanceName], startDependencies, rootNodeDependencies, verbose, debug, incell, assumeYes)
	if err != nil {
		util.ExitWithErrorMessage("Failed to test Cell instance "+instanceName, err)
	}
	util.PrintSuccessMessage(fmt.Sprintf("Completed running tests for instance %s", util.Bold(instanceName)))
}

func startTestCellInstance(imageDir string, instanceName string, runningNode *dependencyTreeNode,
	envVars []*environmentVariable, startDependencies bool, dependencyLinks map[string]*dependencyInfo, verbose bool, debug bool, incell bool, assumeYes bool) error {
	imageTag := fmt.Sprintf("%s/%s:%s", runningNode.MetaData.Organization, runningNode.MetaData.Name,
		runningNode.MetaData.Version)
	balFileName, err := util.GetSourceFileName(filepath.Join(imageDir, constants.ZIP_BALLERINA_SOURCE))
	if err != nil {
		return fmt.Errorf("failed to find source file in Image %s due to %v", imageTag, err)
	}
	balFilePath := filepath.Join(imageDir, constants.ZIP_BALLERINA_SOURCE, balFileName)

	containsTestFunction, err := util.TestMethodExists(balFilePath)
	if err != nil {
		return fmt.Errorf("failed to check whether test function exists in Image %s due to %v", imageTag, err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		util.ExitWithErrorMessage("Error in determining working directory", err)
	}

	// Preparing the dependency instance map
	dependencyLinksJson, err := json.Marshal(dependencyLinks)
	if err != nil {
		return fmt.Errorf("failed to start the Cell Image %s due to %v", imageTag, err)
	}

	// Calling the test function
	moduleMgr := &util.BLangManager{}
	exePath, err := moduleMgr.GetExecutablePath()
	if err != nil {
		util.ExitWithErrorMessage("Failed to get executable path", err)
	}
	cmd := &exec.Cmd{}
	verboseMode := strconv.FormatBool(verbose)

	testsPath := filepath.Join(imageDir, constants.ZIP_TESTS)
	testsRoot := filepath.Join(currentDir, "target")
	testsTarget := filepath.Join(testsRoot, instanceName)
	isTestDirExists, _ := util.IsExists(testsPath)
	telepresenceYamlPath := filepath.Join(imageDir, "telepresence.yaml")

	if isTestDirExists {
		if exePath == "" {
			util.ExitWithErrorMessage("Ballerina not found. Please install Ballerina to run inline tests", err)
		}
		err = util.CleanOrCreateDir(testsTarget)
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while creating the cell image", err)
		}
		fileCopyError := util.CopyDir(testsPath, filepath.Join(testsTarget, constants.ZIP_TESTS))
		if fileCopyError != nil {
			util.ExitWithErrorMessage("Error occurred while copying tests", err)
		}
		err := os.Chdir(testsTarget)
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while changing working directory", err)
		}
	}

	if containsTestFunction {
		tempTestFileName, err := util.CreateTempExecutableBalFile(balFilePath, "test")
		if err != nil {
			util.ExitWithErrorMessage("Error executing ballerina file", err)
		}
		// Preparing the run command arguments
		cmdArgs := []string{"run"}
		var imageNameStruct = &dependencyInfo{
			Organization: runningNode.MetaData.Organization,
			Name:         runningNode.MetaData.Name,
			Version:      runningNode.MetaData.Version,
			InstanceName: instanceName,
		}
		iName, err := json.Marshal(imageNameStruct)
		if err != nil {
			util.ExitWithErrorMessage("Error in generating cellery:CellImageName construct", err)
		}
		var startDependenciesFlag = "false"
		if startDependencies {
			startDependenciesFlag = "true"
		}
		cmdArgs = append(cmdArgs, tempTestFileName, "test", string(iName), string(dependencyLinksJson), startDependenciesFlag)
		os.Mkdir(currentDir+string(os.PathSeparator)+"logs", 0777)
		defer os.Remove(imageDir)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, constants.CELLERY_IMAGE_DIR_ENV_VAR+"="+imageDir)

		if debug {
			content := []string{fmt.Sprintf("DEBUG_MODE=\"%s\"\n", verboseMode)}
			content = append(content, fmt.Sprintf("IMAGE_ORG=\"%s\"\n", runningNode.MetaData.Organization))
			content = append(content, fmt.Sprintf("IMAGE_NAME=\"%s\"\n", runningNode.MetaData.Name))
			content = append(content, fmt.Sprintf("IMAGE_VERSION=\"%s\"\n", runningNode.MetaData.Version))
			content = append(content, fmt.Sprintf("INSTANCE_NAME=\"%s\"\n", instanceName))
			for _, envVar := range envVars {
				content = append(content, fmt.Sprintf(envVar.Key+"=\"%s\"\n", envVar.Value))
			}

			ballerinaConf := filepath.Join(util.UserHomeCelleryDir(), constants.TMP, constants.BALLERINA_CONF)
			isExistsBalConf, err := util.IsExists(ballerinaConf)
			if err != nil {
				util.ExitWithErrorMessage("error while checking if "+ballerinaConf+" exists", err)
			}
			if isExistsBalConf {
				err := os.Remove(ballerinaConf)
				if err != nil {
					util.ExitWithErrorMessage("error while removing "+ballerinaConf+" file", err)
				}
			}
			_, err = os.Create(ballerinaConf)
			if err != nil {
				util.ExitWithErrorMessage("error while creating "+ballerinaConf+" file", err)
			}

			f, err := os.OpenFile(ballerinaConf, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				util.ExitWithErrorMessage("error while opening "+ballerinaConf, err)
			}

			defer f.Close()

			for _, element := range content {
				if _, err = f.WriteString(element); err != nil {
					util.ExitWithErrorMessage("error while writing properties to "+ballerinaConf, err)
				}
			}

			util.PrintSuccessMessage("Note that the following configuration should be added to the launch configuration " +
				"before starting the debug session\n" +
				fmt.Sprintf("--------------------------------------------------------\n\n") +
				fmt.Sprintf("\"commandOptions\": [\"--config\", \"%s\"],\n\n", ballerinaConf) +
				fmt.Sprintln("--------------------------------------------------------"))
		} else {
			cmd.Env = append(cmd.Env, constants.CELLERY_IMAGE_DIR_ENV_VAR+"="+imageDir)
			cmd.Env = append(cmd.Env, fmt.Sprintf("DEBUG_MODE=%s", verboseMode))
			cmd.Env = append(cmd.Env, fmt.Sprintf("INCELL=%s", strconv.FormatBool(incell)))
			cmd.Env = append(cmd.Env, fmt.Sprintf("IMAGE_ORG=%s", runningNode.MetaData.Organization))
			cmd.Env = append(cmd.Env, fmt.Sprintf("IMAGE_NAME=%s", runningNode.MetaData.Name))
			cmd.Env = append(cmd.Env, fmt.Sprintf("IMAGE_VERSION=%s", runningNode.MetaData.Version))
			cmd.Env = append(cmd.Env, fmt.Sprintf("INSTANCE_NAME=%s", instanceName))
		}

		if !assumeYes {
			fmt.Printf("%s Do you wish to continue with testing above Cell instances (Y/n)? ", util.YellowBold("?"))
			reader := bufio.NewReader(os.Stdin)
			confirmation, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			if strings.ToLower(strings.TrimSpace(confirmation)) == "n" {
				return fmt.Errorf("Cell testing aborted")
			}
		}

		for _, envVar := range envVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf(envVar.Key+"=%s", envVar.Value))
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

		if exePath != "" {
			ballerinaArgs := []string{exePath + "ballerina"}
			ballerinaArgs = append(ballerinaArgs, cmdArgs...)
			RunTelepresenceTests(incell, cmd, ballerinaArgs, imageDir, instanceName, debug)
		} else {
			//Retrieve the cellery cli docker instance status.
			cmdDockerPs := exec.Command("docker", "ps", "--filter", "label=ballerina-runtime="+version.BuildVersion(),
				"--filter", "label=currentDir="+currentDir, "--filter", "status=running", "--format", "{{.ID}}")

			containerId, err := cmdDockerPs.Output()
			if err != nil {
				util.ExitWithErrorMessage("Docker Run Error", err)
			}
			if string(containerId) == "" {

				cmdDockerRun := exec.Command("docker", "run", "-e", constants.TEST_DEGUB_FLAG+"="+verboseMode,
					"-d", "-l", "ballerina-runtime="+version.BuildVersion(),
					"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".ballerina,target=/home/cellery/.ballerina",
					"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".cellery,target=/home/cellery/.cellery",
					"--mount", "type=bind,source="+util.UserHomeDir()+string(os.PathSeparator)+".kube,target=/home/cellery/.kube",
					"--mount", "type=bind,source="+currentDir+string(os.PathSeparator)+"logs,target=/home/cellery/logs",
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
			tempTestFileName = re.ReplaceAllString(tempTestFileName, "/home/cellery/.cellery/tmp/cellery-cell-image")
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
					cmd.Args = append(cmd.Args, "-e", envVar.Key+"="+envVar.Value)
				}
			}
			cmd.Args = append(cmd.Args, "-w", "/home/cellery", "-u", exeUid,
				strings.TrimSpace(string(containerId)), constants.DOCKER_CLI_BALLERINA_EXECUTABLE_PATH, "run", tempTestFileName, "run",
				string(iName), string(dependencyLinksJson))

			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("failed to execute test method in Cell instance %s due to %v", instanceName, err)
			}
			err = cmd.Wait()
		}

		_ = os.Remove(tempTestFileName)
		if err != nil {
			return fmt.Errorf("failed to execute test method in Cell instance %s due to %v", instanceName, err)
		}
	} else {
		isBallerinaProject, err := util.IsExists(filepath.Join(currentDir, "Ballerina.toml"))
		if err != nil {
			util.ExitWithErrorMessage("failed to check if Ballerina.toml exists.", err)
		}
		if !isBallerinaProject {
			cmd = exec.Command(exePath+"ballerina", "init")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("error occurred while initializing ballerina project for tests", err)
			}
		}
		cmd = &exec.Cmd{}
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, constants.CELLERY_IMAGE_DIR_ENV_VAR+"="+imageDir)
		if debug {
			content := []string{fmt.Sprintf("DEBUG_MODE=\"%s\"\n", verboseMode)}
			content = append(content, fmt.Sprintf("IMAGE_ORG=\"%s\"\n", runningNode.MetaData.Organization))
			content = append(content, fmt.Sprintf("IMAGE_NAME=\"%s\"\n", runningNode.MetaData.Name))
			content = append(content, fmt.Sprintf("IMAGE_VERSION=\"%s\"\n", runningNode.MetaData.Version))
			content = append(content, fmt.Sprintf("INSTANCE_NAME=\"%s\"\n", instanceName))
			for _, envVar := range envVars {
				content = append(content, fmt.Sprintf(envVar.Key+"=\"%s\"\n", envVar.Value))
			}

			ballerinaConf := filepath.Join(util.UserHomeCelleryDir(), constants.TMP, constants.BALLERINA_CONF)
			isExistsBalConf, err := util.IsExists(ballerinaConf)
			if err != nil {
				util.ExitWithErrorMessage("error while checking if "+ballerinaConf+" exists", err)
			}
			if isExistsBalConf {
				err := os.Remove(ballerinaConf)
				if err != nil {
					util.ExitWithErrorMessage("error while removing "+ballerinaConf+" file", err)
				}
			}
			_, err = os.Create(ballerinaConf)
			if err != nil {
				util.ExitWithErrorMessage("error while creating "+ballerinaConf+" file", err)
			}

			f, err := os.OpenFile(ballerinaConf, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				util.ExitWithErrorMessage("error while opening "+ballerinaConf, err)
			}

			defer f.Close()

			for _, element := range content {
				if _, err = f.WriteString(element); err != nil {
					util.ExitWithErrorMessage("error while writing properties to "+ballerinaConf, err)
				}
			}

			util.PrintSuccessMessage("Note that the following configuration should be added to the launch configuration " +
				"before starting the debug session\n" +
				fmt.Sprintf("--------------------------------------------------------\n\n") +
				fmt.Sprintf("\"commandOptions\": [\"--config\", \"%s\"],\n\n", ballerinaConf) +
				fmt.Sprintln("--------------------------------------------------------"))

		} else {
			cmd.Env = append(cmd.Env, constants.CELLERY_IMAGE_DIR_ENV_VAR+"="+imageDir)
			cmd.Env = append(cmd.Env, fmt.Sprintf("DEBUG_MODE=%s", verboseMode))
			cmd.Env = append(cmd.Env, fmt.Sprintf("IMAGE_ORG=%s", runningNode.MetaData.Organization))
			cmd.Env = append(cmd.Env, fmt.Sprintf("IMAGE_NAME=%s", runningNode.MetaData.Name))
			cmd.Env = append(cmd.Env, fmt.Sprintf("IMAGE_VERSION=%s", runningNode.MetaData.Version))
			cmd.Env = append(cmd.Env, fmt.Sprintf("INSTANCE_NAME=%s", instanceName))
		}

		if !assumeYes {
			fmt.Printf("%s Do you wish to continue with testing above Cell instances (Y/n)? ", util.YellowBold("?"))
			reader := bufio.NewReader(os.Stdin)
			confirmation, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			if strings.ToLower(strings.TrimSpace(confirmation)) == "n" {
				return fmt.Errorf("Cell testing aborted")
			}
		}
		var cmdArgs []string
		if incell {
			cmdArgs = []string{exePath + "ballerina", "test", "--groups", "incell"}
		} else {
			cmdArgs = []string{exePath + "ballerina", "test", "--disable-groups", "incell"}
		}
		RunTelepresenceTests(incell, cmd, cmdArgs, imageDir, instanceName, debug)
	}
	StopTelepresence(telepresenceYamlPath)

	return nil
}
func StopTelepresence(filepath string) error {
	err := kubectl.DeleteFile(filepath)
	if err != nil {
		return fmt.Errorf("error occurred while stopping telepresence", err)
	}
	return nil
}

func RunTelepresenceTests(incell bool, cmd *exec.Cmd, cmdArgs []string, imageDir string, instanceName string, debug bool) {
	var srcYamlFile string
	dstYamlFile := filepath.Join(imageDir, "telepresence.yaml")
	var deploymentName string
	var stsDeploymentName string
	if incell {
		srcYamlFile = filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS, constants.TELEPRESENCE, "telepresence-deployment.yaml")
		err := util.CopyFile(srcYamlFile, dstYamlFile)
		if err != nil {
			util.ExitWithErrorMessage(fmt.Sprintf("error while copying telepresene k8s artifact to %s", imageDir), err)
		}
		util.ReplaceInFile(dstYamlFile, "{{cell}}", instanceName, -1)
		deploymentName = instanceName + "--telepresence"
	} else {
		srcYamlFile = filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS, constants.TELEPRESENCE, "telepresence-cell.yaml")
		err := util.CopyFile(srcYamlFile, dstYamlFile)
		if err != nil {
			util.ExitWithErrorMessage(fmt.Sprintf("error while copying telepresene k8s artifact to %s", imageDir), err)
		}
		deploymentName = "telepresence--telepresence-deployment"
		stsDeploymentName = "telepresence--sts-deployment"

	}

	kubectl.ApplyFile(dstYamlFile)
	kubectl.WaitForDeployment("available", 900, "deployment.extensions/"+deploymentName, "default")

	if !incell {
		kubectl.WaitForDeployment("available", 900, "deployment.extensions/"+stsDeploymentName, "default")
	}

	telepresenceExecPath := filepath.Join(util.CelleryInstallationDir(), constants.TELEPRESENCE_EXEC_PATH, "/telepresence")
	var telArgs = []string{telepresenceExecPath, "--deployment", deploymentName}
	if !debug {
		telArgs = append(telArgs, "--run")
		telArgs = append(telArgs, cmdArgs...)
	}

	cmd.Path = telepresenceExecPath
	cmd.Args = telArgs
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		StopTelepresence(dstYamlFile)
		util.ExitWithErrorMessage("error occurred while running tests", err)
	}
	err = cmd.Wait()
	if err != nil {
		StopTelepresence(dstYamlFile)
		util.ExitWithErrorMessage("error occurred while waiting for tests to complete", err)
	}
}

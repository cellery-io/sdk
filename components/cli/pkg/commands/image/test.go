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
	"os/user"
	"strconv"

	"github.com/cellery-io/sdk/components/cli/cli"

	"github.com/cellery-io/sdk/components/cli/pkg/version"

	"github.com/cellery-io/sdk/components/cli/pkg/image"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

// RunRun starts Cell instance (along with dependency instances if specified by the user)\
func RunTest(cli cli.Cli, cellImageTag string, instanceName string, startDependencies bool, shareDependencies bool,
	dependencyLinks []string, envVars []string, assumeYes bool, debug bool, verbose bool, incell bool, projLocation string) {
	spinner := util.StartNewSpinner("Extracting Cell Image " + util.Bold(cellImageTag))
	parsedCellImage, err := image.ParseImageTag(cellImageTag)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while parsing cell image", err)
	}
	imageDir, err := ExtractImage(cli, parsedCellImage, true)
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
	metadataFileContent, err := ioutil.ReadFile(filepath.Join(imageDir, constants.ZipArtifacts, "cellery",
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
			cellInstance, err := kubernetes.GetCell(dependencyLink.DependencyInstance)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				spinner.Stop(false)
				util.ExitWithErrorMessage("Error occurred while validating dependency links", err)
			} else {
				dependencyLink.IsRunning = err == nil && cellInstance.CellStatus.Status == "Ready"
				parsedDependencyLinks = append(parsedDependencyLinks, dependencyLink)
				continue
			}
			compositeInstance, err := kubernetes.GetComposite(dependencyLink.DependencyInstance)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				spinner.Stop(false)
				util.ExitWithErrorMessage("Error occurred while validating dependency links", err)
			} else {
				dependencyLink.IsRunning = err == nil && compositeInstance.CompositeStatus.Status == "Ready"
				parsedDependencyLinks = append(parsedDependencyLinks, dependencyLink)
			}
		}
	}

	instanceEnvVars := []*environmentVariable{}
	if len(envVars) > 0 {
		// Parsing environment variables
		spinner.SetNewAction("Validating environment variables")
		for _, envVar := range envVars {
			var targetInstance string
			var envVarKey string
			var envVarValue string

			// Parsing the environment variable
			r := regexp.MustCompile(fmt.Sprintf("^%s$", constants.CliArgEnvVarPattern))
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

	dependencyEnv := []string{}
	rootNodeDependencies := map[string]*dependencyInfo{}
	for _, link := range parsedDependencyLinks {
		rootNodeDependencies[link.DependencyAlias] = &dependencyInfo{
			InstanceName: link.DependencyInstance,
		}
		dependencyEnv = append(dependencyEnv, link.DependencyAlias+"="+link.DependencyInstance)
	}

	spinner.Stop(true)
	err = startTestCellInstance(imageDir, instanceName, mainNode, instanceEnvVars, startDependencies,
		shareDependencies, rootNodeDependencies, verbose, debug, incell, assumeYes, projLocation, )
	if err != nil {
		util.ExitWithErrorMessage("Failed to test Cell instance "+instanceName, err)
	}
	//Remove Ballerina toml and local repo to avoid failing cellery build and run
	os.Remove(constants.BallerinaToml)
	os.Remove(constants.BallerinaConf)
	util.PrintSuccessMessage(fmt.Sprintf("Completed running tests for instance %s", util.Bold(instanceName)))
}

func startTestCellInstance(imageDir string, instanceName string, runningNode *dependencyTreeNode,
	envVars []*environmentVariable, startDependencies bool, shareDependencies bool, dependencyLinks map[string]*dependencyInfo,
	verbose bool, debug bool, incell bool, assumeYes bool, projLocation string) error {
	imageTag := fmt.Sprintf("%s/%s:%s", runningNode.MetaData.Organization, runningNode.MetaData.Name,
		runningNode.MetaData.Version)
	balFileName, err := util.GetSourceFileName(filepath.Join(imageDir, constants.ZipBallerinaSource))
	if err != nil {
		return fmt.Errorf("failed to find source file in Image %s due to %v", imageTag, err)
	}
	balFilePath := filepath.Join(imageDir, constants.ZipBallerinaSource, balFileName)

	currentDir, err := os.Getwd()
	if err != nil {
		util.ExitWithErrorMessage("Error in determining working directory", err)
	}

	// Preparing the dependency instance map
	dependencyLinksJson, err := json.Marshal(dependencyLinks)
	if err != nil {
		return fmt.Errorf("failed to start the Cell Image %s due to %v", imageTag, err)
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

	// Get Ballerina Installation if exists
	moduleMgr := &util.BLangManager{}
	exePath, err := moduleMgr.GetExecutablePath()
	if err != nil {
		util.ExitWithErrorMessage("Failed to get executable path", err)
	}
	cmd := &exec.Cmd{}
	verboseMode := strconv.FormatBool(verbose)

	telepresenceYamlPath := filepath.Join(imageDir, "telepresence.yaml")
	if !debug {
		// Construct test artifact names
		balProjectName := constants.TARGET_DIR_NAME
		balTomlPath := filepath.Join(imageDir, constants.ZIP_BALLERINA_SOURCE, constants.BALLERINA_TOML)
		testsPath := filepath.Join(imageDir, constants.ZIP_TESTS)
		testsRoot := filepath.Join(currentDir, balProjectName)
		var balModule string

		if instanceName != "" {
			balModule = filepath.Join(testsRoot, constants.ZIP_BALLERINA_SOURCE, instanceName)
		} else {
			balModule = filepath.Join(testsRoot, constants.ZIP_BALLERINA_SOURCE, constants.TEMP_TEST_MODULE)
		}
		isTestDirExists, _ := util.FileExists(testsPath)
		var isBallerinaProject bool

		//TODO: check if bal file contains @test
		if !isTestDirExists {
			return fmt.Errorf("no tests found in the cell image %v", imageTag)
		}

		err = util.RemoveDir(testsRoot)
		if err != nil {
			return fmt.Errorf("Error occurred while deleting filepath, %s", err)
		}

		// Create Ballerina project
		isBallerinaProject, err = util.FileExists(balTomlPath)
		if exePath != "" {
			cmd := exec.Command(exePath, "new", balProjectName)
			if verbose {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("error occurred while initializing ballerina project for tests %v", err)
			}
		} else {
			err = util.CleanOrCreateDir(testsRoot)
			if err != nil {
				util.ExitWithErrorMessage(fmt.Sprintf("Error occurred while creating %s", testsRoot), err)
			}
		}

		// Copy artifacts for running tests
		if isBallerinaProject {
			fileCopyError := util.CopyFile(balTomlPath, filepath.Join(testsRoot, constants.BallerinaToml))
			if fileCopyError != nil {
				return fmt.Errorf("error occurred while copying file, %s", err)
				util.ExitWithErrorMessage(fmt.Sprintf("Error occurred while copying %s", constants.BallerinaToml), err)
			}
		}
		fileCopyError := util.CopyDir(testsPath, filepath.Join(balModule, constants.ZipTests))
		if fileCopyError != nil {
			return fmt.Errorf("error occurred while copying tests folder, %s", err)
		}

		err = util.CleanAndCreateDir(filepath.Join(testsRoot, "logs"))
		if err != nil {
			return fmt.Errorf("error occurred while creating %s", filepath.Join(testsRoot, "logs"), err)
		}

		// Change working dir to Bal project
		err = os.Chdir(testsRoot)
		if err != nil {
			return fmt.Errorf("error occurred while changing working directory, %s", err)
		}

		isExistsSourceBal, err := util.FileExists(filepath.Join(balModule, filepath.Base(balFilePath)))
		if !isExistsSourceBal {
			fileCopyError := util.CopyFile(balFilePath, filepath.Join(balModule, filepath.Base(balFilePath)))
			if fileCopyError != nil {
				return fmt.Errorf("error occurred while copying source bal file, %s", err)
			}
		}

		if exePath == "" {
			mntImageDir := strings.Replace(imageDir, util.UserHomeDir(), "/home/cellery", 1)
			err = CreateBallerinaConf(string(iName), verboseMode, mntImageDir, string(dependencyLinksJson), envVars, testsRoot)
			if err != nil {
				return fmt.Errorf("error occurred while creating the ballerina.conf, %s", err)
			}
			err = RunDockerTelepresenceTests(incell, cmd, testsRoot, imageDir, instanceName)
			if err != nil{
				StopTelepresence(telepresenceYamlPath)
				return fmt.Errorf("error occured while running tests in ballerina docker image, %s", err)
			}
		} else {
			err = CreateBallerinaConf(string(iName), verboseMode, imageDir, string(dependencyLinksJson), envVars, testsRoot)
			if err != nil {
				return fmt.Errorf("Error occurred while creating the ballerina.conf, %s", err)
			}
			cmd = &exec.Cmd{}
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, constants.CelleryImageDirEnvVar+"="+imageDir)

			//if incell {
			//	cmdArgs = append(cmdArgs, "--groups", "incell")
			//} else {
			//	cmdArgs = append(cmdArgs, "--disable-groups", "incell")
			//}
			err = RunTelepresenceTests(incell, cmd, exePath, imageDir, instanceName, debug)
			if err != nil{
				StopTelepresence(telepresenceYamlPath)
				return fmt.Errorf("error occured while running tests, %s", err)
			}
		}
	} else {
		if exePath == "" {
			return fmt.Errorf("Ballerina should be installed to debug tests.")
		} else {
			balProj, err := filepath.Abs(projLocation)
			if err != nil {
				return fmt.Errorf("error occurred while retrieving absolute path of project, %s", err)
			}
			PromtConfirmation(assumeYes, balProj)
			err = CreateBallerinaConf(string(iName), verboseMode, imageDir, string(dependencyLinksJson), envVars, balProj)
			if err != nil {
				return fmt.Errorf("error occurred while creating the ballerina.conf, %s", err)
			}
			cmd = &exec.Cmd{}
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, constants.CELLERY_IMAGE_DIR_ENV_VAR+"="+imageDir)

			err = RunTelepresenceTests(incell, cmd, exePath, imageDir, instanceName, debug)
			if err != nil{
				StopTelepresence(telepresenceYamlPath)
				return fmt.Errorf("error occured while running tests, %s", err)
			}
		}
	}
	StopTelepresence(telepresenceYamlPath)

	return nil
}
func StopTelepresence(filepath string) error {
	err := kubernetes.DeleteFile(filepath)
	if err != nil {
		return fmt.Errorf("error occurred while stopping telepresence %s", err)
	}
	return nil
}

func RunTelepresenceTests(incell bool, cmd *exec.Cmd, exePath string, imageDir string, instanceName string, debug bool) error {
	deploymentName, err := CreateTelepresenceDeployment(incell, imageDir, instanceName)
	if err != nil {
		return fmt.Errorf("error occurred while creating telepresence deployment", err)
	}

	telepresenceExecPath := filepath.Join(util.CelleryInstallationDir(), constants.TELEPRESENCE_EXEC_PATH, "/telepresence")
	var telArgs = []string{telepresenceExecPath, "--deployment", *deploymentName}
	if !debug {
		balArgs := []string{exePath, "test", "--all"}
		telArgs = append(telArgs, "--run")
		telArgs = append(telArgs, balArgs...)
	}

	cmd.Path = telepresenceExecPath
	cmd.Args = telArgs
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error occurred while running tests", err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error occurred while waiting for tests to complete", err)
	}
	return nil
}

func RunDockerTelepresenceTests(incell bool, cmd *exec.Cmd, testsRoot string, imageDir string, instanceName string) error {

	deploymentName, err := CreateTelepresenceDeployment(incell, imageDir, instanceName)
	if err != nil {
		return fmt.Errorf("error occurred while creating telepresence deployment", err)
	}

	telepresenceExecPath := filepath.Join(util.CelleryInstallationDir(), constants.TELEPRESENCE_EXEC_PATH, "/telepresence")
	var telArgs = []string{telepresenceExecPath, "--deployment", *deploymentName}
	dockerCmdArgs := ConstructDockerArgs(testsRoot, imageDir)
	telArgs = append(telArgs, "--docker-run")
	telArgs = append(telArgs, dockerCmdArgs...)

	bashArgs := []string{"/bin/bash", "-c", strings.Join(telArgs, " ")}

	cmd.Env = os.Environ()
	imageDir = strings.Replace(imageDir, util.UserHomeDir(), "/home/cellery", 1)
	cmd.Env = append(cmd.Env, celleryImageDirEnvVar+"="+imageDir)
	cmd.Path = "/bin/bash"
	cmd.Args = bashArgs
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error occurred while running tests", err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error occurred while waiting for tests to complete", err)
	}
	return nil
}

func PromtConfirmation(assumeYes bool, balProj string) error {
	if !assumeYes {
		balConf := filepath.Join(balProj, constants.BALLERINA_CONF)
		var debugMsg string
		isExists, err := util.FileExists(balConf)
		if err != nil {
			return fmt.Errorf("error occured while checking if %s exists,", balConf, err )
		}
		if isExists {
			debugMsg = constants.BALLERINA_CONF + " already exists in project location: "+ balProj + ". This will be overridden to debug tests. "
		} else {
			debugMsg = constants.BALLERINA_CONF + " file will be created in your Ballerina project to debug tests."
		}
		fmt.Printf("%s " + util.Bold(debugMsg+ "Do you wish to continue with debugging the tests (Y/n)? "), util.YellowBold("?"))
		reader := bufio.NewReader(os.Stdin)
		confirmation, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.ToLower(strings.TrimSpace(confirmation)) == "n" {
			return fmt.Errorf("Cell testing aborted")
		}
	}
	return nil
}

func CreateBallerinaConf(iName string, verboseMode string, imageDir string, dependencyLinks string,
	envVars []*environmentVariable, balProj string) error {

	content := []string{fmt.Sprintf("DEBUG_MODE=\"%s\"\n", verboseMode)}
	content = append(content, fmt.Sprintf(constants.CelleryImageDirEnvVar+"=\"%s\"\n", imageDir))
	for _, envVar := range envVars {
		content = append(content, fmt.Sprintf(envVar.Key+"=\"%s\"\n", envVar.Value))
	}
	content = append(content, fmt.Sprintf("IMAGE_NAME=\"%s\"\n", strings.Replace(iName, "\"", "\\\"", -1)))
	content = append(content, fmt.Sprintf("DEPENDENCY_LINKS=\"%s\"\n",
		strings.Replace(dependencyLinks, "\"", "\\\"", -1)))

	ballerinaConfPath := filepath.Join(balProj, constants.BALLERINA_CONF)

	isExistsBalConf, err := util.FileExists(ballerinaConfPath)
	if err != nil {
		return err
	}
	if isExistsBalConf {
		err := os.Remove(ballerinaConfPath)
		if err != nil {
			return err
		}
	}
	_, err = os.Create(ballerinaConfPath)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(ballerinaConfPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	for _, element := range content {
		if _, err = f.WriteString(element); err != nil {
			return err
		}
	}
	return nil
}

func ConstructDockerArgs(testsRoot string, imageDir string) []string {
	cliUser, err := user.Current()
	if err != nil {
		util.ExitWithErrorMessage("Error while retrieving the current user", err)
	}

	dockerCmdArgs := []string{"-u", cliUser.Uid,
		"-l", "ballerina-runtime=" + version.BuildVersion(),
		"-e", fmt.Sprintf("CELLERY_IMAGE_DIR=%s", strings.Replace(imageDir, util.UserHomeDir(), "/home/cellery", 1)),
		"--mount", "type=bind,source=" + util.UserHomeDir() + string(os.PathSeparator) + ".ballerina,target=/home/cellery/.ballerina",
		"--mount", "type=bind,source=" + util.UserHomeDir() + string(os.PathSeparator) + ".cellery,target=/home/cellery/.cellery",
		"--mount", "type=bind,source=" + util.UserHomeDir() + string(os.PathSeparator) + ".kube,target=/home/cellery/.kube",
		"--mount", "type=bind,source=" + testsRoot + ",target=/home/cellery/tmp",
		"-w", "/home/cellery/",
		"wso2cellery/ballerina-runtime:" + version.BuildVersion(),
	}

	dockerCommand := []string{"./test.sh", cliUser.Uid}
	dockerCmdArgs = append(dockerCmdArgs, dockerCommand...)
	return dockerCmdArgs
}

func CreateTelepresenceDeployment(incell bool, imageDir string, instanceName string) (*string, error) {
	var srcYamlFile string
	dstYamlFile := filepath.Join(imageDir, "telepresence.yaml")
	var deploymentName string
	var spinner *util.Spinner
	var spinnerMsg string
	if incell {
		srcYamlFile = filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS, constants.TELEPRESENCE, "telepresence-deployment.yaml")
		err := util.CopyFile(srcYamlFile, dstYamlFile)
		if err != nil {
			return nil, fmt.Errorf("error while copying telepresene k8s artifact to %s", imageDir, err)
		}
		util.ReplaceInFile(dstYamlFile, "{{cell}}", instanceName, -1)
		deploymentName = instanceName + "--telepresence"
		spinnerMsg = "Creating telepresence deployment"
	} else {
		srcYamlFile = filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS, constants.TELEPRESENCE, "telepresence-cell.yaml")
		err := util.CopyFile(srcYamlFile, dstYamlFile)
		if err != nil {
			return nil, fmt.Errorf("error while copying telepresene k8s artifact to %s", imageDir, err)
		}
		deploymentName = "telepresence--telepresence-deployment"
		spinnerMsg = "Creating telepresence instance"
	}
	kubectl.ApplyFile(dstYamlFile)
	spinner = util.StartNewSpinner(spinnerMsg)
	time.Sleep(5 * time.Second)
	err := kubectl.WaitForDeployment("available", 900, deploymentName, "default")
	if err != nil {
		return nil, fmt.Errorf("error waiting for telepresence deployment %v to be available", deploymentName, err)
	}

	if !incell {
		err = kubectl.WaitForCell("Ready", 30*60, "telepresence", "default")
		if err != nil {
			return nil, fmt.Errorf("error waiting for instance telepresence to be ready", err)
		}
	}
	spinner.Stop(true)
	return &deploymentName, nil
}

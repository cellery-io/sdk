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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/ballerina"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

// RunTest starts Cell instance (along with dependency instances if specified by the user)\
func RunTest(cli cli.Cli, cellImageTag string, instanceName string, startDependencies bool, shareDependencies bool,
	dependencyLinks []string, envVars []string, assumeYes bool, debug bool, verbose bool, incell bool, projLocation string) error {
	var err error
	var imageDir string
	var parsedCellImage *image.CellImage
	if parsedCellImage, err = image.ParseImageTag(cellImageTag); err != nil {
		return fmt.Errorf("error occurred while parsing cell image, %v", err)
	}
	if err = cli.ExecuteTask("Extracting cell image", "Failed to extract cell image",
		"", func() error {
			imageDir, err = ExtractImage(cli, parsedCellImage, true)
			return err
		}); err != nil {
		return err
	}
	defer func() error {
		err = os.RemoveAll(imageDir)
		if err != nil {
			return fmt.Errorf("error occurred while cleaning up, %v", err)
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
			cellInstance, err := kubernetes.GetCell(dependencyLink.DependencyInstance)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("error occurred while validating dependency links, %v", err)
			} else {
				dependencyLink.IsRunning = err == nil && cellInstance.CellStatus.Status == "Ready"
				parsedDependencyLinks = append(parsedDependencyLinks, dependencyLink)
				continue
			}
			compositeInstance, err := kubernetes.GetComposite(dependencyLink.DependencyInstance)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("error occurred while validating dependency links, %v", err)
			} else {
				dependencyLink.IsRunning = err == nil && compositeInstance.CompositeStatus.Status == "Ready"
				parsedDependencyLinks = append(parsedDependencyLinks, dependencyLink)
			}
		}
	}

	instanceEnvVars := []*environmentVariable{}
	if len(envVars) > 0 {
		// Parsing environment variables
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

	err = startTestCellInstance(cli, imageDir, instanceName, mainNode, instanceEnvVars, startDependencies,
		shareDependencies, rootNodeDependencies, verbose, debug, incell, assumeYes, projLocation)
	if err != nil {
		return fmt.Errorf("failed to test Cell instance "+instanceName+", %v", err)
	}
	//Remove Ballerina toml and local repo to avoid failing cellery build and run
	if !debug {
		os.Remove(constants.BallerinaToml)
		os.Remove(constants.BallerinaConf)
	}
	util.PrintSuccessMessage(fmt.Sprintf("Completed running tests for instance %s", util.Bold(instanceName)))
	return nil
}

func startTestCellInstance(cli cli.Cli, imageDir string, instanceName string, runningNode *dependencyTreeNode,
	envVars []*environmentVariable, startDependencies bool, shareDependencies bool, dependencyLinks map[string]*dependencyInfo,
	verbose bool, debug bool, incell bool, assumeYes bool, projLocation string) error {
	imageTag := fmt.Sprintf("%s/%s:%s", runningNode.MetaData.Organization, runningNode.MetaData.Name,
		runningNode.MetaData.Version)
	balFileName, err := util.GetSourceFileName(filepath.Join(imageDir, constants.ZipBallerinaSource))
	if err != nil {
		return fmt.Errorf("failed to find source file in Image %s due to %v", imageTag, err)
	}
	balFilePath := filepath.Join(imageDir, constants.ZipBallerinaSource, balFileName)

	currentDir := cli.FileSystem().CurrentDir()
	if err != nil {
		return fmt.Errorf("error in determining working directory, %v", err)
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
		return fmt.Errorf("error in generating cellery:CellImageName construct, %v", err)
	}

	// Get Ballerina Installation if exists
	exePath, err := cli.BalExecutor().ExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path, %v", err)
	}
	verboseMode := strconv.FormatBool(verbose)

	telepresenceYamlPath := filepath.Join(imageDir, "telepresence.yaml")
	if debug && exePath == "" {
		return fmt.Errorf("Ballerina should be installed to debug tests.")
	}
	balProjectName := constants.TargetDirName

	balTomlPath := filepath.Join(imageDir, constants.ZipBallerinaSource, constants.BallerinaToml)
	testsPath := filepath.Join(imageDir, constants.ZipTests)
	testsRoot := filepath.Join(currentDir, balProjectName)

	var balEnvVars []*ballerina.EnvironmentVariable
	// Set celleryImageDirEnvVar environment variable.

	balEnvVars = append(balEnvVars, &ballerina.EnvironmentVariable{
		Key:   celleryImageDirEnvVar,
		Value: imageDir})
	for _, envVar := range envVars {
		// Export environment variables defined by user for root instance
		if envVar.InstanceName == "" || envVar.InstanceName == instanceName {
			balEnvVars = append(balEnvVars, &ballerina.EnvironmentVariable{
				Key:   envVar.Key,
				Value: envVar.Value,
			})
		}
		// Export environment variables defined by user for dependent instances
		if !(envVar.InstanceName == "" || envVar.InstanceName == instanceName) {
			balEnvVars = append(balEnvVars, &ballerina.EnvironmentVariable{
				Key:   celleryEnvVarPrefix + envVar.InstanceName + "." + envVar.Key,
				Value: envVar.Value,
			})
		}
	}

	var testCmdArgs []string
	if testCmdArgs, err = testCommandArgs(cli, incell, imageDir, instanceName); err != nil {
		return fmt.Errorf("failed to get run command arguements, %v", err)
	}

	// Construct test artifact names
	if debug {
		balProjectName = filepath.Base(projLocation)
		testsRoot, err = filepath.Abs(projLocation)
		if err != nil {
			return err
		}
		PromtConfirmation(assumeYes, testsRoot)
		balEnvVars = append(balEnvVars, &ballerina.EnvironmentVariable{
			Key:   "DEBUG",
			Value: "true",
		})
	} else {
		var balModule string

		if instanceName != "" {
			balModule = filepath.Join(testsRoot, constants.ZipBallerinaSource, instanceName)
		} else {
			balModule = filepath.Join(testsRoot, constants.ZipBallerinaSource, constants.TempTestModule)
		}
		isTestDirExists, _ := util.FileExists(testsPath)
		var isBallerinaProject bool

		read, err := ioutil.ReadFile(balFilePath)
		containsTests := strings.Contains(string(read), "@test")
		if err != nil {
			return err
		}
		if !isTestDirExists && !containsTests {
			return fmt.Errorf("no tests found in the cell image %v", imageTag)
		}
		err = util.RemoveDir(testsRoot)
		if err != nil {
			return fmt.Errorf("error occurred while deleting filepath, %s", err)
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
				return fmt.Errorf("error occurred while creating %s, %v", testsRoot, err)
			}
		}

		// Copy artifacts for running tests
		if isBallerinaProject {
			fileCopyError := util.CopyFile(balTomlPath, filepath.Join(testsRoot, constants.BallerinaToml))
			if fileCopyError != nil {
				return fmt.Errorf("error occurred while copying %s, %v", constants.BallerinaToml, fileCopyError)
			}
		}
		fileCopyError := util.CopyDir(testsPath, filepath.Join(balModule, constants.ZipTests))
		if fileCopyError != nil {
			return fmt.Errorf("error occurred while copying tests folder, %v", fileCopyError)
		}

		err = util.CleanAndCreateDir(filepath.Join(testsRoot, "logs"))
		if err != nil {
			return fmt.Errorf("error occurred while creating %s, %v", filepath.Join(testsRoot, "logs"), err)
		}

		// Change working dir to Bal project
		err = os.Chdir(testsRoot)
		if err != nil {
			return fmt.Errorf("error occurred while changing working directory, %v", err)
		}

		isExistsSourceBal, err := util.FileExists(filepath.Join(balModule, filepath.Base(balFilePath)))
		if !isExistsSourceBal {
			fileCopyError := util.CopyFile(balFilePath, filepath.Join(balModule, filepath.Base(balFilePath)))
			if fileCopyError != nil {
				return fmt.Errorf("error occurred while copying source bal file, %v", fileCopyError)
			}
		}
	}

	err = CreateBallerinaConf(string(iName), verboseMode, imageDir, string(dependencyLinksJson), envVars, testsRoot)
	if err != nil {
		return fmt.Errorf("error occurred while creating the ballerina.conf, %v", err)
	}

	if err = cli.BalExecutor().Test(filepath.Join(testsRoot, constants.BallerinaConf), testCmdArgs, balEnvVars); err != nil {
		StopTelepresence(telepresenceYamlPath)
		return fmt.Errorf("failed to run bal file, %v", err)
	}
	if err != nil {
		StopTelepresence(telepresenceYamlPath)
		return fmt.Errorf("error occured while running tests in ballerina docker image, %v", err)
	}
	StopTelepresence(telepresenceYamlPath)
	return nil
}
func StopTelepresence(filepath string) error {
	err := kubernetes.DeleteFile(filepath)
	if err != nil {
		return fmt.Errorf("error occurred while stopping telepresence %v", err)
	}
	return nil
}

func PromtConfirmation(assumeYes bool, balProj string) error {
	if !assumeYes {
		balConf := filepath.Join(balProj, constants.BallerinaConf)
		var debugMsg string
		isExists, err := util.FileExists(balConf)
		if err != nil {
			return fmt.Errorf("error occured while checking if %v exists, %v", balConf, err)
		}
		if isExists {
			debugMsg = constants.BallerinaConf + " already exists in project location: " + balProj + ". This will be overridden to debug tests. "
		} else {
			debugMsg = constants.BallerinaConf + " file will be created in project location: " + balProj + " to debug tests."
		}
		fmt.Printf("%s "+util.Bold(debugMsg+"Do you wish to continue with debugging the tests (Y/n)? "), util.YellowBold("?"))
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

	ballerinaConfPath := filepath.Join(balProj, constants.BallerinaConf)

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

	f, err := os.OpenFile(ballerinaConfPath, os.O_CREATE|os.O_WRONLY, 0600)
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

func ConstructDockerArgs(testsRoot string, imageDir string) ([]string, error) {
	cliUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("error while retrieving the current user, %v", err)
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

	dockerCommand := []string{"./" + constants.BalTestExecFIle, cliUser.Uid}
	dockerCmdArgs = append(dockerCmdArgs, dockerCommand...)
	return dockerCmdArgs, nil
}

func CreateTelepresenceDeployment(incell bool, imageDir string, instanceName string) (*string, error) {
	var srcYamlFile string
	dstYamlFile := filepath.Join(imageDir, "telepresence.yaml")
	var deploymentName string

	if incell {
		srcYamlFile = filepath.Join(util.CelleryInstallationDir(), constants.K8sArtifacts, constants.TELEPRESENCE, "telepresence-deployment.yaml")
		err := util.CopyFile(srcYamlFile, dstYamlFile)
		if err != nil {
			return nil, fmt.Errorf("error while copying telepresene k8s artifact to %s, %v", imageDir, err)
		}
		util.ReplaceInFile(dstYamlFile, "{{cell}}", instanceName, -1)
		deploymentName = instanceName + "--telepresence"
	} else {
		srcYamlFile = filepath.Join(util.CelleryInstallationDir(), constants.K8sArtifacts, constants.TELEPRESENCE, "telepresence-cell.yaml")
		err := util.CopyFile(srcYamlFile, dstYamlFile)
		if err != nil {
			return nil, fmt.Errorf("error while copying telepresene k8s artifact to %s, %v", imageDir, err)
		}
		deploymentName = "telepresence--telepresence-deployment"
	}
	kubernetes.ApplyFile(dstYamlFile)
	time.Sleep(5 * time.Second)
	err := kubernetes.WaitForDeployment("available", 900, deploymentName, "default")
	if err != nil {
		return nil, fmt.Errorf("error waiting for telepresence deployment %s to be available, %v", deploymentName, err)
	}

	if !incell {
		err = kubernetes.WaitForCell("Ready", 30*60, "telepresence", "default")
		if err != nil {
			return nil, fmt.Errorf("error waiting for instance telepresence to be ready, %v", err)
		}
	}
	return &deploymentName, nil
}

func testCommandArgs(cli cli.Cli, incell bool, imageDir string, instanceName string) ([]string, error) {
	var err error
	var deploymentName *string
	if err = cli.ExecuteTask("Creating telepresence k8s resources",
		"Failed to create telepresence k8s resources",
		"", func() error {
			deploymentName, err = CreateTelepresenceDeployment(incell, imageDir, instanceName)
			return err
		}); err != nil {
		return nil, err
	}

	telepresenceExecPath := filepath.Join(util.CelleryInstallationDir(), constants.TelepresenceExecPath, "/telepresence")
	var telArgs = []string{telepresenceExecPath, "--deployment", *deploymentName}
	return telArgs, nil
}

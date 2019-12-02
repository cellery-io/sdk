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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/ballerina"
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/image"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/util"
)

const CelleryTestVerboseMode = "CELLERY_DEBUG_MODE"

// RunTest starts Cell instance (along with dependency instances if specified by the user)\
func RunTest(cli cli.Cli, cellImageTag string, instanceName string, startDependencies bool, shareDependencies bool,
	dependencyLinks []string, envVars []string, assumeYes bool, debug bool, verbose bool, disableTelepresence bool, incell bool, projLocation string) error {
	var err error
	var imageDir string
	var parsedCellImage *image.CellImage
	if parsedCellImage, err = image.ParseImageTag(cellImageTag); err != nil {
		return err
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
			return err
		}
		return nil
	}()

	// Reading Cell Image metadata
	var metadataFileContent []byte
	if metadataFileContent, err = ioutil.ReadFile(filepath.Join(imageDir, artifacts, "cellery",
		"metadata.json")); err != nil {
		return err
	}
	cellImageMetadata := &image.MetaData{}
	if err = json.Unmarshal(metadataFileContent, cellImageMetadata); err != nil {
		return err
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
		shareDependencies, rootNodeDependencies, verbose, debug, disableTelepresence, incell, assumeYes, projLocation)
	if err != nil {
		return fmt.Errorf("failed to test Cell instance "+instanceName+", %v", err)
	}

	//Cleanup telepresence deployment started for tests
	err = kubernetes.DeleteFile(filepath.Join(imageDir, "telepresence.yaml"))
	if err != nil {
		return err
	}
	util.PrintSuccessMessage(fmt.Sprintf("Completed running tests for instance %s", util.Bold(instanceName)))
	return nil
}

func startTestCellInstance(cli cli.Cli, imageDir string, instanceName string, runningNode *dependencyTreeNode,
	envVars []*environmentVariable, startDependencies bool, shareDependencies bool, dependencyLinks map[string]*dependencyInfo,
	verbose bool, debug bool, onlyDocker bool, incell bool, assumeYes bool, projLocation string) error {
	imageTag := fmt.Sprintf("%s/%s:%s", runningNode.MetaData.Organization, runningNode.MetaData.Name,
		runningNode.MetaData.Version)
	balFileName, err := util.GetSourceName(filepath.Join(imageDir, constants.ZipBallerinaSource))
	if err != nil {
		return err
	}
	cellBalFilePath := filepath.Join(imageDir, constants.ZipBallerinaSource, balFileName)

	currentDir := cli.FileSystem().CurrentDir()
	if err != nil {
		return err
	}

	// Preparing the dependency instance map
	dependencyLinksJson, err := json.Marshal(dependencyLinks)
	if err != nil {
		return err
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
		return err
	}

	// Get Ballerina Installation if exists
	exePath, err := cli.BalExecutor().ExecutablePath()
	if err != nil {
		return err
	}
	verboseMode := strconv.FormatBool(verbose)
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
	balEnvVars = append(balEnvVars, &ballerina.EnvironmentVariable{
		Key:   CelleryTestVerboseMode,
		Value: verboseMode})

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

	// If --debug flag is passed, start a Telepresence shell
	// Else create a ballerina project, copy files from image and execute ballerina test command via Telepresence
	if debug {
		balProjectName = filepath.Base(projLocation)
		testsRoot, err = filepath.Abs(projLocation)
		if err != nil {
			return err
		}
		if !assumeYes {
			isConfirmed, err := PromtConfirmation(testsRoot, debug)
			if err != nil {
				return err
			}
			if !isConfirmed {
				return fmt.Errorf("cell testing aborted")
			}
		}
		balEnvVars = append(balEnvVars, &ballerina.EnvironmentVariable{
			Key:   "DEBUG",
			Value: "true",
		})
	} else {
		var isBallerinaProject bool
		var balModulePath string
		var isTestsContainBeforeSuite bool
		var isTestsContainAfterSuite bool

		//Check if tests exist in the cell image
		// Return an error if tests not found
		isTestDirExists, err := util.FileExists(testsPath)
		if err != nil {
			return err
		}
		isSourceBalContainsTests, err := util.FindPatternInFile("@test:Config.*", cellBalFilePath)
		if err != nil {
			return err
		}
		if !isTestDirExists && !isSourceBalContainsTests {
			return fmt.Errorf("no tests found in the cell image %v", imageTag)
		}

		// Check if @test:BeforeSuite and @test:AfterSuite functions are defined in cell bal file or test bal files
		// Print a warning if at least one of the functions are missing and prompt for confirmation to continue
		if isSourceBalContainsTests {
			isTestsContainBeforeSuite, err = util.FindPatternInFile("@test:BeforeSuite", cellBalFilePath)
			if err != nil {
				return err
			}
			isTestsContainAfterSuite, err = util.FindPatternInFile("@test:AfterSuite", cellBalFilePath)
			if err != nil {
				return err
			}
		}

		testFiles, err := ioutil.ReadDir(testsPath)
		if err != nil {
			return err
		}
		if !isTestsContainBeforeSuite {
			for _, f := range testFiles {
				isTestsContainBeforeSuite, err = util.FindPatternInFile(
					"@test:BeforeSuite", filepath.Join(testsPath, f.Name()))
				if err != nil {
					return err
				}
				if isTestsContainBeforeSuite {
					break
				}
			}
		}
		if !isTestsContainAfterSuite {
			for _, f := range testFiles {
				isTestsContainAfterSuite, err = util.FindPatternInFile(
					"@test:AfterSuite", filepath.Join(testsPath, f.Name()))
				if err != nil {
					return err
				}
				if isTestsContainAfterSuite {
					break
				}
			}
		}

		// Print a warning message if atleast one of the functions are missing
		// Prompt for confirmation to continue
		if !isTestsContainBeforeSuite || !isTestsContainAfterSuite {
			var missingFunctions []string
			if !isTestsContainBeforeSuite {
				missingFunctions = append(missingFunctions, "@BeforeSuite")
			}
			if !isTestsContainAfterSuite {
				missingFunctions = append(missingFunctions, "@AfterSuite")
			}

			util.PrintWarningMessage(fmt.Sprintf(
				"Missing function(s) %s in test files. "+
					"Please make sure the instances are already available in the cluster or abort to write the missing functions.",
				util.CyanBold(strings.Join(missingFunctions, ", "))))

			if !assumeYes {
				// passing an empty string for bal project path
				// since it is used only to print the message for the debug mode
				isConfirmed, err := PromtConfirmation("", debug)
				if err != nil {
					return err
				}
				if !isConfirmed {
					return fmt.Errorf("cell testing aborted")
				}
			}
			err = util.RemoveDir(testsRoot)
			if err != nil {
				return err
			}
		}

		// Create a Ballerina project named target
		isBallerinaProject, err = util.FileExists(balTomlPath)
		if err != nil {
			return err
		}
		err = util.RemoveDir(testsRoot)
		if err != nil {
			return err
		}
		if exePath != "" {
			cmd := exec.Command(exePath, "new", balProjectName)
			if verbose {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}
			err = cmd.Run()
			if err != nil {
				return err
			}
		}

		// Copy Ballerina.toml from image if exists for running tests
		if isBallerinaProject {
			fileCopyError := util.CopyFile(balTomlPath, filepath.Join(testsRoot, constants.BallerinaToml))
			if fileCopyError != nil {
				return fileCopyError
			}
		}

		// Set ballerina module name and copy source file and tests
		if instanceName != "" && instanceName != "test" {
			balModulePath = filepath.Join(testsRoot, constants.ZipBallerinaSource, instanceName)
		} else {
			balModulePath = filepath.Join(testsRoot, constants.ZipBallerinaSource, constants.TempTestModule)
		}
		err = util.CreateDir(balModulePath)
		if err != nil {
			return err
		}

		fileCopyError := util.CopyFile(cellBalFilePath, filepath.Join(balModulePath, filepath.Base(cellBalFilePath)))
		if fileCopyError != nil {
			return fileCopyError
		}

		fileCopyError = util.CopyDir(testsPath, filepath.Join(balModulePath, constants.ZipTests))
		if fileCopyError != nil {
			return fileCopyError
		}

		// Change working dir to Bal project to execute tests
		err = os.Chdir(testsRoot)
		if err != nil {
			return fmt.Errorf("error occurred while changing working directory, %v", err)
		}
	}

	//Create ballerina.conf to pass CLI flags to the Ballerina process
	err = CreateBallerinaConf(string(iName), string(dependencyLinksJson), startDependencies, shareDependencies,
		verboseMode, imageDir, envVars, testsRoot)
	if err != nil {
		return err
	}

	//Create logs directory for test execution logs
	err = util.CleanAndCreateDir(filepath.Join(testsRoot, "logs"))
	if err != nil {
		return err
	}

	var testCmdArgs []string
	// if --disable-telepresence flag is passed, pass an empty array to Bal executor since telepresence is not required
	if !onlyDocker {
		if testCmdArgs, err = testCommandArgs(cli, incell, imageDir, instanceName); err != nil {
			return err
		}
	}
	if err = cli.BalExecutor().Test(filepath.Join(testsRoot, constants.BallerinaConf), testCmdArgs, balEnvVars); err != nil {
		DeleteTelepresenceResouces(imageDir)
		return err
	}
	DeleteTelepresenceResouces(imageDir)
	return nil
}

func PromtConfirmation(balProj string, debug bool) (bool, error) {
	if !debug {
		fmt.Printf("%s "+util.Bold("Do you wish to continue running tests (Y/n)? "), util.YellowBold("?"))
	} else {
		debugMsg := "The following will be created/overridden in your project location %s:\n" +
			"  1) %s\n" +
			"  2) %s\n"
		fmt.Printf(util.CyanBold(fmt.Sprintf(debugMsg, balProj, constants.BallerinaConf, "logs/")))
		fmt.Printf("%s"+util.Bold(" Do you wish to continue debugging tests (Y/n)? "), util.YellowBold("?"))
	}

	reader := bufio.NewReader(os.Stdin)
	confirmation, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	if strings.ToLower(strings.TrimSpace(confirmation)) == "n" {
		return false, nil
	}
	return true, nil
}

func CreateBallerinaConf(iName string, dependencyLinks string, startDependencies bool, shareDependencies bool,
	verboseMode string, imageDir string, envVars []*environmentVariable, balProj string) error {

	content := []string{fmt.Sprintf(CelleryTestVerboseMode+"=\"%s\"\n", verboseMode)}
	content = append(content, fmt.Sprintf(constants.CelleryImageDirEnvVar+"=\"%s\"\n", imageDir))
	for _, envVar := range envVars {
		content = append(content, fmt.Sprintf(envVar.Key+"=\"%s\"\n", envVar.Value))
	}
	content = append(content, "[test.config]\n")
	content = append(content, fmt.Sprintf("IMAGE_NAME=\"%s\"\n", strings.Replace(iName, "\"", "\\\"", -1)))
	content = append(content, fmt.Sprintf("DEPENDENCY_LINKS=\"%s\"\n",
		strings.Replace(dependencyLinks, "\"", "\\\"", -1)))
	content = append(content, fmt.Sprintf("START_DEPENDENCIES=%s\n", strconv.FormatBool(startDependencies)))
	content = append(content, fmt.Sprintf("SHARE_DEPENDENCIES=%s\n", strconv.FormatBool(shareDependencies)))

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

func CreateTelepresenceResources(incell bool, imageDir string, instanceName string) (*string, error) {
	var srcYamlFile string
	dstYamlFile := filepath.Join(imageDir, "telepresence.yaml")
	var deploymentName string

	if incell {
		srcYamlFile = filepath.Join(util.CelleryInstallationDir(), constants.K8sArtifacts, constants.TELEPRESENCE, "telepresence-deployment.yaml")
		err := util.CopyFile(srcYamlFile, dstYamlFile)
		if err != nil {
			return nil, err
		}
		util.ReplaceInFile(dstYamlFile, "{{cell}}", instanceName, -1)
		deploymentName = instanceName + "--telepresence"
	} else {
		srcYamlFile = filepath.Join(util.CelleryInstallationDir(), constants.K8sArtifacts, constants.TELEPRESENCE, "telepresence-cell.yaml")
		err := util.CopyFile(srcYamlFile, dstYamlFile)
		if err != nil {
			return nil, err
		}
		deploymentName = "telepresence--telepresence-deployment"
	}
	kubernetes.ApplyFile(dstYamlFile)
	time.Sleep(5 * time.Second)
	err := kubernetes.WaitForDeployment("available", 900, deploymentName, "default")
	if err != nil {
		return nil, err
	}

	if !incell {
		err = kubernetes.WaitForCell("Ready", 30*60, "telepresence", "default")
		if err != nil {
			return nil, err
		}
	}
	return &deploymentName, nil
}

func DeleteTelepresenceResouces(imageDir string) {
	err := kubernetes.DeleteFile(filepath.Join(imageDir, "telepresence.yaml"))
	if err != nil {
		util.PrintWarningMessage(fmt.Sprintf("Failed to delete telepresence resources in the Kubernetes cluster, %v ", err))
	}
}

func testCommandArgs(cli cli.Cli, incell bool, imageDir string, instanceName string) ([]string, error) {
	var err error
	var deploymentName *string
	if err = cli.ExecuteTask("Creating telepresence k8s resources",
		"Failed to create telepresence k8s resources",
		"", func() error {
			deploymentName, err = CreateTelepresenceResources(incell, imageDir, instanceName)
			return err
		}); err != nil {
		return nil, err
	}
	var telArgs = []string{"--deployment", *deploymentName}
	return telArgs, nil
}

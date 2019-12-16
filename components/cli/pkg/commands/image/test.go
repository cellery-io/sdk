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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/ballerina"
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/util"
)

const CelleryTestVerboseMode = "CELLERY_DEBUG_MODE"
const Logs = "logs"

// RunTest starts Cell instance (along with dependency instances if specified by the user)\
func RunTest(cli cli.Cli, cellImageTag string, instanceName string, startDependencies bool, shareDependencies bool,
	dependencyLinks []string, envVars []string, assumeYes bool, debug bool, verbose bool, disableTelepresence bool, incell bool, projLocation string) error {
	extractedImage, err := extractImage(cli, cellImageTag, instanceName, dependencyLinks, envVars)
	err = startTestCellInstance(cli, extractedImage, instanceName, startDependencies,
		shareDependencies, verbose, debug, disableTelepresence, incell, assumeYes, projLocation)
	//Cleanup telepresence deployment started for tests
	if !disableTelepresence {
		DeleteTelepresenceResouces(extractedImage.ImageDir)
	}
	if err != nil {
		return fmt.Errorf("failed to test Cell instance "+instanceName+", %v", err)
	}

	util.PrintSuccessMessage(fmt.Sprintf("Successfully tested image %s", util.Bold(cellImageTag)))
	return nil
}

func startTestCellInstance(cli cli.Cli, extractedImage *ExtractedImage, instanceName string, startDependencies bool, shareDependencies bool,
	verbose bool, debug bool, disableTelepresence bool, incell bool, assumeYes bool, projLocation string) error {
	currentDir := cli.FileSystem().CurrentDir()
	imageDir := extractedImage.ImageDir
	runningNode := extractedImage.MainNode
	envVars := extractedImage.InstanceEnvVars
	imageTag := fmt.Sprintf("%s/%s:%s", runningNode.MetaData.Organization, runningNode.MetaData.Name,
		runningNode.MetaData.Version)
	verboseMode := strconv.FormatBool(verbose)
	var projectDir, balModulePath string

	var balEnvVars []*ballerina.EnvironmentVariable
	// Set celleryImageDirEnvVar environment variable.
	balEnvVars = append(balEnvVars, &ballerina.EnvironmentVariable{
		Key:   celleryImageDirEnvVar,
		Value: imageDir})
	// Set verbose
	// mode to print debug logs
	balEnvVars = append(balEnvVars, &ballerina.EnvironmentVariable{
		Key:   CelleryTestVerboseMode,
		Value: verboseMode})
	// Setting user defined environment variables.
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

	// Get Ballerina Installation if exists
	exePath, err := cli.BalExecutor().ExecutablePath()
	if err != nil {
		return err
	}
	if debug && exePath == "" {
		return fmt.Errorf("Ballerina should be installed to debug tests.")
	}

	// If --debug flag is passed, start a Telepresence shell
	// Else create a ballerina project, copy files from image and execute ballerina test command via Telepresence
	if debug {
		projectDir, err = filepath.Abs(projLocation)
		if err != nil {
			return err
		}
		if !assumeYes {
			isConfirmed, err := PromtConfirmation(projectDir, debug)
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
		testDir, err := util.FindRecursiveInDirectory(extractedImage.ImageDir, "tests")
		if err != nil {
			return err
		}
		if testDir == nil {
			return fmt.Errorf("no tests found in Image %s", imageTag)
		}
		balSourceName, err := util.GetSourceName(filepath.Join(imageDir, src))
		if err != nil {
			return fmt.Errorf("failed to find source file in Image %s due to %v", imageTag, err)
		}
		projectDir = filepath.Join(imageDir, src, balSourceName)
		var modules []os.FileInfo
		if modules, err = ioutil.ReadDir(filepath.Join(projectDir, src)); err != nil {
			return err
		}
		balModulePath = filepath.Join(projectDir, src, modules[0].Name())

		// Check if @test:BeforeSuite and @test:AfterSuite functions are defined in any bal file
		// Print a warning if at least one of the functions are missing and prompt for confirmation to continue
		var isTestsContainBeforeSuite bool
		var isTestsContainAfterSuite bool
		balFilesList, err := util.FindRecursiveInDirectory(filepath.Join(balModulePath, "tests"), "*.bal")
		if err != nil {
			return err
		}
		for _, balFile := range balFilesList {
			isTestsContainBeforeSuite, err = util.FindPatternInFile("@test:BeforeSuite", balFile)
			if err != nil {
				return err
			}
			if isTestsContainBeforeSuite {
				break
			}
		}
		for _, balFile := range balFilesList {
			isTestsContainAfterSuite, err = util.FindPatternInFile("@test:AfterSuite", balFile)
			if err != nil {
				return err
			}
			if isTestsContainAfterSuite {
				break
			}
		}

		// Print a warning message if atleast one of the functions are missing
		// Prompt for confirmation to continue if -y flag is not passed
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
				// passing an empty string for bal project path since it is used only for the message for debugging
				isConfirmed, err := PromtConfirmation("", debug)
				if err != nil {
					return err
				}
				if !isConfirmed {
					return fmt.Errorf("cell testing aborted")
				}
			}
			if err != nil {
				return err
			}
		}

	}

	//Create ballerina.conf to pass info given in CLI flags to the Ballerina process
	err = CreateBallerinaConf(instanceName, extractedImage, startDependencies, shareDependencies, verboseMode, projectDir)
	if err != nil {
		return err
	}

	//Create logs directory for test execution logs
	err = util.CleanAndCreateDir(filepath.Join(projectDir, "logs"))
	if err != nil {
		return err
	}
	target := filepath.Join(currentDir, constants.TargetDirName)
	if err = util.CleanOrCreateDir(target); err != nil {
		return err
	}

	var testCmdArgs []string
	// if --disable-telepresence flag is passed, pass an empty array to Bal executor since telepresence is not required
	if !disableTelepresence {
		if testCmdArgs, err = testCommandArgs(cli, incell, imageDir, instanceName); err != nil {
			return err
		}
	}

	if err = cli.BalExecutor().Test(testCmdArgs, balEnvVars, projectDir); err != nil {
		if err = util.CopyDir(filepath.Join(projectDir, Logs), filepath.Join(target, Logs)); err != nil {
			return err
		}
		return err
	}
	if err = util.CopyDir(filepath.Join(projectDir, Logs), filepath.Join(target, Logs)); err != nil {
		return err
	}
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

func CreateBallerinaConf(instanceName string, extractedImage *ExtractedImage, startDependencies bool, shareDependencies bool,
	verboseMode string, balProj string) error {
	imageDir := extractedImage.ImageDir
	envVars := extractedImage.InstanceEnvVars
	var imageNameStruct = &dependencyInfo{
		Organization: extractedImage.MainNode.MetaData.Organization,
		Name:         extractedImage.MainNode.MetaData.Name,
		Version:      extractedImage.MainNode.MetaData.Version,
		InstanceName: instanceName,
		IsRoot:       true,
	}
	iName, err := json.Marshal(imageNameStruct)
	if err != nil {
		return err
	}
	iNameStr := string(iName)

	dependencyLinks := extractedImage.RootNodeDependencies
	dependencyLinksJson, err := json.Marshal(dependencyLinks)
	if err != nil {
		return err
	}
	dependencyLinksStr := string(dependencyLinksJson)

	content := []string{fmt.Sprintf(CelleryTestVerboseMode+"=\"%s\"\n", verboseMode)}
	content = append(content, fmt.Sprintf(constants.CelleryImageDirEnvVar+"=\"%s\"\n", imageDir))
	for _, envVar := range envVars {
		content = append(content, fmt.Sprintf(envVar.Key+"=\"%s\"\n", envVar.Value))
	}
	content = append(content, "[test.config]\n")
	content = append(content, fmt.Sprintf("IMAGE_NAME=\"%s\"\n", strings.Replace(iNameStr, "\"", "\\\"", -1)))
	content = append(content, fmt.Sprintf("DEPENDENCY_LINKS=\"%s\"\n",
		strings.Replace(dependencyLinksStr, "\"", "\\\"", -1)))
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
	err := kubernetes.WaitForDeployment("available", 900, deploymentName)
	if err != nil {
		return nil, err
	}

	if !incell {
		err = kubernetes.WaitForCell("Ready", 30*60, "telepresence")
		if err != nil {
			return nil, err
		}
	}
	return &deploymentName, nil
}

func DeleteTelepresenceResouces(imageDir string) {
	out, err := kubernetes.DeleteResource("cells.mesh.cellery.io", "telepresence")
	if err != nil {
		util.PrintWarningMessage(
			fmt.Sprintf("Failed to delete telepresence resources in the Kubernetes cluster, %v ", fmt.Errorf(out)))
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

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

	"github.com/olekukonko/tablewriter"

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

// RunRun starts Cell instance (along with dependency instances if specified by the user)
// This also support linking instances to parts of the dependency tree
// This command also strictly validates whether the requested Cell (and the dependencies are valid)
func RunTest(cellImageTag string, instanceName string, startDependencies bool, shareDependencies bool,
	dependencyLinks []string, envVars []string, assumeYes bool, debug bool) {
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
	spinner.SetNewAction("Reading Cell Image " + util.Bold(cellImageTag))
	metadataFileContent, err := ioutil.ReadFile(filepath.Join(imageDir, constants.ZIP_ARTIFACTS, "cellery",
		"metadata.json"))
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred while reading Cell Image metadata", err)
	}
	cellImageMetadata := &image.CellImageMetaData{}
	err = json.Unmarshal(metadataFileContent, cellImageMetadata)
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred while reading Cell Image metadata", err)
	}

	if instanceName == "" {
		// Setting a unique instance name since it is not provided
		instanceName, err = generateRandomInstanceName(cellImageMetadata)
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error occurred while preparing", err)
		}
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
			}
			dependencyLink.IsRunning = err == nil && cellInstance.CellStatus.Status == "Ready"
			parsedDependencyLinks = append(parsedDependencyLinks, dependencyLink)
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
	spinner.SetNewAction("Validating dependencies")
	immediateDependencies := map[string]*dependencyTreeNode{}
	// Check if the provided links are immediate dependencies of the root Cell
	for _, link := range parsedDependencyLinks {
		if metadata, hasKey := cellImageMetadata.Dependencies[link.DependencyAlias]; hasKey {
			immediateDependencies[link.DependencyAlias] = &dependencyTreeNode{
				Instance:  link.DependencyInstance,
				MetaData:  metadata,
				IsShared:  false,
				IsRunning: link.IsRunning,
			}
		} else {
			// If cellImageMetadata does not contain the provided link, there is a high chance that the user
			// made a mistake in the command. Therefore, this is validated strictly
			var allowedAliases []string
			for alias := range cellImageMetadata.Dependencies {
				allowedAliases = append(allowedAliases, alias)
			}
			spinner.Stop(false)
			util.ExitWithErrorMessage("Invalid links",
				fmt.Errorf("only aliases of the main Cell instance %s: [%s] are allowed when running "+
					"without starting dependencies, received %s", instanceName,
					strings.Join(allowedAliases, ", "), link.DependencyAlias))
		}
	}

	// Check if instances are provided for all the dependencies of the root Cell
	for alias := range cellImageMetadata.Dependencies {
		isLinkProvided := false
		for _, link := range parsedDependencyLinks {
			if link.DependencyAlias == alias {
				isLinkProvided = true
				break
			}
		}
		if !isLinkProvided {
			// If a link is not provided for a particular dependency, the main instance cannot be started.
			// The links is required for the main instance to discover the dependency in the runtime
			spinner.Stop(false)
			util.ExitWithErrorMessage("Links for all the dependencies not found",
				fmt.Errorf("required link for alias %s in instance %s not found", alias, instanceName))
		}

	}
	mainNode = &dependencyTreeNode{
		Instance:     instanceName,
		MetaData:     cellImageMetadata,
		IsRunning:    false,
		IsShared:     false,
		Dependencies: immediateDependencies,
	}
	err = validateDependencyTree(mainNode)
	if err != nil {
		util.ExitWithErrorMessage("Invalid instance linking", err)
	}
	spinner.SetNewAction("")
	err = confirmTestDependencyTree(mainNode, assumeYes)
	if err != nil {
		util.ExitWithErrorMessage("Failed to confirm the dependency tree", err)
	}

	util.PrintSuccessMessage("Starting execution of tests for " + util.Bold(cellImageTag) + "...")
	err = startTestCellInstance(imageDir, instanceName, mainNode, instanceEnvVars[instanceName], debug)
	if err != nil {
		util.ExitWithErrorMessage("Failed to start Cell instance "+instanceName, err)
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Completed running tests for instance %s", util.Bold(instanceName)))
}

// confirmDependencyTree confirms from the user whether the intended dependency tree had been resolved
func confirmTestDependencyTree(tree *dependencyTreeNode, assumeYes bool) error {
	var dependencyData [][]string
	var traversedInstances []string
	// Preparing instances table data
	var extractDependencyTreeData func(subTree *dependencyTreeNode)
	extractDependencyTreeData = func(subTree *dependencyTreeNode) {
		for _, dependency := range subTree.Dependencies {
			// Traversing the dependency tree
			if !dependency.IsRunning {
				extractDependencyTreeData(dependency)
			}

			// Adding used instances table content
			instanceAlreadyAdded := false
			for _, instance := range traversedInstances {
				if instance == dependency.Instance {
					instanceAlreadyAdded = true
					break
				}
			}
			if !instanceAlreadyAdded {
				var usedInstance string
				if dependency.IsRunning {
					usedInstance = "Available in Runtime"
				} else {
					usedInstance = "To be Created"
				}
				var sharedSymbol string
				if dependency.IsShared {
					sharedSymbol = "Shared"
				} else {
					sharedSymbol = " - "
				}
				dependencyData = append(dependencyData, []string{
					dependency.Instance,
					dependency.MetaData.Organization + "/" + dependency.MetaData.Name + ":" +
						dependency.MetaData.Version,
					usedInstance,
					sharedSymbol,
				})
				traversedInstances = append(traversedInstances, dependency.Instance)
			}
		}
	}
	extractDependencyTreeData(tree)
	usedInstance := "To be Created"
	_, err := kubectl.GetCell(tree.Instance)
	if err == nil {
		usedInstance = "Available in Runtime"
	}
	dependencyData = append(dependencyData, []string{
		tree.Instance,
		tree.MetaData.Organization + "/" + tree.MetaData.Name + ":" + tree.MetaData.Version,
		usedInstance,
		" - ",
	})

	// Rendering the instances table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"INSTANCE NAME", "CELL IMAGE", "USED INSTANCE", "SHARED"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{})
	table.AppendBulk(dependencyData)
	fmt.Printf("\n%s:\n\n", util.Bold("Instance to be tested"))
	table.Render()

	// Printing the dependency tree
	fmt.Printf("\n%s:\n\n", util.Bold("Dependency Tree to be Used"))
	var printDependencyTree func(subTree *dependencyTreeNode, nestingLevel int, ancestorBranchPrintRequirement []bool)
	printDependencyTree = func(subTree *dependencyTreeNode, nestingLevel int, ancestorBranchPrintRequirement []bool) {
		var index = 0
		for alias, dependency := range subTree.Dependencies {
			// Adding the dependency tree visualization content
			for j := 0; j < nestingLevel; j++ {
				if ancestorBranchPrintRequirement[j] {
					fmt.Print("   │ ")
				} else {
					fmt.Print("     ")
				}
			}
			if index == len(subTree.Dependencies)-1 {
				fmt.Print("   └")
			} else {
				fmt.Print("   ├")
			}
			fmt.Printf("── %s: %s\n", util.Bold(alias), dependency.Instance)

			// Traversing the dependency tree
			if !dependency.IsRunning {
				printDependencyTree(dependency, nestingLevel+1,
					append(ancestorBranchPrintRequirement, index != len(subTree.Dependencies)-1))
			}
			index++
		}
	}
	if len(tree.Dependencies) > 0 {
		fmt.Printf(" %s\n", util.Bold(tree.Instance))
		printDependencyTree(tree, 0, []bool{})
	} else {
		fmt.Printf(" %s\n", util.Bold("No Dependencies"))
	}
	fmt.Println()

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
	fmt.Println()
	return nil
}

func startTestCellInstance(imageDir string, instanceName string, runningNode *dependencyTreeNode,
	envVars []*environmentVariable, debug bool) error {
	imageTag := fmt.Sprintf("%s/%s:%s", runningNode.MetaData.Organization, runningNode.MetaData.Name,
		runningNode.MetaData.Version)
	balFileName, err := util.GetSourceFileName(filepath.Join(imageDir, constants.ZIP_BALLERINA_SOURCE))
	if err != nil {
		return fmt.Errorf("failed to find source file in Cell Image %s due to %v", imageTag, err)
	}
	balFilePath := filepath.Join(imageDir, constants.ZIP_BALLERINA_SOURCE, balFileName)

	containsTestFunction, err := util.TestMethodExists(balFilePath)
	if err != nil {
		return fmt.Errorf("failed to check whether test function exists in Cell Image %s due to %v", imageTag, err)
	}
	if containsTestFunction {
		// Generating the first level dependency map
		dependencies := map[string]*dependencyInfo{}
		for alias, dependency := range runningNode.Dependencies {
			dependencies[alias] = &dependencyInfo{
				Organization: dependency.MetaData.Organization,
				Name:         dependency.MetaData.Name,
				Version:      dependency.MetaData.Version,
				InstanceName: dependency.Instance,
			}
		}

		// Preparing the dependency instance map
		dependenciesJson, err := json.Marshal(dependencies)
		if err != nil {
			return fmt.Errorf("failed to start the Cell Image %s due to %v", imageTag, err)
		}
		tempTestFileName, err := util.CreateTempExecutableBalFile(balFilePath, "test")
		if err != nil {
			util.ExitWithErrorMessage("Error executing ballerina file", err)
		}
		// Preparing the run command arguments
		cmdArgs := []string{"run"}
		for _, envVar := range envVars {
			cmdArgs = append(cmdArgs, "-e", envVar.Key+"="+envVar.Value)
		}
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
		cmdArgs = append(cmdArgs, tempTestFileName, "test", string(iName), string(dependenciesJson))

		// Calling the test function
		moduleMgr := &util.BLangManager{}
		exePath, err := moduleMgr.GetExecutablePath()
		if err != nil {
			util.ExitWithErrorMessage("Failed to get executable path", err)
		}

		cmd := &exec.Cmd{}
		debugMode := strconv.FormatBool(debug)
		if exePath != "" {
			currentDir, err := os.Getwd()
			if err != nil {
				util.ExitWithErrorMessage("Error in determining working directory", err)
			}
			os.Mkdir(currentDir+string(os.PathSeparator)+"logs", 0777)
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
			os.Mkdir(currentDir+string(os.PathSeparator)+"logs", 0777)
			if string(containerId) == "" {

				cmdDockerRun := exec.Command("docker", "run", "-e", constants.TEST_DEGUB_FLAG+"="+debugMode,
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
				string(iName), string(dependenciesJson))
		}
		defer os.Remove(imageDir)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, constants.CELLERY_IMAGE_DIR_ENV_VAR+"="+imageDir)
		cmd.Env = append(cmd.Env, fmt.Sprintf("DEBUG_MODE=%s", debugMode))
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
			return fmt.Errorf("failed to execute test method in Cell instance %s due to %v", instanceName, err)
		}
		err = cmd.Wait()
		defer os.Remove(tempTestFileName)
		if err != nil {
			return fmt.Errorf("failed to execute test method in Cell instance %s due to %v", instanceName, err)
		}
	}
	return nil
}

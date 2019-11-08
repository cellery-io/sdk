/*
 * Copyright (c) 2018 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package commands

import (
	"bufio"
	"fmt"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/fatih/color"
	"github.com/oxequa/interact"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const CellTemplate = `
import ballerina/config;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    // Hello Component
    // This Components exposes the HTML hello world page
    cellery:Component helloComponent = {
        name: "hello",
        src: {
            image: "wso2cellery/samples-hello-world-webapp"
        },
        ingresses: {
            webUI: <cellery:WebIngress>{ // Web ingress will be always exposed globally.
                port: 80,
                gatewayConfig: {
                    vhost: "hello-world.com",
                    context: "/"
                }
            }
        },
        envVars: {
            HELLO_NAME: { value: "Cellery" }
        }
    };

    // Cell Initialization
    cellery:CellImage helloCell = {
        components: {
            helloComp: helloComponent
        }
    };
    return <@untainted> cellery:createImage(helloCell, iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:CellImage helloCell = check cellery:constructCellImage(iName);
    string vhostName = config:getAsString("VHOST_NAME");
    if (vhostName !== "") {
        cellery:WebIngress web = <cellery:WebIngress>helloCell.components["helloComp"]["ingresses"]["webUI"];
        web.gatewayConfig.vhost = vhostName;
    }

    string helloName = config:getAsString("HELLO_NAME");
    if (helloName !== "") {
        helloCell.components["helloComp"]["envVars"]["HELLO_NAME"].value = helloName;
    }
    return <@untainted> cellery:createInstance(helloCell, iName, instances, startDependencies, shareDependencies);
}
`
const testImports  = `
import ballerina/test;
`
const TestTemplate  = `
import ballerina/test;
import ballerina/io;
import celleryio/cellery;

# Handle creation of instances for running tests
@test:BeforeSuite
function setup() {
    cellery:InstanceState[]|error? result = runInstances(conf);
    cellery:Reference cellEndpoints = <cellery:Reference>cellery:getCellEndpoints(<@untainted> instanceList);
}

# Tests inserting order from an external cell by calling the pet-be gateway
@test:Config {}
function testFunction() {
    io:println("I'm in test function!");
    test:assertTrue(true, msg = "Failed!");
}

# Handle deletion of instances for running tests
@test:AfterSuite
public function cleanUp() {
    error? err = cellery:stopInstances(instanceList);
}
`
func RunInit(projectName string, testArg string) {
	if testArg == ""{
		initProject(projectName)
	} else if testArg == "test" {
		err := initTest(projectName)
		if err != nil {
			util.ExitWithErrorMessage("error occurred while initializing tests", err)
		}
	} else {
		//throw error message
	}
}

func initProject(projectName string) {
	prefix := util.CyanBold("?")
	if projectName == "" {
		err := interact.Run(&interact.Interact{
			Before: func(c interact.Context) error {
				c.SetPrfx(color.Output, prefix)
				return nil
			},
			Questions: []*interact.Question{
				{
					Before: func(c interact.Context) error {
						c.SetPrfx(nil, util.CyanBold("?"))
						c.SetDef("my-project", util.Faint("[my-project]"))
						return nil
					},
					Quest: interact.Quest{
						Msg: util.Bold("Project name: "),
					},
					Action: func(c interact.Context) interface{} {
						projectName, _ = c.Ans().String()
						return nil
					},
				},
			},
		})
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while initializing the project", err)
		}
	}
	currentDir, err := os.Getwd()
	if err != nil {
		util.ExitWithErrorMessage("Error in getting current directory location", err)
	}
	writer, projectDir := createBalFile(currentDir, projectName)
	writeCellTemplate(writer, CellTemplate)
	util.PrintSuccessMessage(fmt.Sprintf("Initialized project in directory: %s", util.Faint(projectDir)))
	util.PrintWhatsNextMessage("build the image",
		"cellery build "+projectName+"/"+projectName+".bal"+" organization/image_name:version")
}

func initTest(balFile string) error {
	moduleMgr := &util.BLangManager{}
	exePath, err := moduleMgr.GetExecutablePath()
	if err != nil {
		util.ExitWithErrorMessage("Failed to get executable path", err)
	}
	balAbsPath, err := filepath.Abs(balFile)
	if err != nil {
		return fmt.Errorf("error occurred while retriving absolute path of project", err)
	}
	balModuleDir := filepath.Dir(balAbsPath)
	balProjectDir := filepath.Dir(balModuleDir)
	err = os.Chdir(balProjectDir)
	if err != nil {
		return fmt.Errorf("error occurred while changing directory", err)
	}
	//Append tests
	//appendTestsToBalFile(balFile)

	writer, projectTestDir := createTestBalFile(balAbsPath)
	writeCellTemplate(writer, TestTemplate)
	util.PrintSuccessMessage(fmt.Sprintf("Initialized tests for cell file in: %s", util.Faint(projectTestDir)))
	//util.PrintWhatsNextMessage("build the image",
	//	"cellery build "+projectName+"/"+projectName+".bal"+" organization/image_name:version")


	//Create ballerina project
	balProjectName := filepath.Base(balModuleDir) + "_proj"
	if exePath != ""{
		cmd := exec.Command(exePath, "new", balProjectName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("error occurred while initializing ballerina project for tests %v", err)
		}
	}

	//Move project as a bal module
	destModulePath := filepath.Join(balProjectDir, balProjectName, "src", filepath.Base(balModuleDir))
	err = util.CopyDir(balModuleDir, destModulePath)
	if err != nil {
		return err
	}
	err = util.RemoveDir(balModuleDir)
	if err != nil {
		return err
	}
	return nil
}

// createBalFile creates a bal file at the current location.
func createBalFile(currentDir, projectName string) (*os.File, string) {
	projectDir := filepath.Join(currentDir, projectName)
	err := util.CreateDir(projectDir)
	if err != nil {
		util.ExitWithErrorMessage("Failed to initialize project", err)
	}

	balFile, err := os.Create(filepath.Join(projectDir, projectName+".bal"))
	if err != nil {
		util.ExitWithErrorMessage("Error in creating Ballerina File", err)
	}
	return balFile, projectDir
}

// writeCellTemplate writes the content to a bal file.
func writeCellTemplate(writer io.Writer, cellTemplate string) {
	balW := bufio.NewWriter(writer)
	_, err := balW.WriteString(cellTemplate)
	if err != nil {
		util.ExitWithErrorMessage("Failed to create writer for cell file", err)
	}
	_ = balW.Flush()
	if err != nil {
		util.ExitWithErrorMessage("Failed to cleanup writer for cell file", err)
	}
}

// createBalFile creates a bal file at the current location.
func createTestBalFile(balFilePath string) (*os.File, string) {
	projectTestDir := filepath.Join(filepath.Dir(balFilePath), "tests")
	err := util.CreateDir(projectTestDir)
	if err != nil {
		util.ExitWithErrorMessage("Failed to initialize tests", err)
	}

	testFileName := strings.TrimSuffix(filepath.Base(balFilePath), path.Ext(balFilePath)) + "_test.bal"
	testBalFile, err := os.Create(filepath.Join(projectTestDir, testFileName))
	if err != nil {
		util.ExitWithErrorMessage("Error in creating Ballerina File", err)
	}
	return testBalFile, projectTestDir
}

func appendTestsToBalFile(balFile string) error {
	tmpBalFile := "tmp.bal"
	_, err := os.Create(tmpBalFile)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(tmpBalFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	//Add test import
	if _, err = f.WriteString(testImports); err != nil {
		return err
	}
	content, err := ioutil.ReadFile(balFile)
	if err != nil {
		return err
	}

	// Convert []byte to string and print to screen
	existingBalText := string(content)
	if _, err = f.WriteString(existingBalText); err != nil {
		return err
	}
	if _, err = f.WriteString(TestTemplate); err != nil {
		return err
	}
	util.CopyFile(tmpBalFile, balFile)
	os.Remove(tmpBalFile)
	return nil
}

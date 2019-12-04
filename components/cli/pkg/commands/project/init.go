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

package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/oxequa/interact"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/util"
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
    cellery:CellImage|cellery:Composite helloCell = cellery:constructImage(<@untainted> iName);
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

const TestTemplate = `
import ballerina/test;
import ballerina/io;
import celleryio/cellery;

# Handle creation of instances for running tests
@test:BeforeSuite
function setup() {
    cellery:TestConfig testConfig = <@untainted>cellery:getTestConfig();
    cellery:runInstances(testConfig);
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
    error? err = cellery:stopInstances();
}
`

func RunInit(cli cli.Cli, projectName string, isBallerinaProject bool) error {
	pathExists, err := util.FileExists(filepath.Join(cli.FileSystem().CurrentDir(), projectName))
	if err != nil {
		return fmt.Errorf("error occurred while checking if filepath exists, %v", err)
	}
	if pathExists {
		return fmt.Errorf("project with name %s already exists", util.Faint(projectName))
	}
	if !isBallerinaProject {
		initProject(cli, projectName)
	} else {
		initBallerinaProject(cli, projectName)
	}
	return nil
}

func initProject(cli cli.Cli, projectName string) error {
	err := getProjectName(&projectName)
	if err != nil {
		return fmt.Errorf("error occurred while initializing the project, %v", err)
	}
	projectDir := filepath.Join(cli.FileSystem().CurrentDir(), projectName)
	if err := util.CreateDir(projectDir); err != nil {
		return fmt.Errorf("failed to initialize project, %v", err)
	}

	if err = writeCellTemplate(filepath.Join(projectDir, projectName+".bal"), CellTemplate); err != nil {
		return fmt.Errorf("failed to create cell file. %v", err)
	}
	util.PrintSuccessMessage(fmt.Sprintf("Initialized project in directory: %s", util.Faint(projectDir)))
	util.PrintWhatsNextMessage("build the image",
		"cellery build "+projectName+"/"+projectName+".bal"+" organization/image_name:version")
	return nil
}

func initBallerinaProject(cli cli.Cli, projectName string) error {
	var workingDir string
	exePath, err := cli.BalExecutor().ExecutablePath()
	if err != nil {
		return err
	}
	if exePath == "" {
		currentTime := time.Now()
		timestamp := currentTime.Format("27065102350415")
		workingDir = filepath.Join(cli.FileSystem().TempDir(), timestamp)
		util.CreateDir(workingDir)
	} else {
		workingDir = cli.FileSystem().CurrentDir()
	}

	err = getProjectName(&projectName)
	if err != nil {
		return fmt.Errorf("error occurred while initializing the project, %v", err)
	}
	moduleName := strings.Replace(projectName, "-", "_", -1)
	if err := cli.BalExecutor().Init(workingDir, projectName, moduleName); err != nil {
		return err
	}

	//Remove unnecessary files from Ballerina project
	moduleDir := filepath.Join(workingDir, projectName, "src", moduleName)
	if err = os.Remove(filepath.Join(moduleDir, "main.bal")); err != nil {
		return err
	}
	if err = os.Remove(filepath.Join(moduleDir, "tests", "main_test.bal")); err != nil {
		return err
	}
	if err = os.Remove(filepath.Join(moduleDir, "Module.md")); err != nil {
		return err
	}

	// Create cell and test files
	if err = writeCellTemplate(filepath.Join(moduleDir, moduleName+".bal"), CellTemplate); err != nil {
		return fmt.Errorf("failed to create cell file. %v", err)
	}
	if err = writeCellTemplate(
		filepath.Join(moduleDir, "tests", moduleName+"_test.bal"), TestTemplate); err != nil {
		return fmt.Errorf("failed to create test file. %v", err)
	}
	if exePath == "" {
		err = util.CopyDir(filepath.Join(workingDir, projectName), filepath.Join(cli.FileSystem().CurrentDir(), projectName))
		if err != nil {
			return fmt.Errorf("error occurred while copying generated ballerina project to %s", cli.FileSystem().CurrentDir())
		}
	}
	util.PrintSuccessMessage(fmt.Sprintf("Initialized Ballerina project: %s", util.Faint(projectName)))
	util.PrintWhatsNextMessage("build the image",
		"cellery build "+projectName+" organization/image_name:version")
	return nil
}

// Get project name from user if not passed to the CLI command
func getProjectName(projectName *string) error {
	if *projectName == "" {
		prefix := util.CyanBold("?")
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
						*projectName, _ = c.Ans().String()
						return nil
					},
				},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// writeCellTemplate writes the content to a bal file.
func writeCellTemplate(cellFilePath string, cellTemplate string) error {
	writer, err := os.Create(cellFilePath)
	if err != nil {
		return fmt.Errorf("error in creating Ballerina File, %v", err)
	}
	balW := bufio.NewWriter(writer)
	_, err = balW.WriteString(cellTemplate)
	if err != nil {
		return fmt.Errorf("failed to create writer for cell file, %v", err)
	}
	_ = balW.Flush()
	if err != nil {
		return fmt.Errorf("failed to cleanup writer for cell file, %v", err)
	}
	return nil
}

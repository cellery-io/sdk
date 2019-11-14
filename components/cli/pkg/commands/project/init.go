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
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cellery-io/sdk/components/cli/cli"

	"github.com/fatih/color"
	"github.com/oxequa/interact"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
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

const TestTemplate = `
import ballerina/test;
import ballerina/io;
import celleryio/cellery;

# Handle creation of instances for running tests
@test:BeforeSuite
function setup() {
    cellery:ImageName iName = <cellery:ImageName>cellery:getCellImage();
    cellery:InstanceState[]|error? result = run(iName, {}, true, true);
    if (result is error) {
        cellery:InstanceState iNameState = {
            iName : iName, 
            isRunning: true
        };
        instanceList[instanceList.length()] = <@untainted> iNameState;
    } else {
        instanceList = <cellery:InstanceState[]>result;
    }
    cellery:Reference petBeUrls = <cellery:Reference>cellery:getCellEndpoints(<@untainted> instanceList);
	PET_BE_CONTROLLER_ENDPOINT= <string>petBeUrls["controller_ingress_api_url"];
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
const projStructure = `
New project structure:

%s/
├── Ballerina.toml
└── src
    └── %s
        ├── %s
        └── tests
            └── %s
`

func RunInit(cli cli.Cli, projectName string, testArg string) error {
	if testArg == "" {
		initProject(cli, projectName)
	} else {
		err := initTest(cli, projectName)
		if err != nil {
			return err
		}
	}
	return nil
}

func initProject(cli cli.Cli, projectName string) error {
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
			return fmt.Errorf("error occurred while initializing the project, %v", err)
		}
	}
	projectDir := filepath.Join(cli.FileSystem().CurrentDir(), projectName)
	if err := util.CreateDir(projectDir); err != nil {
		return fmt.Errorf("failed to initialize project, %v", err)
	}
	writer, err := os.Create(filepath.Join(projectDir, projectName+".bal"))
	if err != nil {
		return fmt.Errorf("error in creating Ballerina File, %v", err)
	}
	balW := bufio.NewWriter(writer)
	if _, err = balW.WriteString(CellTemplate); err != nil {
		return fmt.Errorf("failed to create writer for cell file, %v", err)
	}
	if _ = balW.Flush(); err != nil {
		return fmt.Errorf("failed to cleanup writer for cell file, %v", err)
	}
	util.PrintSuccessMessage(fmt.Sprintf("Initialized project in directory: %s", util.Faint(projectDir)))
	util.PrintWhatsNextMessage("build the image",
		"cellery build "+projectName+"/"+projectName+".bal"+" organization/image_name:version")
	return nil
}

func initTest(cli cli.Cli, balFile string) error {
	balAbsPath, err := filepath.Abs(balFile)
	if err != nil {
		return err
	}
	balModuleDir := filepath.Dir(balAbsPath)
	balProjectDir := filepath.Dir(balModuleDir)
	err = os.Chdir(balProjectDir)
	if err != nil {
		return err
	}

	writer, projectTestDir, err := createTestBalFile(balAbsPath)
	err = writeCellTemplate(writer, TestTemplate)
	if err != nil {
		return err
	}

	//Create ballerina project
	balProjectName := filepath.Base(balModuleDir) + "_proj"
	projExists, err := util.FileExists(balProjectName)
	if err != nil {
		return fmt.Errorf("error while checking if project %s exists", balProjectName)
	}
	if projExists {
		return fmt.Errorf("%s project already exists", balProjectName)
	}

	if err := cli.BalExecutor().Init(balModuleDir); err != nil {
		return err
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

	util.PrintSuccessMessage(fmt.Sprintf("Initialized tests for %s in: %s", balFile, util.Faint(*projectTestDir)))
	util.PrintInfoMessage(fmt.Sprintf(projStructure, balProjectName,
		util.Faint(filepath.Base(balModuleDir)), util.Faint(filepath.Base(balAbsPath)),
		strings.TrimSuffix(filepath.Base(balAbsPath), path.Ext(balAbsPath))+"_test.bal"))
	return nil
}

// writeCellTemplate writes the content to a bal file.
func writeCellTemplate(writer io.Writer, cellTemplate string) error {
	balW := bufio.NewWriter(writer)
	_, err := balW.WriteString(cellTemplate)
	if err != nil {
		return fmt.Errorf("failed to create writer for cell file, %v", err)
	}
	_ = balW.Flush()
	if err != nil {
		return fmt.Errorf("failed to cleanup writer for cell file, %v", err)
	}
	return nil
}

// createBalFile creates a bal file at the current location.
func createTestBalFile(balFilePath string) (*os.File, *string, error) {
	projectTestDir := filepath.Join(filepath.Dir(balFilePath), "tests")
	err := util.CreateDir(projectTestDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize tests, %v", err)
	}

	testFileName := strings.TrimSuffix(filepath.Base(balFilePath), path.Ext(balFilePath)) + "_test.bal"
	testBalFile, err := os.Create(filepath.Join(projectTestDir, testFileName))
	if err != nil {
		return nil, nil, fmt.Errorf("error in creating Ballerina File, %v", err)
	}
	return testBalFile, &projectTestDir, nil
}

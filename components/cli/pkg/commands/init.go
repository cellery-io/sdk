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
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/oxequa/interact"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

const CellTemplate = "import ballerina/config;\n" +
	"import celleryio/cellery;\n" +
	"\n" +
	"public function build(cellery:ImageName iName) returns error? {\n" +
	"    // Hello Component\n" +
	"    // This Components exposes the HTML hello world page\n" +
	"    cellery:Component helloComponent = {\n" +
	"        name: \"hello\",\n" +
	"        source: {\n" +
	"            image: \"wso2cellery/samples-hello-world-webapp\"\n" +
	"        },\n" +
	"        ingresses: {\n" +
	"            webUI: <cellery:WebIngress>{ // Web ingress will be always exposed globally.\n" +
	"                port: 80,\n" +
	"                gatewayConfig: {\n" +
	"                    vhost: \"hello-world.com\",\n" +
	"                    context: \"/\"\n" +
	"                }\n" +
	"            }\n" +
	"        },\n" +
	"        envVars: {\n" +
	"            HELLO_NAME: { value: \"Cellery\" }\n" +
	"        }\n" +
	"    };\n" +
	"\n" +
	"    // Cell Initialization\n" +
	"    cellery:CellImage helloCell = {\n" +
	"        components: {\n" +
	"            helloComp: helloComponent\n" +
	"        }\n" +
	"    };\n" +
	"    return cellery:createImage(helloCell, untaint iName);\n" +
	"}\n" +
	"\n" +
	"public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {\n" +
	"    cellery:CellImage helloCell = check cellery:constructCellImage(untaint iName);\n" +
	"    string vhostName = config:getAsString(\"VHOST_NAME\");\n" +
	"    if (vhostName !== \"\") {\n" +
	"        cellery:WebIngress web = <cellery:WebIngress>helloCell.components.helloComp.ingresses.webUI;" +
	"\n" +
	"        web.gatewayConfig.vhost = vhostName;\n" +
	"    }\n" +
	"\n" +
	"    string helloName = config:getAsString(\"HELLO_NAME\");\n" +
	"    if (helloName !== \"\") {\n" +
	"        helloCell.components.helloComp.envVars.HELLO_NAME.value = helloName;\n" +
	"    }\n" +
	"    return cellery:createInstance(helloCell, iName, instances, startDependencies, shareDependencies);\n" +
	"}\n"

func RunInit(projectName string) {
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

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
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/oxequa/interact"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

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

	cellTemplate := "import ballerina/config;\n" +
		"import celleryio/cellery;\n" +
		"\n" +
		"// Hello Component\n" +
		"// This Components exposes the HTML hello world page\n" +
		"cellery:Component helloComponent = {\n" +
		"    name: \"hello\",\n" +
		"    source: {\n" +
		"        image: \"wso2cellery/samples-hello-world-webapp\"\n" +
		"    },\n" +
		"    ingresses: {\n" +
		"        webUI: <cellery:WebIngress> { // Web ingress will be always exposed globally.\n" +
		"            port: 80,\n" +
		"            gatewayConfig: {\n" +
		"                vhost: \"hello-world.com\",\n" +
		"                context: \"/\"\n" +
		"            }\n" +
		"        }\n" +
		"    },\n" +
		"    envVars: {\n" +
		"        HELLO_NAME: {value: \"Cellery\"}\n" +
		"    }\n" +
		"};\n" +
		"\n" +
		"// Cell Initialization\n" +
		"cellery:CellImage helloCell = {\n" +
		"    components: {\n" +
		"        helloComp: helloComponent\n" +
		"    }\n" +
		"};\n" +
		"\n" +
		"# The Cellery Lifecycle Build method which is invoked for building the Cell Image.\n" +
		"#\n" +
		"# + iName - The Image name\n" +
		"# + return - The created Cell Image\n" +
		"public function build(cellery:ImageName iName) returns error? {\n" +
		"    return cellery:createImage(helloCell, iName);\n" +
		"}\n" +
		"\n" +
		"# The Cellery Lifecycle Run method which is invoked for creating a Cell Instance.\n" +
		"#\n" +
		"# + iName - The Image name\n" +
		"# + instances - The map dependency instances of the Cell instance to be created\n" +
		"# + return - The Cell instance\n" +
		"public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {\n" +
		"    string vhostName = config:getAsString(\"VHOST_NAME\");\n" +
		"    if (vhostName !== \"\"){\n" +
		"        cellery:WebIngress web = <cellery:WebIngress> helloComponent.ingresses.webUI;\n" +
		"        web.gatewayConfig.vhost = vhostName;\n" +
		"    }\n" +
		"\n" +
		"    string helloName = config:getAsString(\"HELLO_NAME\");\n" +
		"    if (helloName !== \"\"){\n" +
		"        helloComponent.envVars.HELLO_NAME.value = helloName;\n" +
		"    }\n" +
		"    return cellery:createInstance(helloCell, iName);\n" +
		"}\n" +
		"\n"

	currentDir, err := os.Getwd()
	if err != nil {
		util.ExitWithErrorMessage("Error in getting current directory location", err)
	}

	projectDir := filepath.Join(currentDir, projectName)
	err = util.CreateDir(projectDir)
	if err != nil {
		util.ExitWithErrorMessage("Failed to initialize project", err)
	}

	balFile, err := os.Create(filepath.Join(projectDir, projectName+".bal"))
	if err != nil {
		util.ExitWithErrorMessage("Error in creating Ballerina File", err)
	}
	defer func() {
		err = balFile.Close()
		if err != nil {
			util.ExitWithErrorMessage("Failed to clean up", err)
		}
	}()
	balW := bufio.NewWriter(balFile)
	_, err = balW.WriteString(cellTemplate)
	if err != nil {
		util.ExitWithErrorMessage("Failed to create cell file", err)
	}
	_ = balW.Flush()
	if err != nil {
		util.ExitWithErrorMessage("Failed to create cell file", err)
	}

	util.PrintSuccessMessage(fmt.Sprintf("Initialized project in directory: %s", util.Faint(projectDir)))
	util.PrintWhatsNextMessage("build the image",
		"cellery build "+projectName+"/"+projectName+".bal"+" organization/image_name:version")
}

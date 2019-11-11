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

// RunInit initializes a cellery project.
func RunInit(cli cli.Cli, projectName string) error {
	prefix := util.CyanBold("?")
	if projectName == "" {
		if err := interact.Run(&interact.Interact{
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
		}); err != nil {
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

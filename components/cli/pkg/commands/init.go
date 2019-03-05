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

func RunInit() {
	prefix := util.CyanBold("?")

	projectName := ""
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

	cellTemplate := "import celleryio/cellery;\n" +
		"\n" +
		"cellery:Component helloWorldComp = {\n" +
		"    name: \"hello-world\",\n" +
		"    source: {\n" +
		"        image: \"sumedhassk/hello-world:1.0.0\"\n" +
		"    },\n" +
		"    ingresses: {\n" +
		"        hello: new cellery:HTTPIngress(\n" +
		"                   9090,\n" +
		"\n" +
		"                   \"hello\",\n" +
		"                   [\n" +
		"                       {\n" +
		"                           path: \"/*\",\n" +
		"                           method: \"GET\"\n" +
		"                       }\n" +
		"                   ]\n" +
		"\n" +
		"        )\n" +
		"    }\n" +
		"};\n" +
		"\n" +
		"cellery:CellImage helloCell = new();\n" +
		"\n" +
		"public function build(string orgName, string imageName, string imageVersion) {\n" +
		"    helloCell.addComponent(helloWorldComp);\n" +
		"\n" +
		"    helloCell.exposeGlobal(helloWorldComp);\n" +
		"\n" +
		"    _= cellery:createImage(helloCell, orgName, imageName, imageVersion);\n" +
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
		"cellery build "+projectName+"/"+projectName+".bal"+" -t [repo/]organization/image_name:version")
}

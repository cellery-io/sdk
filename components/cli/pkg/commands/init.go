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
	"os/user"
	"path/filepath"

	"github.com/fatih/color"
	i "github.com/oxequa/interact"

	"github.com/celleryio/sdk/components/cli/pkg/util"
)

func RunInit() error {

	userVar, err := user.Current()
	if err != nil {
		panic(err)
	}

	prefix := util.CyanBold("?")
	orgName := ""
	projectName := ""
	projectVersion := ""
	projectPath := ""

	i.Run(&i.Interact{
		Before: func(c i.Context) error {
			c.SetPrfx(color.Output, prefix)
			return nil
		},
		Questions: []*i.Question{
			{
				Before: func(c i.Context) error {
					c.SetPrfx(nil, util.CyanBold("?"))
					c.SetDef(userVar.Username, util.Faint("["+userVar.Username+"]"))
					return nil
				},
				Quest: i.Quest{
					Msg: util.Bold("Organization name: "),
				},
				Action: func(c i.Context) interface{} {
					orgName, _ = c.Ans().String()
					return nil
				},
			},
			{
				Before: func(c i.Context) error {
					c.SetPrfx(nil, util.CyanBold("?"))
					c.SetDef("my-project", util.Faint("[my-project]"))
					return nil
				},
				Quest: i.Quest{
					Msg: util.Bold("Project name: "),
				},
				Action: func(c i.Context) interface{} {
					projectName, _ = c.Ans().String()
					return nil
				},
			},
			{
				Before: func(c i.Context) error {
					c.SetPrfx(nil, util.CyanBold("?"))
					c.SetDef("1.0.0", util.Faint("[1.0.0]"))
					return nil
				},
				Quest: i.Quest{
					Msg: util.Bold("Project version: "),
				},
				Action: func(c i.Context) interface{} {
					projectVersion, _ = c.Ans().String()
					return nil
				},
			},
		},
	})

	cellTemplate := "import ballerina/io;\n" +
		"import celleryio/cellery;\n" +
		"\n" +
		"cellery:Component helloWorldComp = {\n" +
		"    name: \"HelloWorld\",\n" +
		"    source: {\n" +
		"        image: \"sumedhassk/hello-world:1.0.0\"\n" +
		"    },\n" +
		"    ingresses: {\n" +
		"        \"hello\": new cellery:HTTPIngress(9090, \"hello\",\n" +
		"            [\n" +
		"                {\n" +
		"                    path: \"/\",\n" +
		"                    method: \"GET\"\n" +
		"                }\n" +
		"            ]\n" +
		"        )\n" +
		"    }\n" +
		"};\n" +
		"\n" +
		"cellery:CellImage helloCell = new(\"Hello-World\");\n" +
		"\n" +
		"public function build() {\n" +
		"    helloCell.addComponent(helloWorldComp);\n" +
		"\n" +
		"    helloCell.exposeGlobalAPI(helloWorldComp);\n" +
		"\n" +
		"    var out = cellery:createImage(helloCell);\n" +
		"    if (out is boolean) {\n" +
		"        io:println(\"Hello Cell Built successfully.\");\n" +
		"    }\n" +
		"}\n" +
		"\n"

	tomlTemplate := "[project]\n" +
		"organization = \"" + orgName + "\"\n" +
		"version = \"" + projectVersion + "\"\n"

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("Error in getting current directory location: " + err.Error())
		os.Exit(1)
	}

	if _, err := os.Stat(dir + "/" + projectName); os.IsNotExist(err) {
		projectPath = dir + "/" + projectName
		os.Mkdir(projectPath, os.ModePerm)
	}

	balFile, err := os.Create(projectPath + "/" + projectName + ".bal")
	if err != nil {
		fmt.Println("Error in creating Ballerina File: " + err.Error())
		os.Exit(1)
	}
	defer balFile.Close()
	balW := bufio.NewWriter(balFile)
	balW.WriteString(fmt.Sprintf("%s", cellTemplate))
	balW.Flush()

	tomlFile, err := os.Create(projectPath + "/Cellery.toml")
	if err != nil {
		fmt.Println("Error in creating Toml File: " + err.Error())
		os.Exit(1)
	}
	defer tomlFile.Close()
	tomlW := bufio.NewWriter(tomlFile)
	tomlW.WriteString(fmt.Sprintf("%s", tomlTemplate))
	tomlW.Flush()

	fmt.Println(util.GreenBold("\n\U00002714") + " Initialized project in directory: " + util.Faint(projectPath))
	util.PrintWhatsNextMessage("cellery build " + projectName + ".bal")
	return nil
}

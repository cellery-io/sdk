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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/ghodss/yaml"
)

func RunListComponents(name string) {
	instancePattern, _ := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), name)
	if instancePattern {
		displayComponentsTable(getCellInstanceComponents(name))
	} else {
		displayComponentsTable(getCellImageCompoents(name))
	}
}

func getCellImageCompoents(cellImage string) []string {
	var components []string
	cellYamlContent := util.ReadCellImageYaml(cellImage)
	cellImageContent := &util.Cell{}
	err := yaml.Unmarshal(cellYamlContent, cellImageContent)
	if err != nil {
		util.ExitWithErrorMessage("Error while reading cell image content", err)
	}
	for i := 0; i < len(cellImageContent.CellSpec.ComponentTemplates); i++ {
		components = append(components, cellImageContent.CellSpec.ComponentTemplates[i].Metadata.Name)
	}
	return components
}

func getCellInstanceComponents(cellName string) []string {
	var components []string
	cmd := exec.Command("kubectl", "get", "services", "-l", constants.GROUP_NAME+"/cell="+cellName, "-o", "json")
	outfile, errPrint := os.Create("./out.txt")
	if errPrint != nil {
		util.ExitWithErrorMessage("Error occurred while fetching cell status", errPrint)
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()
	err := cmd.Start()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while fetching components", err)
	}
	err = cmd.Wait()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while fetching components", err)
	}

	outputByteArray, err := ioutil.ReadFile("./out.txt")
	os.Remove("./out.txt")
	jsonOutput := &util.Service{}

	errJson := json.Unmarshal(outputByteArray, jsonOutput)
	if errJson != nil {
		fmt.Println(errJson)
	}
	if len(jsonOutput.Items) == 0 {
		util.ExitWithErrorMessage("Error listing components", fmt.Errorf("Cannot find cell: %v \n", cellName))
	} else {
		for i := 0; i < len(jsonOutput.Items); i++ {
			var name string
			name = strings.Replace(jsonOutput.Items[i].Metadata.Name, cellName+"--", "", -1)
			name = strings.Replace(name, "-service", "", -1)
			components = append(components, name)
		}
	}
	return components
}

func displayComponentsTable(components []string) {
	var tableData [][]string

	for i := 0; i < len(components); i++ {
		component := []string{components[i]}
		tableData = append(tableData, component)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"COMPONENT NAME"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(tablewriter.Colors{tablewriter.FgHiBlueColor})

	table.AppendBulk(tableData)
	table.Render()
}

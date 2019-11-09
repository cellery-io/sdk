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
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
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
	cellYamlContent := image.ReadCellImageYaml(cellImage)
	cellImageContent := &image.Cell{}
	err := yaml.Unmarshal(cellYamlContent, cellImageContent)
	if err != nil {
		util.ExitWithErrorMessage("Error while reading cell image content", err)
	}
	for _, component := range cellImageContent.Spec.Components {
		components = append(components, component.Metadata.Name)
	}
	return components
}

func getCellInstanceComponents(cellName string) []string {
	var components []string
	services, err := kubernetes.GetServices(cellName)
	if err != nil {
		util.ExitWithErrorMessage("Error getting list of components", err)
	}
	if len(services.Items) == 0 {
		util.ExitWithErrorMessage("Error listing components", fmt.Errorf("cannot find cell: %v \n", cellName))
	} else {
		for i := 0; i < len(services.Items); i++ {
			var name string
			name = strings.Replace(services.Items[i].Metadata.Name, cellName+"--", "", -1)
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

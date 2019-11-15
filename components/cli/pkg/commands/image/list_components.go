/*
 * Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package image

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
)

func RunListComponents(cli cli.Cli, name string) error {
	var err error
	instancePattern, _ := regexp.MatchString(fmt.Sprintf("^%s$", celleryIdPattern), name)
	if instancePattern {
		var instanceComponents []string
		if instanceComponents, err = getCellInstanceComponents(cli, name); err != nil {
			return err
		}
		displayComponentsTable(instanceComponents)
	} else {
		imageComponents, err := getCellImageCompoents(cli, name)
		if err != nil {
			return err
		}
		displayComponentsTable(imageComponents)
	}
	return nil
}

func getCellImageCompoents(cli cli.Cli, cellImage string) ([]string, error) {
	var components []string
	cellYamlContent, err := image.ReadCellImageYaml(cli.FileSystem().Repository(), cellImage)
	if err != nil {
		return nil, fmt.Errorf("error while reading cell image content, %v", err)
	}
	cellImageContent := &image.Cell{}
	if err := yaml.Unmarshal(cellYamlContent, cellImageContent); err != nil {
		return nil, fmt.Errorf("error while unmarshalling cell image content, %v", err)
	}
	for _, component := range cellImageContent.Spec.Components {
		components = append(components, component.Metadata.Name)
	}
	return components, nil
}

func getCellInstanceComponents(cli cli.Cli, cellName string) ([]string, error) {
	var components []string
	services, err := cli.KubeCli().GetServices(cellName)
	if err != nil {
		return nil, fmt.Errorf("error getting list of components, %v", err)
	}
	if len(services.Items) == 0 {
		return nil, fmt.Errorf("error listing components, %v", fmt.Errorf("cannot find cell: %v \n", cellName))
	} else {
		for i := 0; i < len(services.Items); i++ {
			var name string
			name = strings.Replace(services.Items[i].Metadata.Name, cellName+"--", "", -1)
			name = strings.Replace(name, "-service", "", -1)
			components = append(components, name)
		}
	}
	return components, nil
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

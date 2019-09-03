/*
 * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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

	"github.com/cellery-io/sdk/components/cli/pkg/routing"

	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
)

func RunListDependencies(instanceName string) error {
	cellInst, err := kubectl.GetCell(instanceName)
	if err != nil {
		return err
	}
	dependencies, err := routing.ExtractDependencies(cellInst.CellMetaData.Annotations.Dependencies)
	if err != nil {
		return err
	}
	if len(dependencies) == 0 {
		// No dependencies
		return fmt.Errorf("no dependencies found in cell instance %s", instanceName)
	}
	var tableData [][]string
	for _, dependency := range dependencies {
		record := []string{dependency["instance"], dependency["name"], dependency["version"]}
		tableData = append(tableData, record)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"CELL INSTANCE", "IMAGE", "VERSION"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
	)
	table.SetColumnColor(
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
	)

	table.AppendBulk(tableData)
	table.Render()

	return nil
}

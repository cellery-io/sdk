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

package instance

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunListInstances(cli cli.Cli) error {
	var err error
	var cellTableData, compositeTableData [][]string
	if cellTableData, err = getCellTableData(cli); err != nil {
		return fmt.Errorf("error getting cell data, %v", err)
	}
	if len(cellTableData) > 0 {
		displayCellTable(cli, cellTableData)
	}
	if compositeTableData, err = getCompositeTableData(cli); err != nil {
		return fmt.Errorf("error getting composite data, %v", err)
	}
	if len(compositeTableData) > 0 {
		displayCompositeTable(cli, compositeTableData)
	}
	if len(cellTableData) == 0 && len(compositeTableData) == 0 {
		fmt.Fprintln(cli.Out(), "No running instances.")
	}
	return nil
}

func displayCellTable(cli cli.Cli, tableData [][]string) error {
	if len(tableData) > 0 {
		fmt.Fprintf(cli.Out(), "\n %s\n", util.Bold("Cell Instances:"))
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"INSTANCE", "IMAGE", "STATUS", "GATEWAY", "COMPONENTS", "AGE"})
		table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
		table.SetAlignment(3)
		table.SetRowSeparator("-")
		table.SetCenterSeparator(" ")
		table.SetColumnSeparator(" ")
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold})
		table.SetColumnColor(
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{})

		table.AppendBulk(tableData)
		table.Render()
	}
	return nil
}

func displayCompositeTable(cli cli.Cli, tableData [][]string) error {
	if len(tableData) > 0 {
		fmt.Fprintf(cli.Out(), " \n %s\n", util.Bold("Composite Instances:"))
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"INSTANCE", "IMAGE", "STATUS", "COMPONENTS", "AGE"})
		table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
		table.SetAlignment(3)
		table.SetRowSeparator("-")
		table.SetCenterSeparator(" ")
		table.SetColumnSeparator(" ")
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold})
		table.SetColumnColor(
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{})

		table.AppendBulk(tableData)
		table.Render()
	}
	return nil
}

func getCellTableData(cli cli.Cli) ([][]string, error) {
	var tableData [][]string
	cells, err := cli.KubeCli().GetCells()
	if err != nil {
		return nil, fmt.Errorf("error getting information of cells, %v", err)
	}
	for i := 0; i < len(cells); i++ {
		age := util.GetDuration(util.ConvertStringToTime(cells[i].CellMetaData.CreationTimestamp))
		instance := cells[i].CellMetaData.Name
		cellImage := cells[i].CellMetaData.Annotations.Organization + "/" + cells[i].CellMetaData.Annotations.Name + ":" + cells[i].CellMetaData.Annotations.Version
		gateway := cells[i].CellStatus.Gateway
		components := cells[i].CellStatus.ServiceCount
		status := cells[i].CellStatus.Status
		tableRecord := []string{instance, cellImage, status, gateway, strconv.Itoa(components), age}
		tableData = append(tableData, tableRecord)
	}
	return tableData, nil
}

func getCompositeTableData(cli cli.Cli) ([][]string, error) {
	compositeData, err := cli.KubeCli().GetComposites()
	if err != nil {
		return nil, fmt.Errorf("error getting information of composites, %v", err)
	}
	var tableData [][]string
	for i := 0; i < len(compositeData); i++ {
		age := util.GetDuration(util.ConvertStringToTime(compositeData[i].CompositeMetaData.CreationTimestamp))
		instance := compositeData[i].CompositeMetaData.Name
		cellImage := compositeData[i].CompositeMetaData.Annotations.Organization + "/" +
			compositeData[i].CompositeMetaData.Annotations.Name + ":" +
			compositeData[i].CompositeMetaData.Annotations.Version
		components := compositeData[i].CompositeStatus.ServiceCount
		status := compositeData[i].CompositeStatus.Status
		tableRecord := []string{instance, cellImage, status, strconv.Itoa(components), age}
		tableData = append(tableData, tableRecord)
	}
	return tableData, nil
}

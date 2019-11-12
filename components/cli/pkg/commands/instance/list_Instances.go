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
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunListInstances(cli cli.Cli) error {
	if err := displayCellTable(cli); err != nil {
		return fmt.Errorf("error displaying cell table, %v", err)
	}
	displayCompositeTable(cli)
	return nil
}

func displayCellTable(cli cli.Cli) error {
	tableData, err := getCellTableData(cli.KubeCli())
	if err != nil {
		return fmt.Errorf("error getting cell table data, %v", err)
	}
	if len(tableData) > 0 {
		fmt.Printf("\n %s\n", util.Bold("Cell Instances:"))
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
	} else {
		fmt.Println(cli.Out(), "No running cell instances.")
	}
	return nil
}

func displayCompositeTable(cli cli.Cli) {
	compositeData, err := cli.KubeCli().GetComposites()
	if err != nil {
		util.ExitWithErrorMessage("Error getting information of composites", err)
	}
	if len(compositeData) > 0 {
		fmt.Printf(" \n %s\n", util.Bold("Composite Instances:"))

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
}

func getCellTableData(kubecli kubernetes.KubeCli) ([][]string, error) {
	var tableData [][]string
	cells, err := kubecli.GetCells()
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

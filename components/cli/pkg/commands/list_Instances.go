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
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunListInstances(verboseMode bool) {
	cells, err := kubectl.GetCells(verboseMode)
	if err != nil {
		util.ExitWithErrorMessage("Error getting information of cells", err)
	}
	displayCellTable(cells)
}

func displayCellTable(cellData kubectl.Cells) {
	var tableData [][]string

	for i := 0; i < len(cellData.Items); i++ {
		age := util.GetDuration(util.ConvertStringToTime(cellData.Items[i].CellMetaData.CreationTimestamp))
		instance := cellData.Items[i].CellMetaData.Name
		cellImage := cellData.Items[i].CellMetaData.Annotations.Organization + "/" + cellData.Items[i].CellMetaData.Annotations.Name + ":" + cellData.Items[i].CellMetaData.Annotations.Version
		gateway := cellData.Items[i].CellStatus.Gateway
		components := cellData.Items[i].CellStatus.ServiceCount
		status := cellData.Items[i].CellStatus.Status
		tableRecord := []string{instance, cellImage, status, gateway, strconv.Itoa(components), age}
		tableData = append(tableData, tableRecord)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"INSTANCE", "CELL IMAGE", "STATUS", "GATEWAY", "COMPONENTS", "AGE"})
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

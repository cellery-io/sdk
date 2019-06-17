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
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"

	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunStatus(cellName string, verboseMode bool) {
	cellCreationTime, cellStatus, err := getCellSummary(cellName, verboseMode)
	if err != nil {
		util.ExitWithErrorMessage("Error running cell status", fmt.Errorf("cannot find cell %v \n", cellName))
	}
	displayStatusSummaryTable(cellCreationTime, cellStatus)
	fmt.Println()
	fmt.Println("  -COMPONENTS-")
	fmt.Println()
	pods, err := kubectl.GetPods(cellName, verboseMode)
	if err != nil {
		util.ExitWithErrorMessage("Error getting pods information of cell "+cellName, err)
	}
	displayStatusDetailedTable(pods, cellName)
}

func getCellSummary(cellName string, verboseMode bool) (cellCreationTime, cellStatus string, err error) {
	cellCreationTime = ""
	cellStatus = ""
	cell, err := kubectl.GetCell(cellName, verboseMode)
	if err != nil {
		util.ExitWithErrorMessage("Error getting information of cell "+cellName, err)
	}

	duration := util.GetDuration(util.ConvertStringToTime(cell.CellMetaData.CreationTimestamp))
	cellStatus = cell.CellStatus.Status

	return duration, cellStatus, err
}

func displayStatusSummaryTable(cellCreationTime, cellStatus string) error {
	tableData := []string{cellCreationTime, cellStatus}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"CREATED", "STATUS"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(
		tablewriter.Colors{},
		tablewriter.Colors{})

	table.Append(tableData)
	table.Render()

	return nil
}

func displayStatusDetailedTable(pods kubectl.Pods, cellName string) error {
	var tableData [][]string
	for _, pod := range pods.Items {
		name := strings.Replace(strings.Split(pod.MetaData.Name, "-deployment-")[0], cellName+"--", "", -1)
		state := pod.PodStatus.Phase

		if strings.EqualFold(state, "Running") {
			duration := util.GetDuration(util.ConvertStringToTime(pod.PodStatus.Conditions[1].LastTransitionTime))
			state = "Up for " + duration
		}
		status := []string{name, state}
		tableData = append(tableData, status)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "STATUS"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{})

	table.AppendBulk(tableData)
	table.Render()

	return nil
}

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

	"github.com/olekukonko/tablewriter"

	errorpkg "github.com/cellery-io/sdk/components/cli/pkg/error"
	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunStatus(instance string) {
	creationTime, status, err := getCellSummary(instance)
	var canBeComposite bool
	if err != nil {
		if cellNotFound, _ := errorpkg.IsCellInstanceNotFoundError(instance, err); cellNotFound {
			// could be a composite
			canBeComposite = true
		} else {
			util.ExitWithErrorMessage("Error running status command ", err)
		}
	}
	if canBeComposite {
		creationTime, status, err = getCompositeSummary(instance)
		if err != nil {
			if compositeNotFound, _ := errorpkg.IsCompositeInstanceNotFoundError(instance, err); compositeNotFound {
				// given instance name does not correspond either to a cell or a composite
				util.ExitWithErrorMessage("Error running status command ", fmt.Errorf("No instance found with name %s", instance))
			} else {
				util.ExitWithErrorMessage("Error running status command", err)
			}
		}
		displayCompositeStatus(instance, creationTime, status)
	} else {
		displayCellStatus(instance, creationTime, status)
	}
}

func displayCellStatus(instance, cellCreationTime, cellStatus string) {
	displayStatusSummaryTable(cellCreationTime, cellStatus)
	fmt.Println()
	fmt.Println("  -COMPONENTS-")
	fmt.Println()
	pods, err := kubectl.GetPodsForCell(instance)
	if err != nil {
		util.ExitWithErrorMessage("Error getting pods information of cell "+instance, err)
	}
	displayStatusDetailedTable(pods, instance)
}

func displayCompositeStatus(instance, cellCreationTime, cellStatus string) {
	displayStatusSummaryTable(cellCreationTime, cellStatus)
	fmt.Println()
	fmt.Println("  -COMPONENTS-")
	fmt.Println()
	pods, err := kubectl.GetPodsForComposite(instance)
	if err != nil {
		util.ExitWithErrorMessage("Error getting pods information of composite "+instance, err)
	}
	displayStatusDetailedTable(pods, instance)
}

func getCellSummary(cellName string) (cellCreationTime, cellStatus string, err error) {
	cellCreationTime = ""
	cellStatus = ""
	cell, err := kubectl.GetCell(cellName)
	if err != nil {
		return "", cellStatus, err
	}
	// Get the time since cell instance creation
	duration := util.GetDuration(util.ConvertStringToTime(cell.CellMetaData.CreationTimestamp))
	// Get the current status of the cell
	cellStatus = cell.CellStatus.Status
	return duration, cellStatus, err
}

func getCompositeSummary(compName string) (compCreationTime, compStatus string, err error) {
	compCreationTime = ""
	compStatus = ""
	composite, err := kubectl.GetComposite(compName)
	if err != nil {
		return "", compStatus, err
	}
	// Get the time since composite instance creation
	duration := util.GetDuration(util.ConvertStringToTime(composite.CompositeMetaData.CreationTimestamp))
	// Get the current status of the composite
	compStatus = composite.CompositeStatus.Status
	return duration, compStatus, err
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
			// Get the time since pod's last transition to running state
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

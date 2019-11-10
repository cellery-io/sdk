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

package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/cellery-io/sdk/components/cli/kubernetes"
	errorpkg "github.com/cellery-io/sdk/components/cli/pkg/error"
	"github.com/cellery-io/sdk/components/cli/pkg/routing"
	"github.com/cellery-io/sdk/components/cli/pkg/util"

	"github.com/olekukonko/tablewriter"
)

func RunListDependencies(instanceName string) error {
	var depJson string
	var canBeComposite bool
	cellInst, err := kubernetes.GetCell(instanceName)
	if err != nil {
		if cellNotFound, _ := errorpkg.IsCellInstanceNotFoundError(instanceName, err); cellNotFound {
			canBeComposite = true
		} else {
			util.ExitWithErrorMessage("Failed to check available Cells", err)
		}
	} else {
		depJson = cellInst.CellMetaData.Annotations.Dependencies
	}

	if canBeComposite {
		compositeInst, err := kubernetes.GetComposite(instanceName)
		if err != nil {
			if compositeNotFound, _ := errorpkg.IsCompositeInstanceNotFoundError(instanceName, err); compositeNotFound {
				util.ExitWithErrorMessage("Failed to retrieve dependencies of "+instanceName,
					errors.New(instanceName+" instance not available in the runtime"))
			} else {
				util.ExitWithErrorMessage("Failed to check available Composites", err)
			}
		} else {
			depJson = compositeInst.CompositeMetaData.Annotations.Dependencies
		}
	}

	dependencies, err := routing.ExtractDependencies(depJson)
	if err != nil {
		return err
	}
	if len(dependencies) == 0 {
		// No dependencies
		return fmt.Errorf("no dependencies found in cell instance %s", instanceName)
	}
	var tableData [][]string
	for _, dependency := range dependencies {
		record := []string{dependency["instance"], fmt.Sprintf("%s/%s", dependency["org"], dependency["name"]),
			dependency["version"]}
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

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

package setup

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/runtime"
)

var componentColor = color.New(color.FgWhite).Add(color.Bold).SprintFunc()
var componentLabelColor = color.New(color.FgHiGreen).Add(color.Bold).SprintFunc()

type SystemComponent struct {
	component runtime.SystemComponent
	enabled   bool
	status    string
}

func RunSetupStatus(cli cli.Cli) error {
	k8sServerVersion, _, err := cli.KubeCli().Version()
	if err != nil {
		if strings.Contains(k8sServerVersion, "Unable to connect to the server") {
			return fmt.Errorf("failed to get setup status, %v", fmt.Errorf(
				"unable to connect to the kubernetes cluster"))
		}
	}
	var clusterName string
	systemComponents := []*SystemComponent{{runtime.ApiManager, false, "Disabled"},
		{runtime.Observability, false, "Disabled"},
		{runtime.ScaleToZero, false, "Disabled"},
		{runtime.HPA, false, "Disabled"}}

	for _, systemComponent := range systemComponents {
		systemComponent.enabled, err = cli.Runtime().IsComponentEnabled(systemComponent.component)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("error checking if %s is enabled",
				systemComponent.component), err)
		}
		if systemComponent.enabled {
			systemComponent.status = "Enabled"
		}
	}
	clusterName, err = cli.KubeCli().GetContext()
	if err != nil {
		return fmt.Errorf("error getting cluster name, %v", err)
	}
	fmt.Fprintf(cli.Out(), componentLabelColor("cluster name: %s\n\n"), componentColor(clusterName))
	displayClusterComponentsTable(systemComponents)
	return nil
}

func displayClusterComponentsTable(systemComponents []*SystemComponent) {
	var tableData [][]string
	for _, systemComponent := range systemComponents {
		tableData = append(tableData, []string{string(systemComponent.component), systemComponent.status})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"SYSTEM COMPONENT", "STATUS"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{})

	table.AppendBulk(tableData)
	table.Render()
}

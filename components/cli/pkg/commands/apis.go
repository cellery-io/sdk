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
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/olekukonko/tablewriter"
)

func RunApis(cellName string) error {
	cmd := exec.Command("kubectl", "get", "gateways", cellName+"--gateway", "-o", "json")
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	output := ""
	go func() {
		for stdoutScanner.Scan() {
			output = output + stdoutScanner.Text()
		}
	}()

	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)

	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()
	err := cmd.Start()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while fetching APIs", err)
	}
	err = cmd.Wait()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while fetching APIs", err)
	}

	jsonOutput := &util.Gateway{}

	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson != nil {
		fmt.Println(errJson)
	}

	displayApisTable(jsonOutput.GatewaySpec.Apis, cellName)
	return nil
}

func displayApisTable(apiArray []util.GatewayApi, cellName string) error {
	tableData := [][]string{}

	for i := 0; i < len(apiArray); i++ {
		for j := 0; j < len(apiArray[i].Definitions); j++ {
			url := cellName + "/" + cellName + "-gateway-service" + "/" + apiArray[i].Context + "/" + strings.Split(apiArray[i].Backend, "/")[2]

			if apiArray[i].Definitions[j].Path != "/" {
				url = url + "/" + apiArray[i].Definitions[j].Path
			}
			tableRecord := []string{apiArray[i].Context, apiArray[i].Definitions[j].Method, url}
			tableData = append(tableData, tableRecord)
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"CONTEXT", "METHOD", "URL"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{})

	table.AppendBulk(tableData)
	table.Render()

	return nil
}

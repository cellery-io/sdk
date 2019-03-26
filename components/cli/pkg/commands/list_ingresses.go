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
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"os"
	"os/exec"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunListIngresses(cellName string) {
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

	displayApisTable(jsonOutput.GatewaySpec.HttpApis, cellName)
}

func displayApisTable(apiArray []util.GatewayHttpApi, cellName string) {
	var tableData [][]string

	for i := 0; i < len(apiArray); i++ {
		for j := 0; j < len(apiArray[i].Definitions); j++ {
			url := cellName + "--gateway-service"
			path := apiArray[i].Definitions[j].Path
			context := apiArray[i].Context
			method := apiArray[i].Definitions[j].Method

			// Add the context of the Cell
			if !strings.HasPrefix(context, "/") {
				url += "/"
			}
			url += context

			// Add the path of the API definition
			if path != "/" {
				if !strings.HasSuffix(url, "/") {
					if !strings.HasPrefix(path, "/") {
						url += "/"
					}
				} else {
					if strings.HasPrefix(path, "/") {
						url = strings.TrimSuffix(url, "/")
					}
				}
				url += path
			}

			// Add the global api url if globally exposed
			globalUrl := ""
			if apiArray[i].Global {
				if path != "/" {
					globalUrl = constants.WSO2_APIM_HOST + "/" + cellName + "/" + context + path
				} else {
					globalUrl = constants.WSO2_APIM_HOST + "/" + cellName + "/" + context
				}
			}

			tableRecord := []string{context, method, url, globalUrl}
			tableData = append(tableData, tableRecord)
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"CONTEXT", "METHOD", "LOCAL CELL GATEWAY", "GLOBAL API URL"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{})

	table.AppendBulk(tableData)
	table.Render()
}

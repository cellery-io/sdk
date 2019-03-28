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
	"regexp"
	"strconv"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"

	"github.com/olekukonko/tablewriter"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunListIngresses(name string) {
	instancePattern, _ := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), name)
	if instancePattern {
		displayCellInstanceApisTable(name)
	} else {
		displayCellImageApisTable(name)
	}
}

func displayCellInstanceApisTable(cellInstanceName string) {
	cmd := exec.Command("kubectl", "get", "gateways", cellInstanceName+"--gateway", "-o", "json")
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
	apiArray := jsonOutput.GatewaySpec.HttpApis

	var tableData [][]string

	for i := 0; i < len(apiArray); i++ {
		for j := 0; j < len(apiArray[i].Definitions); j++ {
			url := cellInstanceName + "--gateway-service"
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
					globalUrl = constants.WSO2_APIM_HOST + "/" + cellInstanceName + "/" + context + path
				} else {
					globalUrl = constants.WSO2_APIM_HOST + "/" + cellInstanceName + "/" + context
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

func displayCellImageApisTable(cellImageName string) {
	cellYamlContent := util.ReadCellImageYaml(cellImageName)
	cellImageContent := &util.Cell{}
	err := yaml.Unmarshal(cellYamlContent, cellImageContent)
	if err != nil {
		util.ExitWithErrorMessage("Error while reading cell image content", err)
	}

	var tableData [][]string

	for i := 0; i < len(cellImageContent.CellSpec.ComponentTemplates); i++ {
		componentName := cellImageContent.CellSpec.ComponentTemplates[i].Metadata.Name
		// Iterate HTTP and Web ingresses
		for j := 0; j < len(cellImageContent.CellSpec.GateWayTemplate.GatewaySpec.HttpApis); j++ {
			ingress := cellImageContent.CellSpec.GateWayTemplate.GatewaySpec.HttpApis[j]
			backend := ingress.Backend
			var ingressType = "web"
			if componentName == backend {
				var ingressData []string
				ingressData = append(ingressData, componentName)

				if len(ingress.Vhost) == 0 {
					ingressType = "HTTP"
				}
				ingressData = append(ingressData, ingressType)
				ingressData = append(ingressData, ingress.Context)
				ingressData = append(ingressData, strconv.Itoa(80))
				if ingress.Global {
					ingressData = append(ingressData, "True")
				} else {
					ingressData = append(ingressData, "False")
				}
				tableData = append(tableData, ingressData)
			}
		}
		// Iterate TCP ingresses
		for j := 0; j < len(cellImageContent.CellSpec.GateWayTemplate.GatewaySpec.TcpApis); j++ {
			ingress := cellImageContent.CellSpec.GateWayTemplate.GatewaySpec.TcpApis[j]
			backend := ingress.Backend
			if componentName == backend {
				var ingressData []string
				ingressData = append(ingressData, componentName)
				ingressData = append(ingressData, "tcp")
				ingressData = append(ingressData, "N/A")
				ingressData = append(ingressData, strconv.Itoa(8080))
				ingressData = append(ingressData, "False")
				tableData = append(tableData, ingressData)
			}
		}
		// Iterate GRPC ingresses
		for j := 0; j < len(cellImageContent.CellSpec.GateWayTemplate.GatewaySpec.TcpApis); j++ {
			ingress := cellImageContent.CellSpec.GateWayTemplate.GatewaySpec.TcpApis[j]
			backend := ingress.Backend
			if componentName == backend {
				var ingressData []string
				ingressData = append(ingressData, componentName)
				ingressData = append(ingressData, "grpc")
				ingressData = append(ingressData, "N/A")
				ingressData = append(ingressData, strconv.Itoa(3550))
				ingressData = append(ingressData, "False")
				tableData = append(tableData, ingressData)
			}
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"COMPONENT", "INGRESS TYPE", "INGRESS CONTEXT", "INGRESS PORT", "GLOBALLY EXPOSED"})
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

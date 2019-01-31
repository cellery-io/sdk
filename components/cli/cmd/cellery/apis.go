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

package cellery

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/wso2/cellery/components/cli/pkg/constants"
	"github.com/wso2/cellery/components/cli/pkg/util"
	"os"
	"os/exec"
	"strings"
)

func newApisCommand() *cobra.Command {
	var cellName string
	cmd := &cobra.Command{
		Use:   "apis [OPTIONS]",
		Short: "list the exposed APIs of a cell instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if (len(args) == 0) {
				cmd.Help()
				return nil
			}
			cellName = args[0]
			err := apis(cellName)
			if err != nil{
				cmd.Help()
				return err
			}
			return nil
		},
		Example: "  cellery apis my-project:v1.0 -n myproject-v1.0.0",
	}
	return cmd
}

func apis(cellName string) error {
	cmd := exec.Command("kubectl", "get", "gateways", cellName + "--gateway", "-o", "json")
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
		fmt.Printf("Error in executing cellery apis: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Cellery apis finished with error: \x1b[0m %v \n", err)
		os.Exit(1)
	}

	jsonOutput := &util.Gateway{}

	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson!= nil{
		fmt.Println(errJson)
	}

	displayApisTable(jsonOutput.GatewaySpec.Apis)
	return nil
}

func displayApisTable(apiArray []util.GatewayApi) error {
	tableData := [][]string{}

	for i := 0; i < len(apiArray); i++ {
		api := []string{strings.Split(apiArray[i].Backend, "/")[2], apiArray[i].Context}
		paths := getApiMethodsArray(apiArray[i].Definitions)

		for i := 0; i < len(paths) ; i++  {
			api = append(api, paths[i])
		}
		tableData = append(tableData, api)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"HOST NAME", "CONTEXT", "GET", "POST", "PATCH", "PUT", "DELETE"})
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
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(
		tablewriter.Colors{},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{})

	table.AppendBulk(tableData)
	table.Render()

	return nil
}

func getApiMethodsArray(definitions []util.GatewayDefinition) []string {
	methodArray := make([]string, 5)

	for i := 0; i < len(definitions) ; i++ {
		if strings.EqualFold(definitions[i].Method, constants.HTTP_METHOD_GET) {
			methodArray[0] = definitions[i].Path
		}
		if strings.EqualFold(definitions[i].Method, constants.HTTP_METHOD_POST)  {
			methodArray[1] = definitions[i].Path
		}
		if strings.EqualFold(definitions[i].Method, constants.HTTP_METHOD_PATCH)  {
			methodArray[2] = definitions[i].Path
		}
		if strings.EqualFold(definitions[i].Method, constants.HTTP_METHOD_PUT)  {
			methodArray[3] = definitions[i].Path
		}
		if strings.EqualFold(definitions[i].Method, constants.HTTP_METHOD_DELETE)  {
			methodArray[4] = definitions[i].Path
		}
	}

	return methodArray
}

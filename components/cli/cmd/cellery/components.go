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
	"strconv"
	"strings"
)

func newComponentsCommand() *cobra.Command {
	var cellName string
	cmd := &cobra.Command{
		Use:   "components [OPTIONS]",
		Short: "Lists the components which the cell encapsulates.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if (len(args) == 0) {
				cmd.Help()
				return nil
			}
			cellName = args[0]
			err := components(cellName)
			if err != nil{
				cmd.Help()
				return err
			}
			return nil
		},
		Example: "  cellery components my-project:v1.0 -n myproject-v1.0.0",
	}
	return cmd
}

func components(cellName string) error {
	cmd := exec.Command("kubectl", "get", "services", "-l", constants.GROUP_NAME + "/cell=" + cellName, "-o", "json")
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	output := constants.EMPTY_STRING
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
		fmt.Printf("Error in executing cell components: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Cell components finished with error: \x1b[0m %v \n", err)
		os.Exit(1)
	}

	jsonOutput := &util.Service{}

	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson!= nil{
		fmt.Println(errJson)
	}

	if len(jsonOutput.Items) == 0 {
		fmt.Printf("Cannot find cell: %v \n", cellName)
	} else {
		displayComponentsTable(jsonOutput.Items, cellName)
	}
	return nil
}

func displayComponentsTable(componentArray []util.ServiceItem, cellName string) error {
	tableData := [][]string{}

	for i := 0; i < len(componentArray); i++ {
		var name string
		name = strings.Replace(componentArray[i].Metadata.Name, cellName + "--", "", -1)
		name = strings.Replace(name, "-service", "", -1)

		ports := strconv.Itoa(componentArray[i].Spec.Ports[0].Port)
		for j := 1; j < len(componentArray[i].Spec.Ports); j++ {
			ports = ports + "/" + strconv.Itoa(componentArray[i].Spec.Ports[j].Port)
		}
		component := []string{name, ports}
		tableData = append(tableData, component)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "PORTS"})
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

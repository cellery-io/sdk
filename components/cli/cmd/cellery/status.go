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
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func newStatusCommand() *cobra.Command {
	var cellName string
	cmd := &cobra.Command{
		Use:   "status [OPTIONS]",
		Short: "Performs a health check of a cell.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if (len(args) == 0) {
				cmd.Help()
				return nil
			}
			cellName = args[0]
			err := status(cellName)
			if err != nil{
				cmd.Help()
				return err
			}
			return nil
		},
		Example: "  cellery status my-project:v1.0 -n myproject-v1.0.0",
	}
	return cmd
}

func status(cellName string) error {
	cmd := exec.Command("kubectl", "get", "pods", "-l", constants.GROUP_NAME+ "/cell=" + cellName, "-o", "json")
	output := ""

	outfile, errPrint := os.Create("./out.txt")
	if errPrint != nil {
		fmt.Printf("Error in executing cell status: %v \n", errPrint)
		os.Exit(1)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error in executing cell status: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Cell status finished with error: \x1b[0m %v \n", err)
		os.Exit(1)
	}

	outputByteArray, err := ioutil.ReadFile("./out.txt")
	os.Remove("./out.txt")

	if err != nil {
		fmt.Printf("Error in executing cell status: %v \n", err)
		os.Exit(1)
	}

	output = string(outputByteArray)
	jsonOutput := &util.CellPods{}
	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson!= nil{
		fmt.Println(errJson)
	}

	if len(jsonOutput.Items) == 0 {
		fmt.Printf("Cannot find cell: %v \n", cellName)
	} else {
		cellCreationTime, cellStatus := getCellSummary(cellName)
		displayStatusSummaryTable(cellCreationTime, cellStatus)
		fmt.Println("\n")
		displayStatusDetailedTable(jsonOutput.Items, cellName)
	}
	return nil
}

func getCellSummary(cellName string) (cellCreationTime, cellStatus string) {
	cellCreationTime = ""
	cellStatus = ""
	cmd := exec.Command("kubectl", "get", "cells", cellName, "-o", "json")
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
		fmt.Printf("Error in executing cell status: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Cell status finished with error: \x1b[0m %v \n", err)
		os.Exit(1)
	}

	jsonOutput := &util.Cell{}

	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson!= nil{
		fmt.Println(errJson)
	}
	duration := util.GetDuration(util.ConvertStringToTime(jsonOutput.CellMetaData.CreationTimestamp))
	cellStatus = jsonOutput.CellStatus.Status

	return duration, cellStatus
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

func displayStatusDetailedTable(podItems []util.Pod, cellName string) error {
	tableData := [][]string{}

	for i := 0; i < len(podItems); i++ {
		name := strings.Replace(strings.Split(podItems[i].MetaData.Name, "-deployment-")[0], cellName + "--", "", -1)
		state := podItems[i].PodStatus.Phase

		if strings.EqualFold(state, "Running") {
			duration := util.GetDuration(util.ConvertStringToTime(podItems[i].PodStatus.Conditions[1].LastTransitionTime))
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

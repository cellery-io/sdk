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
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunStatus(cellName string) {
	cellCreationTime, cellStatus, err := getCellSummary(cellName)
	if err != nil {
		util.ExitWithErrorMessage("Error running cell status", fmt.Errorf("cannot find cell %v \n", cellName))
	}
	displayStatusSummaryTable(cellCreationTime, cellStatus)
	fmt.Println("\n")
	fmt.Println("  -COMPONENTS-\n")
	displayStatusDetailedTable(getPodDetails(cellName), cellName)
}

func getCellSummary(cellName string) (cellCreationTime, cellStatus string, err error) {
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
			output = output + stderrScanner.Text()
		}
	}()
	err = cmd.Start()
	if err != nil {
		return "", "", err
	}
	err = cmd.Wait()
	if err != nil {
		return "", "", err
	}

	jsonOutput := &util.Cell{}

	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson != nil {
		fmt.Println(errJson)
	}
	duration := util.GetDuration(util.ConvertStringToTime(jsonOutput.CellMetaData.CreationTimestamp))
	cellStatus = jsonOutput.CellStatus.Status

	return duration, cellStatus, err
}

func getPodDetails(cellName string) []util.Pod {
	cmd := exec.Command("kubectl", "get", "pods", "-l", constants.GROUP_NAME+"/cell="+cellName, "-o", "json")
	output := ""

	outfile, errPrint := os.Create("./out.txt")
	if errPrint != nil {
		util.ExitWithErrorMessage("Error occurred while fetching cell status", errPrint)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	err := cmd.Start()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while fetching cell status", err)
	}
	err = cmd.Wait()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while fetching cell status", err)
	}

	outputByteArray, err := ioutil.ReadFile("./out.txt")
	os.Remove("./out.txt")

	if err != nil {
		util.ExitWithErrorMessage("Error occurred while fetching cell status", err)
	}

	output = string(outputByteArray)
	jsonOutput := &util.CellPods{}
	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson != nil {
		fmt.Println(errJson)
	}
	return jsonOutput.Items
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
		name := strings.Replace(strings.Split(podItems[i].MetaData.Name, "-deployment-")[0], cellName+"--", "", -1)
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

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
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wso2/cellery/components/cli/pkg/constants"
	"os"
	"os/exec"
)

func newLogsCommand() *cobra.Command {
	var cellName, componentName string
	cmd := &cobra.Command{
		Use:   "logs [OPTIONS]",
		Short: "Displays logs for either the cell instance, or a component of a running cell instance.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if (len(args) == 0) {
				cmd.Help()
				return nil
			}
			cellName = args[0]
			if len(args) > 1 {
				componentName = args[1]
				err := componentLogs(cellName, componentName)
				if err != nil{
					cmd.Help()
					return err
				}
			} else {
				err := cellLogs(cellName)
				if err != nil{
					cmd.Help()
					return err
				}
			}
			return nil
		},
		Example: "  cellery logs my-cell:v1.0  cell-component-v1.0.0",
	}
	return cmd
}

func componentLogs(cellName, componentName string) error {
	cmd := exec.Command("kubectl", "logs", "-l", constants.GROUP_NAME + "/service=" + cellName + "--" + componentName, "-c", componentName)
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	output := ""
	go func() {
		for stdoutScanner.Scan() {
			output += stdoutScanner.Text()
			fmt.Println(stdoutScanner.Text())
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
		fmt.Printf("Error in executing cellery logs: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Cellery logs finished with error: \x1b[0m %v \n", err)
		os.Exit(1)
	}
	if output == "" {
		fmt.Printf("Cannot find cell: %v \n", cellName)
	}
	return nil
}

func cellLogs(cellName string) error {
	cmd := exec.Command("kubectl", "logs", "-l", constants.GROUP_NAME + "/cell=" + cellName, "--all-containers=true")
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	output := ""
	go func() {
		for stdoutScanner.Scan() {
			output += stdoutScanner.Text()
			fmt.Println(stdoutScanner.Text())
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
		fmt.Printf("Error in executing cell ps: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Cell ps finished with error: \x1b[0m %v \n", err)
		os.Exit(1)
	}
	if output == "" {
		fmt.Printf("Cannot find cell: %v \n", cellName)
	}
	return nil
}

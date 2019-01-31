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

package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

func newPsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ps [OPTIONS]",
		Short: "list all running cells",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ps()
			if err != nil{
				cmd.Help()
				return err
			}
			return nil
		},
		Example: "  cellery ps",
	}
	return cmd
}

func ps() error {
	cmd := exec.Command("kubectl", "get", "cells")
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Println(stdoutScanner.Text())
		}
	}()
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
			if (stderrScanner.Text() == "No resources found.") {
				os.Exit(0)
			}
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
	return nil
}

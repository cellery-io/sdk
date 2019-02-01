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
	"github.com/celleryio/sdk/components/cli/pkg/commands"
	"github.com/spf13/cobra"
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
				err := commands.RunComponentLogs(cellName, componentName)
				if err != nil{
					cmd.Help()
					return err
				}
			} else {
				err := commands.RunCellLogs(cellName)
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


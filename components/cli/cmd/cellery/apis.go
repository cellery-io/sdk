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
	"github.com/spf13/cobra"
	"github.com/celleryio/sdk/components/cli/pkg/internal"
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
			err := internal.RunApis(cellName)
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

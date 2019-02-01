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

	"github.com/celleryio/sdk/components/cli/pkg/commands"
)

type Component struct {
	name, dockerImage   string
	ports               []int
	deployment, service string
}

func newImageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images [OPTIONS]",
		Short: "List cell images",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return commands.RunImage()
			} else {
				return commands.RunImageInformations(args[0])
			}
		},
	}
	return cmd
}

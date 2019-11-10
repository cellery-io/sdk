/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/commands/image"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// newViewCommand creates a new command which can be executed to view a particular image
func newViewCommand(cli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <organization>/<cell-image>:<version>",
		Short: "View the Cell Image in a browser",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := image.RunView(cli, args[0]); err != nil {
				util.ExitWithErrorMessage("Cellery view command failed", err)
			}
		},
		Example: "  cellery view cellery-samples/employee:1.0.0",
	}
	return cmd
}

/*
 * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/commands"
)

func newDeleteImageCommand(cli cli.Cli) *cobra.Command {
	var deleteAll = false
	var regex = ""
	cmd := &cobra.Command{
		Use:   "delete <cell-image(s)>",
		Short: "Delete cell image(s) from repo",
		Args: func(cmd *cobra.Command, args []string) error {
			if !deleteAll && regex == "" {
				err := cobra.MinimumNArgs(1)(cmd, args)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunDeleteImage(cli, args, regex, deleteAll)
		},
		Example: "  cellery delete cellery-samples/employee:1.0.0  my-org/hr:1.0.0\n" +
			"  cellery delete cellery-samples/employee:1.0.0 --regex '.*/employee:.*'\n" +
			"  cellery delete --all\n" +
			"  cellery delete --regex .*/employee:.*\n",
	}
	cmd.Flags().BoolVar(&deleteAll, "all", false, "Delete all cell images")
	cmd.Flags().StringVar(&regex, "regex", "", "Regular expression of cell images to be deleted")
	return cmd
}

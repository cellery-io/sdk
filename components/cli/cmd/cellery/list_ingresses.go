/*
 * Copyright (c) 2018 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/commands/image"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

// newApisCommand creates a cobra command which can be invoked to get the APIs exposed by a cell
func newListIngressesCommand(cli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ingresses <instance-name|cell-image-name>",
		Aliases: []string{"ingress", "ing"},
		Short:   "List the exposed APIs of a cell instance",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			isCellValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), args[0])
			if err != nil || !isCellValid {
				isCellImageValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CellImagePattern), args[0])
				if err != nil || !isCellImageValid {
					return fmt.Errorf("expects a valid cell instance name or a cell image name, received %s", args[0])
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			image.RunListIngresses(cli, args[0])
		},
		Example: "  cellery list ingresses employee\n" +
			"  cellery list ingresses cellery-samples/employee:1.0.0\n",
	}
	return cmd
}

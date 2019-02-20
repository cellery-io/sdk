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
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/commands"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func newRunCommand() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "run [<registry>/]<organization>/<cell-image>:<version>",
		Short: "Use a cell image to create a running instance",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			err = util.ValidateImageTagWithRegistry(args[0])
			if err != nil {
				return err
			}
			if name != "" {
				isCellValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), name)
				if err != nil || !isCellValid {
					return fmt.Errorf("expects a valid cell name, received %s", args[0])
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunRun(args[0], name)
		},
		Example: "  cellery run cellery-samples/employee:1.0.0 -n employee\n" +
			"  cellery run registry.foo.io/cellery-samples/employee:1.0.0 -n employee" +
			"  cellery run cellery-samples/employee:1.0.0",
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the cell instance")
	return cmd
}

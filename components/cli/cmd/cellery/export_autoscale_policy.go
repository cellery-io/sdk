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
	"fmt"
	"log"
	"regexp"

	"github.com/cellery-io/sdk/components/cli/pkg/util"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/commands"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

func newExportAutoscalePolicies() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "autoscale <cell_instance_name>",
		Short: "Export autocale policies for a cell instance",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.MinimumNArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			isCellInstValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), args[0])
			if err != nil {
				log.Fatal(err)
			}
			if !isCellInstValid {
				return fmt.Errorf("expects a valid cell instance name, received %s", args[0])
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := commands.RunExportAutoscalePoliciesOfCell(args[0], file)
			if err != nil {
				util.ExitWithErrorMessage(fmt.Sprintf("Unable to export autoscale policies from instance %s", args[0]), err)
			}
		},
		Example: "  cellery export-policy autoscale mytestcell1 -f myscalepolicy.yaml",
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "output file for autoscale policy")
	return cmd
}

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

	"github.com/cellery-io/sdk/components/cli/pkg/commands"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func newApplyAutoscalePolicyCommand() *cobra.Command {
	var components string
	cmd := &cobra.Command{
		Use:   "autoscale <file> <instance>",
		Short: "apply autoscale policies for a cell instance",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.MinimumNArgs(2)(cmd, args)
			if err != nil {
				return err
			}
			isCellInstValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), args[1])
			if err != nil {
				log.Fatal(err)
			}
			if !isCellInstValid {
				return fmt.Errorf("expects a valid cell instance name, received %s", args[1])
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if components == "" {
				err = commands.RunApplyAutoscalePolicies(args[0], args[1])
			} else {
				err = commands.RunApplyAutoscalePoliciesToComponents(args[0], args[1], components)
			}
			if err != nil {
				util.ExitWithErrorMessage(fmt.Sprintf("Unable to apply autoscale policies to instance %s", args[1]), err)
			}
		},
		Example: "  cellery apply-policy autoscale myscalepolicy.yaml myinstance --components comp1,comp2",
	}
	cmd.Flags().StringVarP(&components, "components", "c", "", "comma separated components list")
	return cmd
}

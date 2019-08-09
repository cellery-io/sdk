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
	"regexp"
	"strconv"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/util"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/commands"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

func newRouteTrafficCommand() *cobra.Command {
	var sourceInstance string
	var dependencyInstance string
	var targetPercentage string
	var targetInstance string
	var percentage int
	var srcInstances []string
	var enableSessionAwareness bool
	cmd := &cobra.Command{
		Use:   "route-traffic [--source|-s=<list_of_source_cell_instances>] --dependency|-d <dependency_instance_name> --target|-t <target instance name> [--percentage|-p <x>]",
		Short: "route a percentage of the traffic to a cell instance",
		Example: "cellery route-traffic --source hr-client-inst1 --dependency hr-inst-1 --target hr-inst-2 --percentage 20 \n" +
			"cellery route-traffic --dependency hr-inst-1 --target hr-inst-2 --percentage 20 \n" +
			"cellery route-traffic --dependency hr-inst-1 --target hr-inst-2 \n" +
			"cellery route-traffic --dependency hr-inst-1 --target hr-inst-2 --percentage 25 --enable-session-awareness",
		Args: func(cmd *cobra.Command, args []string) error {
			// validate
			err := validateArguments(dependencyInstance, targetInstance)
			if err != nil {
				return err
			}
			// get source cell instances as an array
			srcInstances = getSourceCellInstanceArr(sourceInstance)
			// validate source cell instance name(s)
			for _, srcInstance := range srcInstances {
				err := validateInstanceName(srcInstance)
				if err != nil {
					util.ExitWithErrorMessage("Error in running route traffic command", err)
				}
			}
			// validate target instance name
			err = validateInstanceName(targetInstance)
			if err != nil {
				util.ExitWithErrorMessage("Error in running route traffic command", err)
			}
			// calculate target percentage value
			percentage, err = getTargetInstancePercentage(targetPercentage)
			if err != nil {
				util.ExitWithErrorMessage("Error in running route traffic command", err)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := commands.RunRouteTrafficCommand(srcInstances, dependencyInstance, targetInstance, percentage, enableSessionAwareness)
			if err != nil {
				util.ExitWithErrorMessage(fmt.Sprintf("Unable to route traffic to the target instance: %s, percentage: %d", targetInstance, percentage), err)
			}
		},
	}
	cmd.Flags().StringVarP(&sourceInstance, "source", "s", "", "comma separated source instance list")
	cmd.Flags().StringVarP(&dependencyInstance, "dependency", "d", "", "existing dependency instance name")
	cmd.Flags().StringVarP(&targetInstance, "target", "t", "", "target instance to which the traffic should be re-routed")
	cmd.Flags().StringVarP(&targetPercentage, "percentage", "p", "", "percentage to be switched to the target instance")
	cmd.Flags().BoolVarP(&enableSessionAwareness, "enable-session-awareness", "a", false, "flag to enable session awareness based on user name")
	return cmd
}

func getSourceCellInstanceArr(sourceCellInstances string) []string {
	var trimmedInstances []string
	if len(sourceCellInstances) == 0 {
		return trimmedInstances
	}
	instances := strings.Split(sourceCellInstances, ",")
	for _, instance := range instances {
		trimmedInstances = append(trimmedInstances, strings.TrimSpace(instance))
	}
	return trimmedInstances
}

func getTargetInstancePercentage(percentage string) (int, error) {
	// if the percentage is not given, assume its 100 (full traffic switch - canary)
	if percentage == "" {
		return 100, nil
	}
	intPercentage, err := strconv.Atoi(percentage)
	if err != nil {
		return -1, err
	}
	if intPercentage > 100 {
		return -1, fmt.Errorf("invalid target percentage value %d", intPercentage)
	}
	return intPercentage, nil
}

func validateInstanceName(instanceName string) error {
	isCellInstValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), instanceName)
	if err != nil {
		return err
	}
	if !isCellInstValid {
		return fmt.Errorf("expects a valid cell instance name, received '%s'", instanceName)
	}
	return nil
}

func validateArguments(dependency string, target string) error {
	if dependency == "" {
		return fmt.Errorf("mandatory flag dependency/d not provided")
	}
	if target == "" {
		return fmt.Errorf("mandatory flag target/t not provided")
	}
	return nil
}

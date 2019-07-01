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
	var source string
	var targetPercentage string
	var targetInstance string
	var percentage int
	var srcInstances []string
	cmd := &cobra.Command{
		Use:   "route-traffic [--source|-s=<list_of_source_cell_instances>] <dependency_instance_name> --percentage|-p <target_cell_instance>=<x>",
		Short: "route a percentage of the traffic to a cell instance",
		Example: "cellery route-traffic --source hr-client-inst1 hr-inst-1 --percentage hr-inst-2=20 \n" +
			"cellery route-traffic hr-inst-1 --percentage hr-inst-2=20",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.MinimumNArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			// validate dependency cell instance name
			isCellInstValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), args[0])
			if err != nil {
				util.ExitWithErrorMessage("Error in running route traffic command", err)
			}
			if !isCellInstValid {
				util.ExitWithErrorMessage("Error in running route traffic command", fmt.Errorf("expects a valid cell instance name, received %s", args[0]))
			}
			//validate source cell instances
			srcInstances = getSourceCellInstanceArr(source)
			for _, srcInstance := range srcInstances {
				isCellInstValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), srcInstance)
				if err != nil {
					util.ExitWithErrorMessage("Error in running route traffic command", err)
				}
				if !isCellInstValid {
					util.ExitWithErrorMessage("Error in running route traffic command", fmt.Errorf("expects a valid source cell instance name, received %s", srcInstance))
				}
			}
			targetInstance, percentage, err = getTargetCelInstanceAndPercentage(targetPercentage)
			if err != nil {
				util.ExitWithErrorMessage("Error in running route traffic command", err)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := commands.RunRouteTrafficCommand(srcInstances, args[0], targetInstance, percentage)
			if err != nil {
				util.ExitWithErrorMessage(fmt.Sprintf("Unable to route traffic to the target instance: %s, percentage: %s", targetInstance, percentage), err)
			}
		},
	}
	cmd.Flags().StringVarP(&source, "source", "s", "", "comma separated source instance list")
	cmd.Flags().StringVarP(&targetPercentage, "percentage", "p", "", "target instance and percentage of traffic joined by a '=' mark")
	_ = cmd.MarkFlagRequired("percentage")
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

func getTargetCelInstanceAndPercentage(target string) (string, int, error) {
	parts := strings.Split(target, "=")
	if len(parts) != 2 {
		return "", -1, fmt.Errorf("target instance and percentage in incorrect format %s", target)
	}
	percentage, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", -1, err
	}
	if percentage > 100 {
		return "", -1, fmt.Errorf("invalid percentge provided for target instance: %d", percentage)
	}
	// validate target cell instance name
	isCellInstValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), parts[0])
	if err != nil {
		return "", -1, err
	}
	if !isCellInstValid {
		return "", -1, fmt.Errorf("expects a valid cell instance name, received %s", parts[0])
	}
	return parts[0], percentage, nil
}

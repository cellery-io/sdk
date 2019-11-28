/*
 * Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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
	"log"
	"regexp"

	"github.com/spf13/cobra"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/commands/instance"
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/util"
)

func newExportAutoscalePolicies(cli cli.Cli) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "autoscale <command>",
		Short: "Export autocale policies for a cell/composite instance",
	}
	cmd.PersistentFlags().StringVarP(&file, "file", "f", "", "output file for autoscale policy")
	cmd.AddCommand(
		newExportCellAutoscalePolicies(cli, &file),
		newExportCompositeAutoscalePolicies(cli, &file),
	)
	return cmd
}

func newExportCellAutoscalePolicies(cli cli.Cli, file *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cell <cell_instance_name>",
		Short: "Export autocale policies for a cell instance",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.MinimumNArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			valid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), args[0])
			if err != nil {
				log.Fatal(err)
			}
			if !valid {
				return fmt.Errorf("expects a valid cell instance name, received %s", args[0])
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := instance.RunExportAutoscalePolicies(cli, kubernetes.InstanceKindCell, args[0], *file)
			if err != nil {
				util.ExitWithErrorMessage(fmt.Sprintf("Unable to export autoscale policies from instance %s", args[0]), err)
			}
		},
		Example: "  cellery export-policy autoscale cell mytestcell1 -f myscalepolicy.yaml",
	}
	return cmd
}

func newExportCompositeAutoscalePolicies(cli cli.Cli, file *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "composite <cell_instance_name>",
		Short: "Export autocale policies for a composite instance",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.MinimumNArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			valid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), args[0])
			if err != nil {
				log.Fatal(err)
			}
			if !valid {
				return fmt.Errorf("expects a valid composite instance name, received %s", args[0])
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := instance.RunExportAutoscalePolicies(cli, kubernetes.InstanceKindComposite, args[0], *file)
			if err != nil {
				util.ExitWithErrorMessage(fmt.Sprintf("Unable to export autoscale policies from instance %s", args[0]), err)
			}
		},
		Example: "  cellery export-policy autoscale composite mytestcell1 -f myscalepolicy.yaml",
	}
	return cmd
}

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
	"strings"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/commands"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

func newSetupCleanupGcpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcp",
		Short: "Cleanup gcp setup",
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.ExactArgs(1)(cmd, args)
			if err != nil {
				return err
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			valid, err := commands.ValidateGcpCluster(args[0])
			if !valid || err != nil {
				return fmt.Errorf("Gcp cluster " + args[0] + " doesn't exist")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			uniqueNumber := strings.Split(args[0], constants.GCP_CLUSTER_NAME)[1]
			commands.RunCleanupGcp(constants.GCP_CLUSTER_NAME + uniqueNumber)
		},
	}
	return cmd
}

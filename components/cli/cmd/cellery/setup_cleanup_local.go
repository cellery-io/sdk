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

	"github.com/spf13/cobra"

	"cellery.io/cellery/components/cli/pkg/commands/setup"
	"cellery.io/cellery/components/cli/pkg/vbox"
)

func newSetupCleanupLocalCommand() *cobra.Command {
	var deleteImage = false
	cmd := &cobra.Command{
		Use:   "local",
		Short: "Cleanup local setup",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !vbox.IsVmInstalled() {
				return fmt.Errorf("VM is not installed")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			setup.RunCleanupLocal(true, deleteImage)
		},
		Example: "  cellery setup cleanup local",
	}
	cmd.Flags().BoolVarP(&deleteImage, "delete-image", "d", false, "Delete downloaded image of vm")
	return cmd
}

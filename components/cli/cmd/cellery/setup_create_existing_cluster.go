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

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/commands"
	"github.com/cellery-io/sdk/components/cli/pkg/runtime"
)

func newSetupCreateOnExistingClusterCommand() *cobra.Command {
	var isCompleteSetup = false
	var isPersistentVolume = false
	var hasNfsStorage = false
	var isLoadBalancerIngressMode = false
	var nfs runtime.Nfs
	var nfsServerIp = ""
	var fileShare = ""
	var dbHostName = ""
	var dbUserName = ""
	var dbPassword = ""
	cmd := &cobra.Command{
		Use:   "existing",
		Short: "Create a Cellery runtime in existing cluster",
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if isPersistentVolume && hasNfsStorage {
				nfs = runtime.Nfs{nfsServerIp, fileShare, dbHostName,
					dbUserName, dbPassword}
			}
			return validateUserInputForExistingCluster(hasNfsStorage, nfsServerIp, fileShare, dbHostName, dbUserName,
				dbPassword)
		},
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunSetupCreateOnExistingCluster(isCompleteSetup, isPersistentVolume, hasNfsStorage,
				isLoadBalancerIngressMode, nfs)
		},
		Example: "  cellery setup create existing",
	}
	cmd.Flags().BoolVarP(&isCompleteSetup, "complete", "c", false, "Create complete setup")
	cmd.Flags().BoolVarP(&isPersistentVolume, "persistent", "p", false, "Persistent volume")
	cmd.Flags().BoolVarP(&hasNfsStorage, "nfs", "n", false, "Has an NFS storage")
	cmd.Flags().BoolVarP(&isLoadBalancerIngressMode, "loadbalancer", "l", false,
		"Ingress mode is load balancer")
	cmd.Flags().StringVarP(&nfsServerIp, "ip", "i", "", "NFS Server Ip")
	cmd.Flags().StringVarP(&fileShare, "fileshare", "f", "", "NFS file share")
	cmd.Flags().StringVarP(&dbHostName, "host", "d", "", "Database host")
	cmd.Flags().StringVarP(&dbUserName, "username", "u", "", "Database user name")
	cmd.Flags().StringVarP(&dbPassword, "password", "w", "", "Database password")
	return cmd
}

func validateUserInputForExistingCluster(hasNfsStorage bool, nfsServerIp, fileShare, dbHostName, dbUserName,
	dbPassword string) error {
	var valid = true
	var errMsg string
	if hasNfsStorage {
		errMsg = "Missing input:"
		if nfsServerIp == "" {
			errMsg += " ip,"
			valid = false
		}
		if fileShare == "" {
			errMsg += " fileshare,"
			valid = false
		}
		if dbHostName == "" {
			errMsg += " host,"
			valid = false
		}
		if dbUserName == "" {
			errMsg += " username,"
			valid = false
		}
		if dbPassword == "" {
			errMsg += " password,"
			valid = false
		}
	} else {
		errMsg = "Unexpected input:"
		if nfsServerIp != "" {
			errMsg += " --ip " + nfsServerIp + ","
			valid = false
		}
		if fileShare != "" {
			errMsg += " --fileshare " + fileShare + ","
			valid = false
		}
		if dbHostName != "" {
			errMsg += " --host " + dbHostName + ","
			valid = false
		}
		if dbUserName != "" {
			errMsg += " --username " + dbUserName + ","
			valid = false
		}
		if dbPassword != "" {
			errMsg += " --password " + dbPassword + ","
			valid = false
		}
	}
	if !valid {
		return fmt.Errorf(errMsg)
	}
	return nil
}

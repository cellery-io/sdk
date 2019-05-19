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
	var db runtime.MysqlDb
	var nfsServerIp = ""
	var fileShare = ""
	var dbHostName = ""
	var dbUserName = ""
	var dbPassword = ""
	cmd := &cobra.Command{
		Use:   "existing",
		Short: "Create a Cellery runtime in existing cluster",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			hasNfsStorage = useNfsStorage(nfsServerIp, fileShare, dbHostName, dbUserName, dbPassword)
			if hasNfsStorage {
				nfs = runtime.Nfs{nfsServerIp, "/" + fileShare}
				db = runtime.MysqlDb{dbHostName, dbUserName, dbPassword}
			}
			return validateUserInputForExistingCluster(hasNfsStorage, nfsServerIp, fileShare, dbHostName, dbUserName,
				dbPassword)
		},
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunSetupCreateOnExistingCluster(isCompleteSetup, isPersistentVolume, hasNfsStorage,
				isLoadBalancerIngressMode, nfs, db)
		},
		Example: "  cellery setup create existing",
	}
	cmd.Flags().BoolVar(&isCompleteSetup, "complete", false, "Create complete setup")
	cmd.Flags().BoolVar(&isPersistentVolume, "persistent", false, "Persistent volume")
	cmd.Flags().BoolVar(&isLoadBalancerIngressMode, "loadbalancer", false,
		"Ingress mode is load balancer")
	cmd.Flags().StringVar(&nfsServerIp, "nfsServerIp", "", "NFS Server Ip")
	cmd.Flags().StringVar(&fileShare, "nfsFileshare", "", "NFS file share")
	cmd.Flags().StringVar(&dbHostName, "dbHost", "", "Database host")
	cmd.Flags().StringVar(&dbUserName, "dbUsername", "", "Database user name")
	cmd.Flags().StringVar(&dbPassword, "dbPassword", "", "Database password")
	return cmd
}

func validateUserInputForExistingCluster(hasNfsStorage bool, nfsServerIp, fileShare, dbHostName, dbUserName,
	dbPassword string) error {
	var valid = true
	var errMsg string
	if hasNfsStorage {
		errMsg = "Missing input:"
		if nfsServerIp == "" {
			errMsg += " nfsServerIp,"
			valid = false
		}
		if fileShare == "" {
			errMsg += " nfsFileshare,"
			valid = false
		}
		if dbHostName == "" {
			errMsg += " dbHost,"
			valid = false
		}
		if dbUserName == "" {
			errMsg += " dbUsername,"
			valid = false
		}
		if dbPassword == "" {
			errMsg += " dbPassword,"
			valid = false
		}
	} else {
		errMsg = "Unexpected input:"
		if nfsServerIp != "" {
			errMsg += " --nfsServerIp " + nfsServerIp + ","
			valid = false
		}
		if fileShare != "" {
			errMsg += " --nfsFileshare " + fileShare + ","
			valid = false
		}
		if dbHostName != "" {
			errMsg += " --dbHost " + dbHostName + ","
			valid = false
		}
		if dbUserName != "" {
			errMsg += " --dbUsername " + dbUserName + ","
			valid = false
		}
		if dbPassword != "" {
			errMsg += " --dbPassword " + dbPassword + ","
			valid = false
		}
	}
	if !valid {
		return fmt.Errorf(errMsg)
	}
	return nil
}

func useNfsStorage(nfsServerIp, fileShare, dbHostName, dbUserName, dbPassword string) bool {
	if nfsServerIp != "" || fileShare != "" || dbHostName != "" || dbUserName != "" || dbPassword != "" {
		return true
	}
	return false
}

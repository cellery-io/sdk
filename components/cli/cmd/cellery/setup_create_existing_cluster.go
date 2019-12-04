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
	"regexp"

	"github.com/spf13/cobra"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/commands/setup"
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
)

func newSetupCreateOnExistingClusterCommand(cli cli.Cli, isComplete bool) *cobra.Command {
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
	var nodePortIpAddress = ""
	cmd := &cobra.Command{
		Use:   "existing",
		Short: "Create a Cellery runtime in existing cluster",
		Args: func(cmd *cobra.Command, args []string) error {
			if nodePortIpAddress != "" {
				isNodePortIpAddressValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.IpAddressPattern), nodePortIpAddress)
				if err != nil || !isNodePortIpAddressValid {
					return fmt.Errorf("expects a valid nodeport ip address, received %s", nodePortIpAddress)
				}
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			hasNfsStorage = useNfsStorage(nfsServerIp, fileShare, dbHostName, dbUserName, dbPassword)
			if hasNfsStorage {
				nfs = runtime.Nfs{NfsServerIp: nfsServerIp, FileShare: "/" + fileShare}
				db = runtime.MysqlDb{DbHostName: dbHostName, DbUserName: dbUserName, DbPassword: dbPassword}
			}
			return validateUserInputForExistingCluster(hasNfsStorage, nfsServerIp, fileShare, dbHostName, dbUserName,
				dbPassword)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := setup.RunSetupCreate(cli, nil, isComplete, isPersistentVolume, hasNfsStorage, isLoadBalancerIngressMode,
				nfs, db, nodePortIpAddress); err != nil {
				util.ExitWithErrorMessage("Cellery setup create existing command failed", err)
			}
		},
		Example: "  cellery setup create existing",
	}
	cmd.Flags().BoolVar(&isPersistentVolume, "persistent", false, "Persistent volume")
	cmd.Flags().BoolVar(&isLoadBalancerIngressMode, "loadbalancer", false,
		"Ingress mode is load balancer")
	cmd.Flags().StringVar(&nfsServerIp, "nfsServerIp", "", "NFS Server Ip")
	cmd.Flags().StringVar(&fileShare, "nfsFileshare", "", "NFS file share")
	cmd.Flags().StringVar(&dbHostName, "dbHost", "", "Database host")
	cmd.Flags().StringVar(&dbUserName, "dbUsername", "", "Database user name")
	cmd.Flags().StringVar(&dbPassword, "dbPassword", "", "Database password")
	cmd.Flags().StringVar(&nodePortIpAddress, "nodePortIp", "", "NodePort Ip Address")
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

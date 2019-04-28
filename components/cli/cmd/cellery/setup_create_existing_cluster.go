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
	"github.com/spf13/cobra"
)

func newSetupCreateOnExistingClusterCommand() *cobra.Command {
	var isCompleteSetup = false
	var isPersistentVolume = false
	var hasNfsStorage = false
	var isLoadBalancerIngressMode = false
	cmd := &cobra.Command{
		Use:   "existing",
		Short: "Create a Cellery runtime in existing cluster",
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Example: "  cellery setup create existing",
	}
	cmd.Flags().BoolVarP(&isCompleteSetup, "complete", "c", false, "Create complete setup")
	cmd.Flags().BoolVarP(&isPersistentVolume, "persistent", "p", false, "Persistent volume")
	cmd.Flags().BoolVarP(&hasNfsStorage, "nfs", "n", false, "Has an NFS storage")
	cmd.Flags().BoolVarP(&isLoadBalancerIngressMode, "loadbalancer", "l", false, "Ingress mode is load balancer")
	return cmd
}

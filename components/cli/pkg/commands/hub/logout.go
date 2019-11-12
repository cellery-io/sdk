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

package hub

import (
	"fmt"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunLogout removes the saved credentials for a particular registry
func RunLogout(cli cli.Cli, registryURL string) error {
	fmt.Fprint(cli.Out(), "Logging out from Registry: "+util.Bold(registryURL))

	credManager := cli.CredManager()
	// Checking if the credentials are present
	isCredentialsPresent, err := credManager.HasCredentials(registryURL)
	if err != nil {
		return fmt.Errorf("error occurred while checking whether credentials are already saved, %v", err)
	}

	if !isCredentialsPresent {
		fmt.Fprint(cli.Out(), fmt.Sprintf("\nYou have not logged into %s Registry\n", util.Bold(registryURL)))
	} else {
		if err = credManager.RemoveCredentials(registryURL); err != nil {
			return fmt.Errorf("error occurred while removing Credentials, %v", err)
		}
		util.PrintSuccessMessage(fmt.Sprintf("Successfully logged out from Registry: %s",
			util.Bold(registryURL)))
	}
	return nil
}

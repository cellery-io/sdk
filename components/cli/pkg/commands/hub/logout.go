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

	"github.com/cellery-io/sdk/components/cli/pkg/registry/credentials"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunLogout removes the saved credentials for a particular registry
func RunLogout(registryURL string) {
	fmt.Print("Logging out from Registry: " + util.Bold(registryURL))

	credManager, err := credentials.NewCredManager()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while creating Credentials Manager", err)
	}

	// Checking if the credentials are present
	isCredentialsPresent, err := credManager.HasCredentials(registryURL)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while checking whether credentials are already saved", err)
	}

	if !isCredentialsPresent {
		fmt.Printf("\nYou have not logged into %s Registry\n", util.Bold(registryURL))
	} else {
		err = credManager.RemoveCredentials(registryURL)
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while removing Credentials", err)
		}

		util.PrintSuccessMessage(fmt.Sprintf("Successfully logged out from Registry: %s",
			util.Bold(registryURL)))
	}
}

/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/99designs/keyring"
	"github.com/nokia/docker-registry-client/registry"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunLogin requests the user for credentials and logs into a Cellery Registry
func RunLogin(registryURL string) {
	fmt.Println("Logging into Registry: " + util.Bold(registryURL))

	// Instantiating a native keyring
	ring, err := keyring.Open(keyring.Config{
		ServiceName: constants.CELLERY_HUB_KEYRING_NAME,
	})
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while logging in", err)
	}

	// Reading the existing credentials
	var registryCredentials = &util.RegistryCredentials{}
	if err == nil {
		ringItem, err := ring.Get(registryURL)
		if err == nil && ringItem.Data != nil {
			err = json.Unmarshal(ringItem.Data, registryCredentials)
		}
	}
	isCredentialsAlreadyPresent := err == nil && registryCredentials.Username != "" &&
		registryCredentials.Password != ""

	if isCredentialsAlreadyPresent {
		fmt.Println("Logging in with existing Credentials")
	} else {
		registryCredentials.Username, registryCredentials.Password, err = util.RequestCredentials()
		if err != nil {
			util.ExitWithErrorMessage("Error occurred while reading Credentials", err)
		}
	}

	// Initiating a connection to Cellery Registry (to validate credentials)
	fmt.Println()
	spinner := util.StartNewSpinner("Logging into Cellery Registry " + registryURL)
	_, err = registry.New("https://"+registryURL, registryCredentials.Username, registryCredentials.Password)
	if err != nil {
		spinner.Stop(false)
		if strings.Contains(err.Error(), "401") {
			util.ExitWithErrorMessage("Invalid Credentials", err)
		} else {
			util.ExitWithErrorMessage("Error occurred while initializing connection to the Cellery Registry", err)
		}
	}

	if !isCredentialsAlreadyPresent {
		// Saving the credentials
		spinner.SetNewAction("Saving credentials")
		credentialsData, err := json.Marshal(registryCredentials)
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error occurred while saving Credentials", err)
		}
		err = ring.Set(keyring.Item{
			Key:  registryURL,
			Data: credentialsData,
		})
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error occurred while saving Credentials", err)
		}
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully logged into Registry: %s", util.Bold(registryURL)))
}

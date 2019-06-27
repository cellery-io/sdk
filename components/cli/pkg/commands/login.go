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
	"fmt"
	"strings"

	"github.com/nokia/docker-registry-client/registry"

	"github.com/cellery-io/sdk/components/cli/pkg/registry/credentials"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunLogin requests the user for credentials and logs into a Cellery Registry
func RunLogin(registryURL string, username string, password string) {
	fmt.Println("Logging into Registry: " + util.Bold(registryURL))

	var registryCredentials = &credentials.RegistryCredentials{
		Registry: registryURL,
		Username: username,
		Password: password,
	}
	isCredentialsProvided := registryCredentials.Username != "" &&
		registryCredentials.Password != ""

	credManager, err := credentials.NewCredManager()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while creating Credentials Manager", err)
	}
	var isCredentialsAlreadyPresent bool
	if !isCredentialsProvided {
		// Reading the existing credentials
		registryCredentials, err = credManager.GetCredentials(registryURL)
		if registryCredentials == nil {
			registryCredentials = &credentials.RegistryCredentials{
				Registry: registryURL,
			}
		}
		// errors are ignored and considered as credentials not present
		isCredentialsAlreadyPresent = err == nil && registryCredentials.Username != "" &&
			registryCredentials.Password != ""
	}
	var isAuthorized chan bool
	var done chan bool
	runPreExitWithErrorTasks := func(spinner *util.Spinner) {
		if isAuthorized != nil {
			isAuthorized <- false
		}
		if done != nil {
			<-done
		}
		if spinner != nil {
			spinner.Stop(false)
		}
	}
	if isCredentialsProvided {
		fmt.Println("Logging in with provided Credentials")
	} else if isCredentialsAlreadyPresent {
		fmt.Println("Logging in with existing Credentials")
	} else {
		if password == "" {
			if username == "" {
				isAuthorized = make(chan bool)
				done = make(chan bool)
				registryCredentials.Username, registryCredentials.Password, err = credentials.FromBrowser(username,
					isAuthorized, done)
			} else {
				registryCredentials.Username, registryCredentials.Password, err = credentials.FromTerminal(username)
			}
			if err != nil {
				runPreExitWithErrorTasks(nil)
				util.ExitWithErrorMessage("Error occurred while reading Credentials", err)
			}
		} else {
			registryCredentials.Username = username
			registryCredentials.Password = password
		}
	}

	// Initiating a connection to Cellery Registry (to validate credentials)
	fmt.Println()
	spinner := util.StartNewSpinner("Logging into Cellery Registry " + registryURL)
	_, err = registry.New("https://"+registryURL, registryCredentials.Username, registryCredentials.Password)
	if err != nil {
		runPreExitWithErrorTasks(spinner)
		if strings.Contains(err.Error(), "401") {
			util.ExitWithErrorMessage("Invalid Credentials", err)
		} else {
			util.ExitWithErrorMessage("Error occurred while initializing connection to the Cellery Registry",
				err)
		}
	}

	if !isCredentialsAlreadyPresent {
		// Saving the credentials
		spinner.SetNewAction("Saving credentials")
		err = credManager.StoreCredentials(registryCredentials)
		if err != nil {
			runPreExitWithErrorTasks(spinner)
			util.ExitWithErrorMessage("Error occurred while saving Credentials", err)
		}
	}
	if isAuthorized != nil {
		isAuthorized <- true
	}
	if done != nil {
		<-done
	}
	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully logged into Registry: %s", util.Bold(registryURL)))
}

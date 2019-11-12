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
	"regexp"
	"strings"

	"github.com/nokia/docker-registry-client/registry"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
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
			isAuthorized, done, err = requestCredentialsFromUser(registryCredentials)
			if err != nil {
				runPreExitWithErrorTasks(nil)
				util.ExitWithErrorMessage("Error occurred while reading Credentials", err)
			}
		} else {
			registryCredentials.Username = username
			registryCredentials.Password = password
		}
	}

	fmt.Println()
	spinner := util.StartNewSpinner("Logging into Cellery Registry " + registryURL)
	err = validateCredentialsWithRegistry(registryCredentials)
	if err != nil {
		runPreExitWithErrorTasks(spinner)
		if strings.Contains(err.Error(), "401") {
			if isCredentialsAlreadyPresent {
				fmt.Println("\r\x1b[2K\U0000274C Failed to authenticate with existing credentials")
				registryCredentials.Username = ""
				registryCredentials.Password = ""
				// Requesting credentials from user since the existing credentials failed
				isAuthorized, done, err = requestCredentialsFromUser(registryCredentials)
				if err != nil {
					runPreExitWithErrorTasks(nil)
					util.ExitWithErrorMessage("Error occurred while reading Credentials", err)
				}

				spinner = util.StartNewSpinner("Logging into Cellery Registry " + registryURL)
				err = validateCredentialsWithRegistry(registryCredentials)
				if err != nil {
					runPreExitWithErrorTasks(spinner)
					if strings.Contains(err.Error(), "401") {
						util.ExitWithErrorMessage("Invalid Credentials", err)
					} else {
						util.ExitWithErrorMessage("Error occurred while initializing connection to the Cellery Registry",
							err)
					}
				}
				isCredentialsAlreadyPresent = false
			} else {
				util.ExitWithErrorMessage("Invalid Credentials", err)
			}
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

// requestCredentialsFromUser requests the user for credentials and returns the channels (or nil) created
func requestCredentialsFromUser(registryCredentials *credentials.RegistryCredentials) (chan bool, chan bool, error) {
	var isAuthorized chan bool
	var done chan bool
	var err error
	regex, err := regexp.Compile(constants.CENTRAL_REGISTRY_HOST_REGX)
	if err != nil {
		return nil, nil, err
	}
	if registryCredentials.Username == "" && regex.MatchString(registryCredentials.Registry) {
		isAuthorized = make(chan bool)
		done = make(chan bool)
		registryCredentials.Username, registryCredentials.Password, err = credentials.FromBrowser(
			registryCredentials.Username, isAuthorized, done)
	} else {
		registryCredentials.Username, registryCredentials.Password, err = credentials.FromTerminal(
			registryCredentials.Username)
	}
	if err != nil {
		return nil, nil, err
	}
	return isAuthorized, done, nil
}

// validateCredentialsWithRegistry initiates a connection to Cellery Registry to validate credentials
func validateCredentialsWithRegistry(registryCredentials *credentials.RegistryCredentials) error {
	registryPassword := registryCredentials.Password
	regex, err := regexp.Compile(constants.CENTRAL_REGISTRY_HOST_REGX)
	if err != nil {
		return err
	}
	if regex.MatchString(registryCredentials.Registry) {
		registryPassword = registryPassword + ":ping"
	}
	_, err = registry.New("https://"+registryCredentials.Registry, registryCredentials.Username, registryPassword)
	if err != nil {
		return err
	}
	return nil
}

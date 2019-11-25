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

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/registry/credentials"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

// RunLogin requests the user for credentials and logs into a Cellery Registry
func RunLogin(cli cli.Cli, registryURL string, username string, password string) error {
	var err error
	cli.CredReader().SetRegistry(registryURL)
	cli.CredReader().SetUserName(username)
	fmt.Fprintln(cli.Out(), "Logging into Registry: "+util.Bold(registryURL))

	var registryCredentials = &credentials.RegistryCredentials{
		Registry: registryURL,
		Username: username,
		Password: password,
	}
	isCredentialsProvided := registryCredentials.Username != "" && registryCredentials.Password != ""
	var isCredentialsAlreadyPresent bool
	if !isCredentialsProvided {
		// Reading the existing credentials
		registryCredentials, err = cli.CredManager().GetCredentials(registryURL)
		if registryCredentials == nil {
			registryCredentials = &credentials.RegistryCredentials{
				Registry: registryURL,
			}
		}
		// errors are ignored and considered as credentials not present
		isCredentialsAlreadyPresent = err == nil && registryCredentials.Username != "" &&
			registryCredentials.Password != ""
	}
	if isCredentialsProvided {
		fmt.Fprintln(cli.Out(), "Logging in with provided Credentials")
	} else if isCredentialsAlreadyPresent {
		fmt.Fprintln(cli.Out(), "Logging in with existing Credentials")
	} else {
		if password == "" {
			registryCredentials.Username, registryCredentials.Password, err = cli.CredReader().Read()
			if err != nil {
				cli.CredReader().Shutdown(false)
				return fmt.Errorf("error occurred while reading Credentials, %v", err)
			}
		}
	}

	fmt.Fprintln(cli.Out(), fmt.Sprintf("Logging into Cellery Registry %s", registryURL))
	err = validateCredentialsWithRegistry(registryCredentials)
	if err != nil {
		cli.CredReader().Shutdown(false)
		if strings.Contains(err.Error(), "401") {
			if isCredentialsAlreadyPresent {
				fmt.Fprintln(cli.Out(), "\r\x1b[2K\U0000274C Failed to authenticate with existing credentials")
				registryCredentials.Username = ""
				registryCredentials.Password = ""
				// Requesting credentials from user since the existing credentials failed
				registryCredentials.Username, registryCredentials.Password, err = cli.CredReader().Read()
				if err != nil {
					cli.CredReader().Shutdown(false)
					return fmt.Errorf("error occurred while reading Credentials, %v", err)
				}
				if err = cli.ExecuteTask(fmt.Sprintf("Logging into Cellery Registry %s", registryURL),
					fmt.Sprintf("Failed logging into Cellery Registry %s", registryURL), "", func() error {
						err = validateCredentialsWithRegistry(registryCredentials)
						if err != nil {
							cli.CredReader().Shutdown(false)
							if strings.Contains(err.Error(), "401") {
								return fmt.Errorf("invalid Credentials, %v", err)
							} else {
								return fmt.Errorf("error occurred while initializing connection to the Cellery Registry, %v",
									err)
							}
						}
						isCredentialsAlreadyPresent = false // To save the new credentials
						return nil
					}); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("invalid Credentials, %v", err)
			}
		} else {
			return fmt.Errorf("error occurred while initializing connection to the Cellery Registry, %v",
				err)
		}
	}

	if !isCredentialsAlreadyPresent {
		// Saving the credentials
		if err = cli.ExecuteTask("Saving credentials", "Failed to save credentials", "", func() error {
			err = cli.CredManager().StoreCredentials(registryCredentials)
			if err != nil {
				cli.CredReader().Shutdown(false)
				return fmt.Errorf("error occurred while saving Credentials, %v", err)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	cli.CredReader().Shutdown(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully logged into Registry: %s", util.Bold(registryURL)))
	return nil
}

// validateCredentialsWithRegistry initiates a connection to Cellery Registry to validate credentials
func validateCredentialsWithRegistry(registryCredentials *credentials.RegistryCredentials) error {
	registryPassword := registryCredentials.Password
	regex, err := regexp.Compile(constants.CentralRegistryHostRegx)
	if err != nil {
		return err
	}
	if regex.MatchString(registryCredentials.Registry) {
		registryPassword = registryPassword + ":ping"
	}
	_, err = registry.New("https://"+registryCredentials.Registry, registryCredentials.Username, registryPassword)
	return err
}

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

package credentials

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/util"
)

// Credentials are stored in a file in the Cellery user home. Only the owner of the file has permissions to this file.
const credentialsFileName = "credentials"
const credentialsFilePermissions os.FileMode = 0600

// FileCredentialsManager manages the credentials in credentials file
type FileCredentialsManager struct {
	credFile string
}

// NewFileCredentialsManager creates a new File based Credentials Manager
func NewFileCredentialsManager() (*FileCredentialsManager, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to identify the current user due to: %v", err)
	}
	credentialsFile := path.Join(currentUser.HomeDir, constants.CelleryHome, credentialsFileName)
	credManager := &FileCredentialsManager{
		credFile: credentialsFile,
	}
	return credManager, nil
}

// StoreCredentials stores the credentials in credentials file
func (credManager FileCredentialsManager) StoreCredentials(credentials *RegistryCredentials) error {
	if credentials.Registry == "" {
		return fmt.Errorf("registry to which the credentials belongs to is required for storing credentials")
	}
	registryKey := getCredManagerKeyForRegistry(credentials.Registry)

	credentialsMap, err := credManager.readCredentials()
	if err != nil {
		return fmt.Errorf("failed to fetch credentials due to: %v", err)
	}
	credentialsMap[registryKey] = credentials
	err = credManager.writeCredentials(credentialsMap)
	if err != nil {
		return fmt.Errorf("failed to save credentials due to: %v", err)
	}
	return nil
}

// GetCredentials retrieves the credentials from the credentials file
func (credManager FileCredentialsManager) GetCredentials(registry string) (*RegistryCredentials, error) {
	if registry == "" {
		return nil, fmt.Errorf(
			"registry to which the credentials belongs to is required for retrieving credentials")
	}
	registryKey := getCredManagerKeyForRegistry(registry)

	credentialsMap, err := credManager.readCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credentials due to: %v", err)
	}
	credentials := credentialsMap[registryKey]
	if credentials == nil {
		credentials = &RegistryCredentials{}
	}
	credentials.Registry = registry
	return credentials, nil
}

// RemoveCredentials removes the stored credentials from the credentials file
func (credManager FileCredentialsManager) RemoveCredentials(registry string) error {
	if registry == "" {
		return fmt.Errorf("registry to which the credentials belongs to is required for removing credentials")
	}
	registryKey := getCredManagerKeyForRegistry(registry)

	credentialsMap, err := credManager.readCredentials()
	if err != nil {
		return fmt.Errorf("failed to fetch credentials due to: %v", err)
	}
	delete(credentialsMap, registryKey)
	err = credManager.writeCredentials(credentialsMap)
	if err != nil {
		return fmt.Errorf("failed to update credentials file due to: %v", err)
	}
	return nil
}

// IsRegistryPresent checks if the registry credentials exists in the credentials file
func (credManager FileCredentialsManager) HasCredentials(registry string) (bool, error) {
	if registry == "" {
		return false, fmt.Errorf(
			"registry to which the credentials belongs to is required for checking for credentials")
	}
	registryKey := getCredManagerKeyForRegistry(registry)

	credentialsMap, err := credManager.readCredentials()
	if err != nil {
		return false, fmt.Errorf("failed to fetch credentials due to: %v", err)
	}
	_, hasKey := credentialsMap[registryKey]
	return hasKey, nil
}

// readCredentials reads the credentials stored in the credentials file
func (credManager FileCredentialsManager) readCredentials() (map[string]*RegistryCredentials, error) {
	// Reading the existing credentials in the file if it exists
	fileExists, err := util.FileExists(credManager.credFile)
	if err != nil {
		return nil, fmt.Errorf("failed to store the credentials in file due to: %v", err)
	}
	storedCredentials := &map[string]*RegistryCredentials{}
	if fileExists {
		credFileBytes, err := ioutil.ReadFile(credManager.credFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read the existing credentials file due to: %v", err)
		}
		err = json.Unmarshal(credFileBytes, storedCredentials)
		if err != nil {
			return nil, fmt.Errorf("failed to decode the credentials json due to: %v", err)
		}
	}
	return *storedCredentials, nil
}

// writeCredentials writes the credentials map to the credentials credentials file
func (credManager FileCredentialsManager) writeCredentials(credentialsMap map[string]*RegistryCredentials) error {
	credentialsBytes, err := json.Marshal(credentialsMap)
	if err != nil {
		return fmt.Errorf("failed to encode credentials for storing due to: %v", err)
	}
	err = ioutil.WriteFile(credManager.credFile, credentialsBytes, credentialsFilePermissions)
	if err != nil {
		return fmt.Errorf("failed to store the credentials in file due to: %v", err)
	}
	return nil
}

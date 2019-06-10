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
	"strings"

	"github.com/99designs/keyring"
)

// celleryHubKeyringName is the name of the keyring that will be created to store the credentials
const celleryHubKeyringName = "registrycelleryio"
const keyringDoesNotExistErrMsg = "The collection \"" + celleryHubKeyringName + "\" does not exist"

// KeyringCredManager holds the core of the KeyringCredManager which can be used to store the registry
// credentials in the native keyring
type KeyringCredManager struct {
	ring keyring.Keyring
}

// NewKeyringCredManager creates a new native keyring based credentials manager
func NewKeyringCredManager() (*KeyringCredManager, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName:              celleryHubKeyringName,
		FileDir:                  "~/.cellery/credentials",
		KeychainTrustApplication: false,
		AllowedBackends: []keyring.BackendType{keyring.SecretServiceBackend, keyring.KeychainBackend,
			keyring.KWalletBackend, keyring.WinCredBackend, keyring.PassBackend},
	})
	if err != nil {
		return nil, fmt.Errorf("keyring Credentials Manager not supported due to: %v", err)
	}
	credManager := &KeyringCredManager{
		ring: ring,
	}
	return credManager, nil
}

// StoreCredentials stores the credentials in the native keyring
func (credManager KeyringCredManager) StoreCredentials(credentials *RegistryCredentials) error {
	if credentials.Registry == "" {
		return fmt.Errorf("registry to which the credentials belongs to is required for storing credentials")
	}
	registryKey := getCredManagerKeyForRegistry(credentials.Registry)

	credentialsData, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("failed to create credentials json due to: %v", err)
	}
	err = credManager.ring.Set(keyring.Item{
		Key:  registryKey,
		Data: credentialsData,
	})
	if err != nil {
		return fmt.Errorf("failed to store credentials in the native keyring due to: %v", err)
	}
	return nil
}

// GetCredentials retrieves the previous stored credentials from the keyring
func (credManager KeyringCredManager) GetCredentials(registry string) (*RegistryCredentials, error) {
	if registry == "" {
		return nil, fmt.Errorf(
			"registry to which the credentials belongs to is required for retrieving credentials")
	}
	registryKey := getCredManagerKeyForRegistry(registry)

	var registryCredentials = &RegistryCredentials{
		Registry: registry,
	}
	ringItem, err := credManager.ring.Get(registryKey)
	if err != nil {
		if strings.Contains(err.Error(), keyringDoesNotExistErrMsg) {
			return registryCredentials, nil
		} else {
			return nil, fmt.Errorf("failed to get credentials from the native keyring due to: %v", err)
		}
	}
	if ringItem.Data == nil {
		return registryCredentials, nil
	}
	err = json.Unmarshal(ringItem.Data, registryCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the stored credentials due to: %v", err)
	}
	return registryCredentials, err
}

// RemoveCredentials removes the stored credentials from the keyring
func (credManager KeyringCredManager) RemoveCredentials(registry string) error {
	if registry == "" {
		return fmt.Errorf("registry to which the credentials belongs to is required for removing credentials")
	}
	registryKey := getCredManagerKeyForRegistry(registry)

	err := credManager.ring.Remove(registryKey)
	if err != nil {
		return fmt.Errorf("failed to remove credentials from the keyring due to: %v", err)
	}
	return nil
}

// IsRegistryPresent checks if the registry (key in the keyring) exists
func (credManager KeyringCredManager) HasCredentials(registry string) (bool, error) {
	if registry == "" {
		return false, fmt.Errorf(
			"registry to which the credentials belongs to is required for checking for credentials")
	}
	registryKey := getCredManagerKeyForRegistry(registry)

	keyList, err := credManager.ring.Keys()
	if err != nil {
		if strings.Contains(err.Error(), keyringDoesNotExistErrMsg) {
			return false, nil
		} else {
			return false, fmt.Errorf(
				"failed to check if the credentials exists due to: %v", err)
		}
	}
	var isCredentialsPresent bool
	for _, key := range keyList {
		if key == registryKey {
			isCredentialsPresent = true
			break
		}
	}
	return isCredentialsPresent, nil
}

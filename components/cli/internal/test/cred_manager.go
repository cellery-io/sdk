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

package test

import "github.com/cellery-io/sdk/components/cli/pkg/registry/credentials"

type MockCredManager struct {
	storedCredentials []*credentials.RegistryCredentials
}

func NewMockCredManager(opts ...func(*MockCredManager)) *MockCredManager {
	cli := &MockCredManager{}
	for _, opt := range opts {
		opt(cli)
	}
	return cli
}

func SetCredentials(registry, userName, password string) func(*MockCredManager) {
	return func(manager *MockCredManager) {
		manager.storedCredentials = append(manager.storedCredentials,
			&credentials.RegistryCredentials{
				Registry: registry,
				Username: userName,
				Password: password,
			})
	}
}

// StoreCredentials stores the mock credentials in credentials file.
func (credManager MockCredManager) StoreCredentials(credentials *credentials.RegistryCredentials) error {
	return nil
}

// GetCredentials retrieves the mock credentials from the credentials file.
func (credManager MockCredManager) GetCredentials(registry string) (*credentials.RegistryCredentials, error) {
	for _, cred := range credManager.storedCredentials {
		if cred.Registry == registry && cred.Username != "" && cred.Password != "" {
			return cred, nil
		}
	}
	return nil, nil
}

// RemoveCredentials mocks removal of the stored credentials from the credentials file.
func (credManager MockCredManager) RemoveCredentials(registry string) error {
	return nil
}

// IsRegistryPresent checks if the registry credentials exists in the credentials file.
func (credManager MockCredManager) HasCredentials(registry string) (bool, error) {
	for _, cred := range credManager.storedCredentials {
		if cred.Registry == registry && cred.Username != "" && cred.Password != "" {
			return true, nil
		}
	}
	return false, nil
}

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
	"fmt"
	"strings"
)

// RegistryCredentials holds the credentials of a registry
type RegistryCredentials struct {
	Registry string
	Username string
	Password string
}

// CredManager interface which defines the behaviour of all the credential managers.
type CredManager interface {
	// StoreCredentials stores the credentials in the relevant credentials store. The Registry field of the credentials
	// struct is required for this operation
	StoreCredentials(credentials *RegistryCredentials) error

	// GetCredentials retrieves the credentials from the relevant credentials store
	GetCredentials(registry string) (*RegistryCredentials, error)

	// RemoveCredentials removes the credentials from the relevant credentials store
	RemoveCredentials(registry string) error

	// HasCredentials checks whether the credentials are currently stored in the relevant credentials store
	HasCredentials(registry string) (bool, error)
}

// NewCredManager creates a new credentials manager instance
func NewCredManager() (CredManager, error) {
	var credManager CredManager
	credManager, err := NewKeyringCredManager()
	if err == nil {
		return credManager, nil
	}
	credManager, err = NewFileCredentialsManager()
	if err == nil {
		return credManager, nil
	}
	return nil, fmt.Errorf("failed to initialize a suitable credentials manager")
}

// Get a proper key for a registry
func getCredManagerKeyForRegistry(registry string) string {
	return strings.Split(registry, ":")[0]
}

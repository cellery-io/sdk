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

import (
	"fmt"

	"cellery.io/cellery/components/cli/pkg/runtime"
)

type MockRuntime struct {
	runtime            runtime.Runtime
	sysComponentStatus map[runtime.SystemComponent]bool
}

// NewMockRuntime returns a mock runtime.
func NewMockRuntime(opts ...func(*MockRuntime)) *MockRuntime {
	mockRuntime := &MockRuntime{}
	for _, opt := range opts {
		opt(mockRuntime)
	}
	return mockRuntime
}

func SetSysComponentStatus(status map[runtime.SystemComponent]bool) func(*MockRuntime) {
	return func(mockRuntime *MockRuntime) {
		mockRuntime.sysComponentStatus = status
	}
}

func (runtime *MockRuntime) IsComponentEnabled(component runtime.SystemComponent) (bool, error) {
	if runtime.sysComponentStatus != nil {
		return runtime.sysComponentStatus[component], nil
	}
	return false, fmt.Errorf("failed to check status of runtime")
}

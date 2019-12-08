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

import "cellery.io/cellery/components/cli/pkg/runtime"

// NewMockPlatform returns a mock platform.
func NewMockPlatform(opts ...func(*MockPlatform)) *MockPlatform {
	platform := &MockPlatform{}
	for _, opt := range opts {
		opt(platform)
	}
	return platform
}

type MockPlatform struct {
}

func (platform *MockPlatform) CreateK8sCluster() error {
	return nil
}

func (platform *MockPlatform) ConfigureSqlInstance() (runtime.MysqlDb, error) {
	return runtime.MysqlDb{}, nil
}

func (platform *MockPlatform) CreateStorage() error {
	return nil
}

func (platform *MockPlatform) CreateNfs() (runtime.Nfs, error) {
	return runtime.Nfs{}, nil
}

func (platform *MockPlatform) UpdateKubeConfig() error {
	return nil
}

func (platform *MockPlatform) RemoveCluster() error {
	return nil
}

func (platform *MockPlatform) RemoveSqlInstance() error {
	return nil
}

func (platform *MockPlatform) RemoveFileSystem() error {
	return nil
}

func (platform *MockPlatform) RemoveStorage() error {
	return nil
}

func (platform *MockPlatform) ClusterName() string {
	return ""
}

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

type MockCredReader struct {
	registry string
	userName string
}

func NewMockCredReader(opts ...func(*MockCredReader)) *MockCredReader {
	manager := &MockCredReader{}
	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

func (mockCredReader *MockCredReader) Read() (string, string, error) {
	return "", "", nil
}

func (mockCredReader *MockCredReader) SetRegistry(registry string) {
	mockCredReader.registry = registry
}

func (mockCredReader *MockCredReader) SetUserName(username string) {
	mockCredReader.userName = username
}

func (mockCredReader *MockCredReader) Shutdown(authorized bool) {
}

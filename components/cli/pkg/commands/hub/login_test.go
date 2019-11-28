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
	"testing"

	"cellery.io/cellery/components/cli/internal/test"
)

func TestRunLogin(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		username    string
		password    string
		credManager *test.MockCredManager
		credReader  *test.MockCredReader
	}{
		{
			name:     "login with credentials provided",
			url:      "registry.myhub.cellery.io",
			username: "alcie",
			password: "alice123",
			credManager: test.NewMockCredManager(test.SetCredentials("registry.myhub.cellery.io", "aclice",
				"alice123")),
			credReader: test.NewMockCredReader(),
		},
		{
			name:     "login without credential provided, with credentials already present",
			url:      "registry.myhub.cellery.io",
			username: "",
			password: "",
			credManager: test.NewMockCredManager(test.SetCredentials("registry.myhub.cellery.io", "aclice",
				"alice123")),
			credReader: test.NewMockCredReader(),
		},
		{
			name:        "login without credential provided, without credentials already present",
			url:         "registry.myhub.cellery.io",
			username:    "",
			password:    "",
			credManager: test.NewMockCredManager(),
			credReader:  test.NewMockCredReader(),
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			mockCli := test.NewMockCli(test.SetCredManager(tst.credManager), test.SetCredReader(tst.credReader))
			RunLogin(mockCli, tst.url, tst.username, tst.password)
		})
	}
}

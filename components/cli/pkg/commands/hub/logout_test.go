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

	"github.com/google/go-cmp/cmp"

	"github.com/cellery-io/sdk/components/cli/internal/test"
)

func TestRunLogout(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		loggedIn    bool
		credManager *test.MockCredManager
	}{
		{
			name:        "logout after logging in",
			url:         "myhub.cellery.io",
			loggedIn:    true,
			credManager: test.NewMockCredManager(test.SetCredentials("myhub.cellery.io", "aclice", "alice123")),
		},
		{
			name:        "logout without logging in",
			url:         "myhub.cellery.io",
			loggedIn:    false,
			credManager: test.NewMockCredManager(),
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			mockCli := test.NewMockCli(test.SetCredManager(tst.credManager))
			err := RunLogout(mockCli, tst.url)
			if tst.loggedIn {
				if err != nil {
					t.Errorf("error in RunLogout, %v", err)
				}
			} else {
				if diff := cmp.Diff("Logging out from Registry: myhub.cellery.io\nYou have not logged into myhub.cellery.io Registry\n", mockCli.OutBuffer().String()); diff != "" {
					t.Errorf("RunLogout: not logged in (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

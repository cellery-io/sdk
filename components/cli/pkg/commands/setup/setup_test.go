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

package setup

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"cellery.io/cellery/components/cli/internal/test"
)

func TestRunSetup(t *testing.T) {
	tests := []struct {
		name      string
		mockCli   *test.MockCli
		selection string
	}{
		{
			name:      "create",
			mockCli:   test.NewMockCli(test.SetActionItem(1)),
			selection: "Create",
		},
		{
			name:      "manage",
			mockCli:   test.NewMockCli(test.SetActionItem(2)),
			selection: "Manage",
		},
		{
			name:      "modify",
			mockCli:   test.NewMockCli(test.SetActionItem(3)),
			selection: "Modify",
		},
		{
			name:      "switch",
			mockCli:   test.NewMockCli(test.SetActionItem(4)),
			selection: "Switch",
		},
		{
			name:      "exit",
			mockCli:   test.NewMockCli(test.SetActionItem(5)),
			selection: "EXIT",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunSetup(tst.mockCli)
			if err != nil {
				t.Errorf("error in RunSetup, %v", err)
			}
			if diff := cmp.Diff(tst.selection, tst.mockCli.Selection().Label); diff != "" {
				t.Errorf("invalid selection (-want, +got)\n%v", diff)
			}
		})
	}
}

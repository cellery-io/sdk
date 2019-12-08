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

func TestRunSetupCleanup(t *testing.T) {
	tests := []struct {
		name             string
		expectedToPass   bool
		expectedErrorMsg string
		mockCli          *test.MockCli
		mockPlatform     *test.MockPlatform
		existingCluster  bool
		removeKnative    bool
		removeIstio      bool
		removeIngress    bool
		removeHpa        bool
	}{
		{
			name: "remove cellery runtime and mock platform",
			mockCli: test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			mockPlatform:   test.NewMockPlatform(),
			expectedToPass: true,
		},
		{
			name: "remove cellery runtime from existing cluster",
			mockCli: test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			existingCluster: true,
			removeKnative:   true,
			removeIstio:     true,
			removeIngress:   true,
			removeHpa:       true,
			expectedToPass:  true,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			var err error
			if tst.existingCluster {
				err = RunSetupCleanup(tst.mockCli, nil, tst.removeKnative, tst.removeIstio, tst.removeIngress,
					tst.removeHpa, true)
			} else {
				err = RunSetupCleanup(tst.mockCli, tst.mockPlatform, tst.removeKnative, tst.removeIstio, tst.removeIngress,
					tst.removeHpa, true)
			}
			if tst.expectedToPass {
				if err != nil {
					t.Errorf("error in RunSetupCleanup, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expectedErrorMsg, err.Error()); diff != "" {
					t.Errorf("invalid error message (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

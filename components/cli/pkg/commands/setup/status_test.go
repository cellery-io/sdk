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

func TestRunSetupStatus(t *testing.T) {
	tests := []struct {
		name             string
		expectedToPass   bool
		expectedErrorMsg string
		mockCli          *test.MockCli
	}{
		{
			name: "connected to a k8s cluster",
			mockCli: test.NewMockCli(test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli(test.SetK8sVersions("v1.14.1", "v1.10.3"),
					test.SetClusterName("my-cluster")))),
			expectedToPass: true,
		},
		{
			name: "not connected to a k8s server",
			mockCli: test.NewMockCli(test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			expectedToPass:   false,
			expectedErrorMsg: "failed to get setup status, unable to connect to the kubernetes cluster",
		},
		{
			name: "not connected to a k8s cluster",
			mockCli: test.NewMockCli(test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli(test.SetK8sVersions("v1.14.1", "v1.10.3")))),
			expectedToPass:   false,
			expectedErrorMsg: "error getting cluster name, not connected to a cluster",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunSetupStatus(tst.mockCli)
			if tst.expectedToPass {
				if err != nil {
					t.Errorf("error in RunSetupStatus, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expectedErrorMsg, err.Error()); diff != "" {
					t.Errorf("invalid error message (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

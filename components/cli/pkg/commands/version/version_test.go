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

package version

import (
	"testing"

	"github.com/cellery-io/sdk/components/cli/internal/test"
)

func TestRunVersion(t *testing.T) {
	mockKubeCli := test.NewMockKubeCli(test.SetK8sVersions("v1.14.1", "v1.10.3"))
	mockDockerCli := test.NewMockDockerCli(test.SetServerVersion("12.0.0"), test.SetClientVersion("10.0.1"))
	mockBalExecutor := test.NewMockBalExecutor(test.SetBalVersion("1.0.3"))
	mockCli := test.NewMockCli(test.SetKubeCli(mockKubeCli), test.SetDockerCli(mockDockerCli), test.SetBalExecutor(mockBalExecutor))
	tests := []struct {
		name string
	}{
		{
			name: "cellery version",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunVersion(mockCli)
			if err != nil {
				t.Errorf("error in RunVersion, %v", err)
			}
		})
	}
}

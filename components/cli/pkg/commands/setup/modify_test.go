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
	"cellery.io/cellery/components/cli/pkg/runtime"
)

func TestRunSetupModify(t *testing.T) {
	tests := []struct {
		name                 string
		expectedToPass       bool
		expectedErrorMsg     string
		apimEnabled          bool
		observabilityEnabled bool
		scaleToZeroEnabled   bool
		hpaEnabled           bool
		apim                 runtime.Selection
		observability        runtime.Selection
		scaleToZero          runtime.Selection
		hpa                  runtime.Selection
	}{
		{
			name:           "enable Apim",
			apimEnabled:    false,
			apim:           runtime.Enable,
			expectedToPass: true,
		},
		{
			name:           "disable Apim",
			apimEnabled:    true,
			apim:           runtime.Disable,
			expectedToPass: true,
		},
		{
			name:                 "enable Apim while observability already enabled",
			apimEnabled:          false,
			apim:                 runtime.Enable,
			observabilityEnabled: true,
			expectedToPass:       true,
		},
		{
			name:                 "enable Observability",
			observabilityEnabled: false,
			observability:        runtime.Enable,
			expectedToPass:       true,
		},
		{
			name:                 "disable Observability",
			observabilityEnabled: true,
			observability:        runtime.Disable,
			expectedToPass:       true,
		},
		{
			name:               "enable Scale to Zero",
			scaleToZeroEnabled: false,
			scaleToZero:        runtime.Enable,
			expectedToPass:     true,
		},
		{
			name:               "disable Scale to Zero",
			scaleToZeroEnabled: true,
			scaleToZero:        runtime.Disable,
			expectedToPass:     true,
		},
		{
			name:           "enable hpa",
			hpaEnabled:     false,
			hpa:            runtime.Enable,
			expectedToPass: true,
		},
		{
			name:           "disable hpa",
			hpaEnabled:     true,
			hpa:            runtime.Disable,
			expectedToPass: true,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			sysComponentStatus := make(map[runtime.SystemComponent]bool)
			sysComponentStatus[runtime.ApiManager] = tst.apimEnabled
			sysComponentStatus[runtime.Observability] = tst.observabilityEnabled
			sysComponentStatus[runtime.ScaleToZero] = tst.scaleToZeroEnabled
			sysComponentStatus[runtime.HPA] = tst.hpaEnabled

			mockCli := test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime(test.SetSysComponentStatus(sysComponentStatus))),
				test.SetKubeCli(test.NewMockKubeCli()))

			err := RunSetupModify(mockCli, tst.apim, tst.observability, tst.scaleToZero, tst.hpa)
			if tst.expectedToPass {
				if err != nil {
					t.Errorf("error in RunSetupModify, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expectedErrorMsg, err.Error()); diff != "" {
					t.Errorf("invalid error message (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

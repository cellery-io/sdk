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

func TestRunSetupCreateOnExistingCluster(t *testing.T) {
	tests := []struct {
		name                      string
		expectedToPass            bool
		expectedErrorMsg          string
		mockCli                   *test.MockCli
		complete                  bool
		persistVolume             bool
		hasNfs                    bool
		nfs                       runtime.Nfs
		isLoadBalancerIngressMode bool
		nodeportIpAddress         string
	}{
		{
			name: "create basic cellery runtime without persistence",
			mockCli: test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			complete:       false,
			persistVolume:  false,
			nfs:            runtime.Nfs{},
			expectedToPass: true,
		},
		{
			name: "create complete cellery runtime without persistence",
			mockCli: test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			complete:       true,
			persistVolume:  false,
			nfs:            runtime.Nfs{},
			expectedToPass: true,
		},
		{
			name: "create basic cellery runtime with persistence",
			mockCli: test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			complete:       false,
			persistVolume:  true,
			nfs:            runtime.Nfs{},
			expectedToPass: true,
		},
		{
			name: "create complete cellery runtime with persistence",
			mockCli: test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			complete:       true,
			persistVolume:  true,
			nfs:            runtime.Nfs{},
			expectedToPass: true,
		},
		{
			name: "create complete cellery runtime with persistence and with nfs",
			mockCli: test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			complete:       true,
			persistVolume:  true,
			hasNfs:         true,
			nfs:            runtime.Nfs{NfsServerIp: "120.120.120.120", FileShare: "/data"},
			expectedToPass: true,
		},
		{
			name: "create complete cellery runtime with nodeport ingress mode (default)",
			mockCli: test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			complete:       true,
			expectedToPass: true,
		},
		{
			name: "create complete cellery runtime with nodeport ingress mode (custom)",
			mockCli: test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			complete:          true,
			nodeportIpAddress: "111.111.111.111",
			expectedToPass:    true,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunSetupCreateCelleryRuntime(tst.mockCli, tst.complete, tst.persistVolume, tst.hasNfs,
				tst.isLoadBalancerIngressMode, tst.nfs, runtime.MysqlDb{}, tst.nodeportIpAddress)
			if tst.expectedToPass {
				if err != nil {
					t.Errorf("error in RunSetupCreateCelleryRuntime, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expectedErrorMsg, err.Error()); diff != "" {
					t.Errorf("invalid error message (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

func TestRunSetupCreateWithPlatform(t *testing.T) {
	tests := []struct {
		name                      string
		expectedToPass            bool
		expectedErrorMsg          string
		mockCli                   *test.MockCli
		mockPlatform              *test.MockPlatform
		complete                  bool
		persistVolume             bool
		hasNfs                    bool
		isLoadBalancerIngressMode bool
	}{
		{
			name: "create basic cellery runtime on a platform",
			mockCli: test.NewMockCli(
				test.SetFileSystem(test.NewMockFileSystem()),
				test.SetRuntime(test.NewMockRuntime()),
				test.SetKubeCli(test.NewMockKubeCli())),
			mockPlatform:              test.NewMockPlatform(),
			complete:                  false,
			persistVolume:             false,
			expectedToPass:            true,
			hasNfs:                    true,
			isLoadBalancerIngressMode: true,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			nodeportIp, db, nfs, err := RunSetupCreateCelleryPlatform(tst.mockCli, tst.mockPlatform)
			if tst.expectedToPass {
				if err != nil {
					t.Errorf("error in RunSetupCreateCelleryPlatform, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expectedErrorMsg, err.Error()); diff != "" {
					t.Errorf("invalid error message (-want, +got)\n%v", diff)
				}
			}
			err = RunSetupCreateCelleryRuntime(tst.mockCli, tst.complete, tst.persistVolume, tst.hasNfs,
				tst.isLoadBalancerIngressMode, nfs, db, nodeportIp)
			if tst.expectedToPass {
				if err != nil {
					t.Errorf("error in RunSetupCreateCelleryRuntime, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expectedErrorMsg, err.Error()); diff != "" {
					t.Errorf("invalid error message (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"cellery.io/cellery/components/cli/internal/test"
)

func TestRunSetupListClusters(t *testing.T) {
	config, err := ioutil.ReadFile(filepath.Join("testdata", "config", "my_config.json"))
	if err != nil {
		t.Errorf("failed to read my_config.json file")
	}
	fooConfig, err := ioutil.ReadFile(filepath.Join("testdata", "config", "foo_config.json"))
	if err != nil {
		t.Errorf("failed to read my_config.json file")
	}
	tests := []struct {
		name             string
		expectedToPass   bool
		expectedErrorMsg string
		mockCli          *test.MockCli
	}{
		{
			name:           "list clusters",
			mockCli:        test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli(test.SetConfig(config)))),
			expectedToPass: true,
		},
		{
			name:             "list clusters without config file set",
			mockCli:          test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli())),
			expectedToPass:   false,
			expectedErrorMsg: "failed to get contexts, error getting context list, failed to get config",
		},
		{
			name:           "list clusters with config file set but having an error in config file",
			mockCli:        test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli(test.SetConfig(fooConfig)))),
			expectedToPass: false,
			expectedErrorMsg: "failed to get contexts, error trying to unmarshal contexts output, invalid character '/' looking for beginning of " +
				"object key string",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunSetupListClusters(tst.mockCli)
			if tst.expectedToPass {
				if err != nil {
					t.Errorf("error in RunSetupListClusters, %v", err)
				}
			} else {
				if diff := cmp.Diff(tst.expectedErrorMsg, err.Error()); diff != "" {
					t.Errorf("invalid error message (-want, +got)\n%v", diff)
				}
			}
		})
	}
}

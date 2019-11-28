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

package instance

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"cellery.io/cellery/components/cli/internal/test"
)

func TestExportAutoscalePolicies(t *testing.T) {
	petBeAutoCell, err := ioutil.ReadFile(filepath.Join("testdata", "cells", "pet-be-auto.json"))
	if err != nil {
		t.Errorf("failed to read mock cell yaml file")
	}
	cellMap := make(map[string][]byte)
	cellMap["pet-be-auto"] = petBeAutoCell
	mockCli := test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli(test.WithCellsAsBytes(cellMap))))
	outputFile, err := ioutil.TempFile("", "exportpolicy*.yaml")
	if err != nil {
		t.Errorf("failed create yaml file to export to")
	}
	tests := []struct {
		name string
	}{
		{
			name: "export autoscale policy",
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunExportAutoscalePolicies(mockCli, celleryInstance, "pet-be-auto", outputFile.Name())
			if err != nil {
				t.Errorf("error in RunExportAutoscalePolicies, %v", err)
			}
		})
	}
}

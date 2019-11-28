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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"cellery.io/cellery/components/cli/internal/test"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
)

func TestRunRouteTraffic(t *testing.T) {
	petBeDepCell, err := ioutil.ReadFile(filepath.Join("testdata", "cells", "pet-be-dep.json"))
	if err != nil {
		t.Errorf("failed to read mock pet-be-dep cell yaml file")
	}
	petBeTargetCell, err := ioutil.ReadFile(filepath.Join("testdata", "cells", "pet-be-target.json"))
	if err != nil {
		t.Errorf("failed to read mock pet-be-target cell yaml file")
	}
	petFeSrcCell, err := ioutil.ReadFile(filepath.Join("testdata", "cells", "pet-fe-src.json"))
	if err != nil {
		t.Errorf("failed to read mock pet-fe-src cell yaml file")
	}
	cellMap := make(map[string][]byte)
	cellMap["pet-be-dep"] = petBeDepCell
	cellMap["pet-be-target"] = petBeTargetCell
	cellMap["pet-fe-src"] = petFeSrcCell

	mockKubeCli := test.NewMockKubeCli(test.WithCellsAsBytes(cellMap))
	mockCli := test.NewMockCli(test.SetKubeCli(mockKubeCli))
	tests := []struct {
		name               string
		MockCli            *test.MockCli
		sourceInstances    []string
		dependencyInstance string
		targetInstance     string
		percentage         int
	}{
		{
			name:               "route traffic",
			MockCli:            test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			sourceInstances:    []string{"pet-fe-src"},
			dependencyInstance: "pet-be-dep",
			targetInstance:     "pet-be-target",
			percentage:         40,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunRouteTrafficCommand(mockCli, tst.sourceInstances, tst.dependencyInstance, tst.targetInstance, tst.percentage, false, true)
			if err != nil {
				t.Errorf("error in RunRouteTrafficCommand, %v", err)
			}
		})
	}
}

func TestBuildRouteArtifact(t *testing.T) {
	petBeDepCell, err := ioutil.ReadFile(filepath.Join("testdata", "cells", "pet-be-dep.json"))
	if err != nil {
		t.Errorf("failed to read mock pet-be-dep cell yaml file")
	}
	petBeTargetCell, err := ioutil.ReadFile(filepath.Join("testdata", "cells", "pet-be-target.json"))
	if err != nil {
		t.Errorf("failed to read mock pet-be-target cell yaml file")
	}
	petFeSrcCell, err := ioutil.ReadFile(filepath.Join("testdata", "cells", "pet-fe-src.json"))
	if err != nil {
		t.Errorf("failed to read mock pet-fe-src cell yaml file")
	}
	cellMap := make(map[string][]byte)
	cellMap["pet-be-dep"] = petBeDepCell
	cellMap["pet-be-target"] = petBeTargetCell
	cellMap["pet-fe-src"] = petFeSrcCell

	petFeSrcVsBytes, err := ioutil.ReadFile(filepath.Join("testdata", "virtual-services", "pet-fe-src-vs.json"))
	if err != nil {
		t.Errorf("failed to read mock pet-fe-src-vs cell yaml file")
	}
	petFeSrcVs := kubernetes.VirtualService{}
	err = json.Unmarshal(petFeSrcVsBytes, &petFeSrcVs)
	if err != nil {
		t.Errorf("failed to unmarshall petFeSrcVsBytes, %v", err)
	}
	vsMap := make(map[string]kubernetes.VirtualService)
	vsMap["pet-fe-src--vs"] = petFeSrcVs

	mockKubeCli := test.NewMockKubeCli(test.WithCellsAsBytes(cellMap), test.WithVirtualServices(vsMap))
	mockCli := test.NewMockCli(test.SetKubeCli(mockKubeCli))
	tests := []struct {
		name               string
		MockCli            *test.MockCli
		sourceInstances    []string
		dependencyInstance string
		targetInstance     string
		percentage         int
	}{
		{
			name:               "route traffic",
			MockCli:            test.NewMockCli(test.SetKubeCli(mockKubeCli)),
			sourceInstances:    []string{"pet-fe-src"},
			dependencyInstance: "pet-be-dep",
			targetInstance:     "pet-be-target",
			percentage:         40,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			artifactFile := fmt.Sprintf("./%s-routing-artifacts.yaml", tst.dependencyInstance)
			defer func() {
				err := os.Remove(artifactFile)
				if err != nil {
					t.Errorf("failed to remove artifacts file, %v", err)
				}
			}()
			err := buildRouteArtifact(mockCli, tst.sourceInstances, tst.dependencyInstance, tst.targetInstance,
				tst.percentage, false, true)
			if err != nil {
				t.Errorf("error in buildRouteArtifact, %v", err)
			}
			actualArtifacts, err := ioutil.ReadFile(artifactFile)
			if err != nil {
				t.Errorf("failed to read route artifact file")
			}
			expectedArtifacts, err := ioutil.ReadFile(filepath.Join("testdata", "expected", artifactFile))
			if err != nil {
				t.Errorf("failed to read expected artifacts file")
			}

			if diff := cmp.Diff(string(expectedArtifacts), string(actualArtifacts)); diff != "" {
				t.Errorf("invalid file content (-want, +got)\n%v", diff)
			}
		})
	}
}

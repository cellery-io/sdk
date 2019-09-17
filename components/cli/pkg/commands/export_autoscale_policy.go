/*
 * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http:www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunExportAutoscalePolicies(kind kubectl.InstanceKind, instance string, outputfile string) error {
	polExportSpinner := util.StartNewSpinner("Exporting autoscale policies")
	ik := string(kind)
	data, err := kubectl.GetInstanceBytes(ik, instance)
	if err != nil {
		polExportSpinner.Stop(false)
		return err
	}
	originalResource := &kubectl.CompositeResource{}
	err = json.Unmarshal(data, originalResource)
	if err != nil {
		polExportSpinner.Stop(false)
		return err
	}
	sp := &kubectl.AutoScalingPolicy{}

	for _, v := range originalResource.Spec.Components {
		sp.Components = append(sp.Components, kubectl.ComponentScalePolicy{
			Name:          v.Metadata.Name,
			ScalingPolicy: v.Spec.ScalingPolicy,
		})
	}

	spData, err := json.Marshal(sp)
	if err != nil {
		polExportSpinner.Stop(false)
		return err
	}
	yamlBytes, err := yaml.JSONToYAML(spData)
	if err != nil {
		polExportSpinner.Stop(false)
		return err
	}
	// write to a file
	file := outputfile
	if file == "" {
		if kind == kubectl.InstanceKindCell {
			file = filepath.Join("./", fmt.Sprintf("%s-%s-autoscalepolicy.yaml", "cell", instance))
		} else {
			file = filepath.Join("./", fmt.Sprintf("%s-%s-autoscalepolicy.yaml", "composite", instance))
		}
	} else {
		err := ensureDir(file)
		if err != nil {
			return err
		}
	}
	err = writeToFile(yamlBytes, file)
	if err != nil {
		polExportSpinner.Stop(false)
		return err
	}
	polExportSpinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully exported autoscale policies for instance %s to %s", instance, file))
	return nil
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0644)
	if err != nil {
		return err
	}
	return nil
}

func writeToFile(content []byte, file string) error {
	err := ioutil.WriteFile(file, content, 0644)
	if err != nil {
		return err
	}
	return nil
}

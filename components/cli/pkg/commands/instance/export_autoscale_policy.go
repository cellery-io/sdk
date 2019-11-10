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

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunExportAutoscalePolicies(cli cli.Cli, kind kubernetes.InstanceKind, instance string, outputfile string) error {
	var err error
	var spData, yamlBytes []byte
	var sp *kubernetes.AutoScalingPolicy
	if err = cli.ExecuteTask("Retrieving autoscale policy data", "Failed to retrieve autoscale policy data",
		"", func() error {
			sp, err = getAutoscalePolicies(cli, kind, instance)
			return err
		}); err != nil {
		return fmt.Errorf("failed to retrieve autoscale policy data, %v", err)
	}
	if spData, err = json.Marshal(sp); err != nil {
		return err
	}
	if yamlBytes, err = yaml.JSONToYAML(spData); err != nil {
		return err
	}
	// write to a file
	file := outputfile
	if file == "" {
		if kind == kubernetes.InstanceKindCell {
			file = filepath.Join("./", fmt.Sprintf("%s-%s-autoscalepolicy.yaml", "cell", instance))
		} else {
			file = filepath.Join("./", fmt.Sprintf("%s-%s-autoscalepolicy.yaml", "composite", instance))
		}
	} else {
		if err := ensureDir(file); err != nil {
			return err
		}
	}
	if err = writeToFile(yamlBytes, file); err != nil {
		return err
	}
	util.PrintSuccessMessage(fmt.Sprintf("Successfully exported autoscale policies for instance %s to %s", instance, file))
	return nil
}

func getAutoscalePolicies(cli cli.Cli, kind kubernetes.InstanceKind, instance string) (*kubernetes.AutoScalingPolicy, error) {
	var err error
	var data []byte
	ik := string(kind)
	if data, err = cli.KubeCli().GetInstanceBytes(ik, instance); err != nil {
		return nil, fmt.Errorf("failed to get instance bytes, %v", err)
	}
	originalResource := &kubernetes.ScaleResource{}
	if err = json.Unmarshal(data, originalResource); err != nil {
		return nil, fmt.Errorf("failed to unmarshall, %v", err)
	}
	sp := &kubernetes.AutoScalingPolicy{}

	for _, v := range originalResource.Spec.Components {
		sp.Components = append(sp.Components, kubernetes.ComponentScalePolicy{
			Name:          v.Metadata.Name,
			ScalingPolicy: v.Spec.ScalingPolicy,
		})
	}

	if kind == kubernetes.InstanceKindCell {
		if gwScalePolicyExists(originalResource) {
			sp.Gateway = kubernetes.GwScalePolicy{
				ScalingPolicy: originalResource.Spec.Gateway.Spec.ScalingPolicy,
			}
		} else {
			sp.Gateway = kubernetes.GwScalePolicy{
				ScalingPolicy: struct {
					Replicas int32 `json:"replicas"`
				}{1},
			}
		}
	}
	return sp, nil
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

func gwScalePolicyExists(scaleResource *kubernetes.ScaleResource) bool {
	return &scaleResource.Spec.Gateway != nil && scaleResource.Spec.Gateway.Spec.ScalingPolicy != nil
}

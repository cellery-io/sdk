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

	"github.com/ghodss/yaml"
	"github.com/mattbaird/jsonpatch"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/util"
)

func RunApplyAutoscalePolicies(cli cli.Cli, kind kubernetes.InstanceKind, instance string, file string) error {
	var err error
	var originalData, desiredData []byte
	ik := string(kind)
	if err = cli.ExecuteTask("Preparing autoscale policy data to apply", "Failed to prepare patch",
		"", func() error {
			originalData, desiredData, err = createPatch(cli, kind, instance, file)
			return err
		}); err != nil {
		return fmt.Errorf("failed to create patch, %v", err)
	}
	patch, err := jsonpatch.CreatePatch(originalData, desiredData)
	if err != nil {
		return fmt.Errorf("failed to create patch, %v", err)
	}

	if len(patch) == 0 {
		util.PrintSuccessMessage(fmt.Sprintf("Nothing to apply. Scaling policies for %q matches with policy file %q", instance, file))
		return nil
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshall patch, %v", err)
	}
	if err = cli.ExecuteTask("Applying autoscale policies", "Failed to apply autoscale policies",
		"", func() error {
			err = cli.KubeCli().JsonPatch(ik, instance, string(patchBytes))
			return err
		}); err != nil {
		return fmt.Errorf("failed to apply patch, %v", err)
	}
	util.PrintSuccessMessage(fmt.Sprintf("Successfully applied autoscale policies for instance %q", instance))
	return nil
}

func createPatch(cli cli.Cli, kind kubernetes.InstanceKind, instance string, file string) ([]byte, []byte, error) {
	var originalData, desiredData []byte
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return originalData, desiredData, fmt.Errorf("error reading file %s, %v", file, err)
	}
	newScalePolicy := &kubernetes.AutoScalingPolicy{}
	if err = yaml.Unmarshal(fileData, &newScalePolicy); err != nil {
		return originalData, desiredData, fmt.Errorf("failed to unmarshall data in file %s, %v", file, err)
	}
	ik := string(kind)
	instanceData, err := cli.KubeCli().GetInstanceBytes(ik, instance)
	if err != nil {
		return originalData, desiredData, fmt.Errorf("failed to get instance bytes of instance %s, %v", instance, err)
	}
	originalResource := &kubernetes.ScaleResource{}
	if err = json.Unmarshal(instanceData, originalResource); err != nil {
		return originalData, desiredData, fmt.Errorf("failed to unmarshall instance data of instance %s, %v", instance, err)
	}
	originalData, err = json.Marshal(originalResource)
	if err != nil {
		return originalData, desiredData, fmt.Errorf("failed to marshall original data, %v", err)
	}
	// we are modifying the original resource here as we already Marshal the required data
	desiredResource := originalResource
	for _, spComponent := range newScalePolicy.Components {
		for i := range originalResource.Spec.Components {
			if spComponent.Name == originalResource.Spec.Components[i].Metadata.Name {
				overridable, err := isOverridable(originalResource.Spec.Components[i].Spec.ScalingPolicy)
				if err != nil {
					return originalData, desiredData, err
				}
				if overridable {
					desiredResource.Spec.Components[i].Spec.ScalingPolicy = spComponent.ScalingPolicy
				}
			}
		}
	}
	if kind == kubernetes.InstanceKindCell && newScalePolicy.Gateway.ScalingPolicy != nil {
		overridable, err := isOverridable(originalResource.Spec.Gateway.Spec.ScalingPolicy)
		if err != nil {
			return originalData, desiredData, err
		}
		if overridable {
			desiredResource.Spec.Gateway.Spec.ScalingPolicy = newScalePolicy.Gateway.ScalingPolicy
		}
	}

	desiredData, err = json.Marshal(desiredResource)
	if err != nil {
		return originalData, desiredData, fmt.Errorf("failed to marshall desired resource, %v", err)
	}
	return originalData, desiredData, nil
}

func isOverridable(o interface{}) (bool, error) {
	hpaBytes, err := json.Marshal(o)
	if err != nil {
		return false, err
	}
	hpa := &kubernetes.ScalingPolicy{}
	err = json.Unmarshal(hpaBytes, hpa)
	if err != nil {
		return false, err
	}
	if &hpa.Hpa == nil {
		// no HPA, can be overridden
		return true, nil
	}
	if hpa.Hpa.Overridable == nil {
		// flag not provided, should be overridable
		return true, nil
	}
	return *hpa.Hpa.Overridable, nil
}

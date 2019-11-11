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

package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/mattbaird/jsonpatch"

	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunApplyAutoscalePolicies(kind kubernetes.InstanceKind, instance string, file string) error {
	spinner := util.StartNewSpinner("Applying autoscale policies")

	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	newScalePolicy := &kubernetes.AutoScalingPolicy{}
	err = yaml.Unmarshal(fileData, &newScalePolicy)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	ik := string(kind)
	instanceData, err := kubernetes.GetInstanceBytes(ik, instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	originalResource := &kubernetes.ScaleResource{}
	err = json.Unmarshal(instanceData, originalResource)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	originalData, err := json.Marshal(originalResource)
	if err != nil {
		return err
	}
	// we are modifying the original resource here as we already Marshal the required data
	desiredResource := originalResource
	for _, spComponent := range newScalePolicy.Components {
		for i := range originalResource.Spec.Components {
			if spComponent.Name == originalResource.Spec.Components[i].Metadata.Name {
				overridable, err := isOverridable(originalResource.Spec.Components[i].Spec.ScalingPolicy)
				if err != nil {
					return err
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
			return err
		}
		if overridable {
			desiredResource.Spec.Gateway.Spec.ScalingPolicy = newScalePolicy.Gateway.ScalingPolicy
		}
	}

	desiredData, err := json.Marshal(desiredResource)
	if err != nil {
		return err
	}
	patch, err := jsonpatch.CreatePatch(originalData, desiredData)
	if err != nil {
		return err
	}

	if len(patch) == 0 {
		spinner.Stop(true)
		util.PrintSuccessMessage(fmt.Sprintf("Nothing to apply. Scaling policies for %q matches with policy file %q", instance, file))
		return nil
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	err = kubernetes.JsonPatch(ik, instance, string(patchBytes))
	if err != nil {
		spinner.Stop(false)
		return err
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully applied autoscale policies for instance %q", instance))
	return nil
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

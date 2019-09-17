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

	"github.com/ghodss/yaml"
	"github.com/mattbaird/jsonpatch"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunApplyAutoscalePolicies(kind kubectl.InstanceKind, instance string, file string) error {
	spinner := util.StartNewSpinner("Applying autoscale policies")

	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	sp := &kubectl.AutoScalingPolicy{}
	err = yaml.Unmarshal(fileData, &sp)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	ik := string(kind)
	instanceData, err := kubectl.GetInstanceBytes(ik, instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	originalResource := &kubectl.CompositeResource{}
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
	for _, v1 := range sp.Components {
		for i := range originalResource.Spec.Components {
			if v1.Name == originalResource.Spec.Components[i].Metadata.Name {
				desiredResource.Spec.Components[i].Spec.ScalingPolicy = v1.ScalingPolicy
			}
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
		util.PrintSuccessMessage(fmt.Sprintf("Nothing to apply. Autoscale policies for instance %q is match with the file %q", instance, file))
		return nil
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	err = kubectl.JsonPatch(ik, instance, string(patchBytes))
	if err != nil {
		spinner.Stop(false)
		return err
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully applied autoscale policies for instance %q", instance))
	return nil
}

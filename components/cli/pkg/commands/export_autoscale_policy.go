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
	"regexp"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/policies"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunExportAutoscalePoliciesOfCell(instance string, outputfile string) error {
	aPolicy := policies.CellPolicy{
		Rules: []policies.Rule{},
		Type:  policies.PolicyTypeAutoscale,
	}
	// set autoscale policy for gw
	gwSpinner := util.StartNewSpinner("Retrieving Gateway autoscale policy")
	err := populateGatewayPolicy(instance, &aPolicy)
	if err != nil {
		if err, ok := err.(policies.PolicyNotFoundError); ok {
			// since this is a not found error, continue and see if components included in the cell has autoscale policies.
			gwSpinner.Stop(false)
			// log.Printf("Gateway autoscaling policy for instance %s not found \n", instance)
		} else {
			gwSpinner.Stop(false)
			return err
		}
	}
	gwSpinner.Stop(true)
	// TODO: get autoscale policy for STS
	// set autoscale policies for all components
	aCell, err := kubectl.GetCell(instance)
	if err != nil {
		return err
	}
	for _, csvc := range aCell.CellSpec.ComponentTemplates {
		currentCompSpinner := util.StartNewSpinner(fmt.Sprintf("Retrieving autoscale policy for component %s", csvc.Metadata.Name))
		err := populateComponentPolicy(instance, csvc.Metadata.Name, &aPolicy)
		if err != nil {
			if err, ok := err.(policies.PolicyNotFoundError); ok {
				// since this is a not found error, continue and see if other components included in the cell has autoscale policies.
				currentCompSpinner.Stop(false)
				// log.Printf("Autoscaling policy for instance %s, component %s not found \n", instance, csvc.Metadata.Name)
			} else {
				currentCompSpinner.Stop(false)
				return err
			}
		}
		currentCompSpinner.Stop(true)
	}
	// check if any autoscaling policies are returned, else return
	if len(aPolicy.Rules) == 0 {
		return fmt.Errorf("No autoscale policies found for instance %s ", instance)
	}

	polExportSpinner := util.StartNewSpinner("Combining and exporting autoscale policies")
	bytes, err := json.Marshal(aPolicy)
	if err != nil {
		polExportSpinner.Stop(false)
		return err
	}
	yamlBytes, err := yaml.JSONToYAML(bytes)
	if err != nil {
		polExportSpinner.Stop(false)
		return err
	}
	// write to a file
	file := outputfile
	if file == "" {
		file = filepath.Join("./", instance+"-autoscalepolicy.yaml")
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

func getScalePolicySpec(name string) (*kubectl.AutoscalePolicySpec, error) {
	autoscalePolicy, err := kubectl.GetAutoscalePolicy(name)
	if err != nil {
		if match, matchErr := regexp.MatchString(policies.BuildAutoscalePolicyNonExistErrorMatcher(name), err.Error()); matchErr == nil {
			if match {
				// return a specific not found error
				return nil, policies.PolicyNotFoundError{}
			} else {
				return nil, err
			}
		} else {
			return nil, matchErr
		}
	}
	return &autoscalePolicy.Spec, nil
}

func populateGatewayPolicy(instance string, policy *policies.CellPolicy) error {
	gwPolSpec, err := getScalePolicySpec(policies.GetGatewayAutoscalePolicyName(instance))
	if err != nil {
		return err
	}
	gwRule := policies.Rule{
		Overridable: gwPolSpec.Overridable,
		Target: policies.Target{
			Type: policies.CellGatewayTargetType,
		},
		Policy: policies.Policy{
			MinReplicas: gwPolSpec.Policy.MinReplicas,
			MaxReplicas: gwPolSpec.Policy.MaxReplicas,
			Metrics:     getMetricsFromPolicySpec(gwPolSpec),
		},
	}
	policy.Rules = append(policy.Rules, gwRule)
	return nil
}

func populateComponentPolicy(instance string, component string, policy *policies.CellPolicy) error {
	compPolSpec, err := getScalePolicySpec(policies.GetComponentAutoscalePolicyName(instance, component))
	if err != nil {
		return err
	}
	compRule := policies.Rule{
		Overridable: compPolSpec.Overridable,
		Target: policies.Target{
			Type: policies.CellComponentTargetType,
			Name: component,
		},
		Policy: policies.Policy{
			MinReplicas: compPolSpec.Policy.MinReplicas,
			MaxReplicas: compPolSpec.Policy.MaxReplicas,
			Metrics:     getMetricsFromPolicySpec(compPolSpec),
		},
	}
	policy.Rules = append(policy.Rules, compRule)
	return nil
}

func getMetricsFromPolicySpec(policySpec *kubectl.AutoscalePolicySpec) []policies.Metric {
	var specMetrics []policies.Metric
	for _, metric := range policySpec.Policy.Metrics {
		specMetric := policies.Metric{
			Type: metric.Type,
			Resource: policies.Resource{
				Name:                     metric.Resource.Name,
				TargetAverageUtilization: metric.Resource.TargetAverageUtilization,
			},
		}
		specMetrics = append(specMetrics, specMetric)
	}
	return specMetrics
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

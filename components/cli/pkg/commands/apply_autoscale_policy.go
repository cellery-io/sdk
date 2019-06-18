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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/policies"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunApplyAutoscalePolicies(file string, instance string) error {
	spinner := util.StartNewSpinner("Applying autoscale policies")
	// check if the cell exists
	_, err := kubectl.GetCell(instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// read the file
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	policy := policies.CellPolicy{}
	err = yaml.Unmarshal(contents, &policy)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	k8sAutoscalePolicies, err := generateCellAutoscalePolicies(&policy, instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// write policies to file one by one, else kubectl apply issues can occur
	policiesFile := filepath.Join("./", fmt.Sprintf("%s-autoscale-policies.yaml", instance))
	defer func() {
		_ = os.Remove(policiesFile)
	}()
	err = writeCellAutoscalePoliciesToFile(policiesFile, k8sAutoscalePolicies)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// kubectl apply
	err = kubectl.ApplyFile(policiesFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully applied autoscale policies for instance %s", instance))
	return nil
}

func RunApplyAutoscalePolicyToComponents(file string, instance string, components string) error {
	spinner := util.StartNewSpinner("Applying autoscale policies")
	// components can be separated with a comma, split
	componentsArr := strings.Split(components, ",")
	// check if the cell exists
	cellInst, err := kubectl.GetCell(instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// check whether the specified components exist in the retrieved cell instance
	if !checkIfComponentsExistInCellInstance(componentsArr, cellInst) {
		return fmt.Errorf("components specified does not match with the given cell instance")
	}
	// read the file
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	policy := policies.CellPolicy{}
	err = yaml.Unmarshal(contents, &policy)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	k8sAutoscalePolicies, err := generataCellAutoscalePoliciesForComponents(&policy, instance, componentsArr)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// write policies to file one by one, else kubectl apply issues can occur
	policiesFile := filepath.Join("./", fmt.Sprintf("%s-autoscale-policies.yaml", instance))
	err = writeCellAutoscalePoliciesToFile(policiesFile, k8sAutoscalePolicies)
	defer func() {
		_ = os.Remove(policiesFile)
	}()
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// kubectl apply
	err = kubectl.ApplyFile(policiesFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully applied autoscale policy for instance %s, components %s", instance, components))
	return nil
}

func RunApplyAutoscalePolicyToCellGw(file string, instance string) error {
	spinner := util.StartNewSpinner("Applying autoscale policies")
	// check if the cell exists
	_, err := kubectl.GetCell(instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// read the file
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	policy := policies.CellPolicy{}
	err = yaml.Unmarshal(contents, &policy)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	k8sAutoscalePolicies, err := generataCellAutoscalePolicyForGw(&policy, instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// write policies to file one by one, else kubectl apply issues can occur
	policiesFile := filepath.Join("./", fmt.Sprintf("%s-autoscale-policies.yaml", instance))
	err = writeCellAutoscalePoliciesToFile(policiesFile, &[]util.AutoscalePolicy{k8sAutoscalePolicies})
	defer func() {
		_ = os.Remove(policiesFile)
	}()
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// kubectl apply
	err = kubectl.ApplyFile(policiesFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully applied autoscale policy for instance %s gateway", instance))
	return nil
}

func generateCellAutoscalePolicies(policy *policies.CellPolicy, instance string) (*[]util.AutoscalePolicy, error) {
	var k8sAutoscalePolicies []util.AutoscalePolicy
	for _, rule := range policy.Rules {
		var k8sAutoscalePolicy util.AutoscalePolicy
		switch polTargetType := rule.Target.Type; polTargetType {
		case policies.CellComponentTargetType:
			k8sAutoscalePolicy = createK8sAutoscalePolicy(rule,
				policies.GetComponentAutoscalePolicyName(instance, rule.Target.Name),
				policies.GetTargetComponentDeploymentName(instance, rule.Target.Name))
		case policies.CellGatewayTargetType:
			k8sAutoscalePolicy = createK8sAutoscalePolicy(rule,
				policies.GetGatewayAutoscalePolicyName(instance),
				policies.GetTargetGatewayeploymentName(instance))
		default:
			return nil, fmt.Errorf("wrong target type %s in the Autoscale policy", polTargetType)
		}
		k8sAutoscalePolicies = append(k8sAutoscalePolicies, k8sAutoscalePolicy)
	}
	return &k8sAutoscalePolicies, nil
}

func generataCellAutoscalePoliciesForComponents(policy *policies.CellPolicy, instance string, components []string) (*[]util.AutoscalePolicy, error) {
	var k8sAutoscalePolicies []util.AutoscalePolicy
	// pick the first rule. If there are more than one, rest will be ignored.
	rule := policy.Rules[0]
	for _, component := range components {
		k8sAutoscalePolicy := createK8sAutoscalePolicy(rule,
			policies.GetComponentAutoscalePolicyName(instance, component),
			policies.GetTargetComponentDeploymentName(instance, component))
		k8sAutoscalePolicies = append(k8sAutoscalePolicies, k8sAutoscalePolicy)
	}
	return &k8sAutoscalePolicies, nil
}

func generataCellAutoscalePolicyForGw(policy *policies.CellPolicy, instance string) (util.AutoscalePolicy, error) {
	// pick the first rule. If there are more than one, rest will be ignored.
	rule := policy.Rules[0]
	return createK8sAutoscalePolicy(rule, policies.GetGatewayAutoscalePolicyName(instance), policies.GetTargetGatewayeploymentName(instance)), nil
}

func createK8sAutoscalePolicy(rule policies.Rule, name string, scaleTargetRefName string) util.AutoscalePolicy {
	return util.AutoscalePolicy{
		Kind:       policies.CelleryAutoscalePolicyKind,
		APIVersion: policies.CelleryApiVersion,
		Metadata: util.AutoscalePolicyMetadata{
			Name: name,
		},
		Spec: util.AutoscalePolicySpec{
			Overridable: rule.Overridable,
			Policy: util.Policy{
				MinReplicas: rule.Policy.MinReplicas,
				MaxReplicas: rule.Policy.MaxReplicas,
				ScaleTargetRef: util.ScaleTargetRef{
					ApiVersion: policies.K8sScaleTargetApiVersion,
					Kind:       policies.K8sScaleTargetKind,
					Name:       scaleTargetRefName,
				},
				Metrics: getK8sAutoscalePolicyMetrics(rule.Policy.Metrics),
			},
		},
	}
}

func getK8sAutoscalePolicyMetrics(metrics []policies.Metric) []util.Metric {
	var k8sMetrics []util.Metric
	for _, metric := range metrics {
		k8sMetric := util.Metric{
			Type: metric.Type,
			Resource: util.Resource{
				Name:                     metric.Resource.Name,
				TargetAverageUtilization: metric.Resource.TargetAverageUtilization,
			},
		}
		k8sMetrics = append(k8sMetrics, k8sMetric)
	}
	return k8sMetrics
}

func checkIfComponentsExistInCellInstance(components []string, cell kubectl.Cell) bool {
	// for each component provided by user, check if they are there in the cell
	for _, component := range components {
		matchFound := false
		for _, cellComponent := range cell.CellSpec.ComponentTemplates {
			if strings.TrimSpace(component) == cellComponent.Metadata.Name {
				matchFound = true
				break
			}
		}
		if !matchFound {
			return false
		}
	}
	return true
}

func writeCellAutoscalePoliciesToFile(policiesFile string, k8sAutoscalePolicies *[]util.AutoscalePolicy) error {
	f, err := os.OpenFile(policiesFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	for _, k8sPolicy := range *k8sAutoscalePolicies {
		yamlContent, err := yaml.Marshal(k8sPolicy)
		if err != nil {
			return err
		}
		if _, err := f.Write(yamlContent); err != nil {
			return err
		}
		if _, err := f.Write([]byte("---\n")); err != nil {
			return err
		}
	}
	return nil
}

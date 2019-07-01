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
	"regexp"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/policies"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunApplyAutoscalePolicies(file string, instance string) error {
	spinner := util.StartNewSpinner("Preparing to apply autoscale policies")
	// check if the cell exists
	cellInst, err := kubectl.GetCell(instance)
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
	// get the list of overridable components (which has spec.overridable = `true`) and whether the gateway
	// policy is overridable. If a particular component does not have a autoscale policy, its considered as overridable.
	overridableComponents, err := getOverridableComponents(&cellInst, instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	isGatewayOverridable, err := isGatewayPolicyOverridable(instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	k8sAutoscalePolicies, err := generateCellAutoscalePolicies(policy, instance, overridableComponents, isGatewayOverridable)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	if len(k8sAutoscalePolicies) == 0 {
		spinner.Stop(false)
		return fmt.Errorf(fmt.Sprintf("No autoscale policies can be updated in cell instance %s", instance))
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
	spinner := util.StartNewSpinner("Preparing to apply autoscale policies")
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
		spinner.Stop(false)
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
	// get the list of overridable components (which has spec.overridable = `true`) and whether the gateway
	// policy is overridable. If a particular component does not have a autoscale policy, its considered as overridable.
	overridableComponents, err := getOverridableComponents(&cellInst, instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	k8sAutoscalePolicies, err := generataCellAutoscalePoliciesForComponents(&policy, instance, componentsArr, overridableComponents)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	if len(k8sAutoscalePolicies) == 0 {
		spinner.Stop(false)
		return fmt.Errorf(fmt.Sprintf("No autoscale policies can be updated in cell instance %s", instance))
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
	spinner := util.StartNewSpinner("Preparing to apply autoscale policies")
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
	isGatewayOverridable, err := isGatewayPolicyOverridable(instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	if !isGatewayOverridable {
		// can't update the gateway!
		spinner.Stop(false)
		return fmt.Errorf(fmt.Sprintf("Gateway of cell instance %s is not overridable", instance))
	}
	k8sAutoscalePolicies, err := generataCellAutoscalePolicyForGw(&policy, instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// write policies to file one by one, else kubectl apply issues can occur
	policiesFile := filepath.Join("./", fmt.Sprintf("%s-autoscale-policies.yaml", instance))
	err = writeCellAutoscalePoliciesToFile(policiesFile, []*kubectl.AutoscalePolicy{&k8sAutoscalePolicies})
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

func generateCellAutoscalePolicies(policy policies.CellPolicy, instance string, overridableComponents []string, gatewayOverridable bool) ([]*kubectl.AutoscalePolicy, error) {
	var k8sAutoscalePolicies []*kubectl.AutoscalePolicy
	for _, rule := range policy.Rules {
		switch polTargetType := rule.Target.Type; polTargetType {
		case policies.CellComponentTargetType:
			if isComponentPolicyOverridable(overridableComponents, rule.Target.Name) {
				k8sAutoscalePolicy, err := createK8sAutoscalePolicy(rule,
					policies.GetComponentAutoscalePolicyName(instance, rule.Target.Name),
					policies.GetTargetComponentDeploymentName(instance, rule.Target.Name))
				if err != nil {
					return k8sAutoscalePolicies, err
				}
				k8sAutoscalePolicies = append(k8sAutoscalePolicies, &k8sAutoscalePolicy)
			}
		case policies.CellGatewayTargetType:
			if gatewayOverridable {
				k8sAutoscalePolicy, err := createK8sAutoscalePolicy(rule,
					policies.GetGatewayAutoscalePolicyName(instance),
					policies.GetTargetGatewayeploymentName(instance))
				if err != nil {
					return k8sAutoscalePolicies, err
				}
				k8sAutoscalePolicies = append(k8sAutoscalePolicies, &k8sAutoscalePolicy)
			}
		default:
			return nil, fmt.Errorf("wrong target type %s in the Autoscale policy", polTargetType)
		}
	}
	return k8sAutoscalePolicies, nil
}

func isComponentPolicyOverridable(overridableComponents []string, targetComponent string) bool {
	for _, component := range overridableComponents {
		if targetComponent == component {
			return true
		}
	}
	return false
}

func generataCellAutoscalePoliciesForComponents(policy *policies.CellPolicy, instance string, components []string, overridableComponents []string) ([]*kubectl.AutoscalePolicy, error) {
	var k8sAutoscalePolicies []*kubectl.AutoscalePolicy
	// pick the first rule. If there are more than one, rest will be ignored.
	rule := policy.Rules[0]
	for _, component := range components {
		if isComponentPolicyOverridable(overridableComponents, component) {
			k8sAutoscalePolicy, err := createK8sAutoscalePolicy(rule,
				policies.GetComponentAutoscalePolicyName(instance, component),
				policies.GetTargetComponentDeploymentName(instance, component))
			if err != nil {
				return k8sAutoscalePolicies, err
			}
			k8sAutoscalePolicies = append(k8sAutoscalePolicies, &k8sAutoscalePolicy)
		}
	}
	return k8sAutoscalePolicies, nil
}

func generataCellAutoscalePolicyForGw(policy *policies.CellPolicy, instance string) (kubectl.AutoscalePolicy, error) {
	// pick the first rule. If there are more than one, rest will be ignored.
	rule := policy.Rules[0]
	return createK8sAutoscalePolicy(rule, policies.GetGatewayAutoscalePolicyName(instance), policies.GetTargetGatewayeploymentName(instance))
}

func createK8sAutoscalePolicy(rule policies.Rule, name string, scaleTargetRefName string) (kubectl.AutoscalePolicy, error) {
	metrics, err := getK8sAutoscalePolicyMetrics(rule.Policy.Metrics)
	if err != nil {
		return kubectl.AutoscalePolicy{}, err
	}
	return kubectl.AutoscalePolicy{
		Kind:       policies.CelleryAutoscalePolicyKind,
		APIVersion: policies.CelleryApiVersion,
		Metadata: kubectl.AutoscalePolicyMetadata{
			Name: name,
		},
		Spec: kubectl.AutoscalePolicySpec{
			Overridable: rule.Overridable,
			Policy: kubectl.Policy{
				MinReplicas: rule.Policy.MinReplicas,
				MaxReplicas: rule.Policy.MaxReplicas,
				ScaleTargetRef: kubectl.ScaleTargetRef{
					ApiVersion: policies.K8sScaleTargetApiVersion,
					Kind:       policies.K8sScaleTargetKind,
					Name:       scaleTargetRefName,
				},
				Metrics: metrics,
			},
		},
	}, nil
}

func getK8sAutoscalePolicyMetrics(metrics []policies.Metric) ([]kubectl.Metric, error) {
	var k8sMetrics []kubectl.Metric
	for _, metric := range metrics {
		res, err := getResourceFromMetric(&metric)
		if err != nil {
			return nil, err
		}
		k8sMetric := kubectl.Metric{
			Type:     metric.Type,
			Resource: *res,
		}
		k8sMetrics = append(k8sMetrics, k8sMetric)
	}
	return k8sMetrics, nil
}

func getResourceFromMetric(metric *policies.Metric) (*kubectl.Resource, error) {
	if metric.Resource.TargetAverageUtilization > 0 && metric.Resource.TargetAverageValue != "" {
		// Cannot provide both
		return nil, fmt.Errorf("TargetAverageUtilization and TargetAverageValue cannot be applied simultaneously in metric %s", metric.Resource.Name)
	}
	if metric.Resource.TargetAverageUtilization > 0 {
		return &kubectl.Resource{
			Name:                     metric.Resource.Name,
			TargetAverageUtilization: metric.Resource.TargetAverageUtilization,
		}, nil
	} else if metric.Resource.TargetAverageValue != "" {
		return &kubectl.Resource{
			Name:               metric.Resource.Name,
			TargetAverageValue: metric.Resource.TargetAverageValue,
		}, nil
	} else {
		return nil, fmt.Errorf("Either TargetAverageUtilization or TargetAverageValue should be provided for metric %s", metric.Resource.Name)
	}
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

func getOverridableComponents(cell *kubectl.Cell, instance string) ([]string, error) {
	var applicableComponents []string
	for _, component := range cell.CellSpec.ComponentTemplates {
		ap, err := kubectl.GetAutoscalePolicy(policies.GetComponentAutoscalePolicyName(instance, component.Metadata.Name))
		if err != nil {
			matches, err := regexp.MatchString(policies.BuildAutoscalePolicyNonExistErrorMatcher(policies.GetComponentAutoscalePolicyName(instance, component.Metadata.Name)), err.Error())
			if err != nil {
				return nil, err
			}
			if matches {
				// no autoscaling policy for this component, hence can be considered applicable
				applicableComponents = append(applicableComponents, component.Metadata.Name)
			} else {
				return nil, err
			}
		} else {
			if ap.Spec.Overridable {
				applicableComponents = append(applicableComponents, component.Metadata.Name)
			}
		}
	}
	return applicableComponents, nil
}

func isGatewayPolicyOverridable(instance string) (bool, error) {
	ap, err := kubectl.GetAutoscalePolicy(policies.GetGatewayAutoscalePolicyName(instance))
	if err != nil {
		if matches, matchErr := regexp.MatchString(policies.BuildAutoscalePolicyNonExistErrorMatcher(policies.GetGatewayAutoscalePolicyName(instance)), err.Error()); matchErr == nil {
			if matches {
				return true, nil
			} else {
				return false, err
			}
		} else {
			return false, matchErr
		}
	}
	return ap.Spec.Overridable, nil
}

func writeCellAutoscalePoliciesToFile(policiesFile string, k8sAutoscalePolicies []*kubectl.AutoscalePolicy) error {
	f, err := os.OpenFile(policiesFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	for _, k8sPolicy := range k8sAutoscalePolicies {
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

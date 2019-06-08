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

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/pkg/cell"
	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/policies"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunApplyAutoscalePolicies(file string, instance string, components string) error {
	spinner := util.StartNewSpinner("Applying autoscale policies")
	// check if the cell exists
	_, err := cell.GetInstance(instance)
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
	f, err := os.OpenFile(policiesFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		spinner.Stop(false)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(policiesFile)
	}()
	for _, k8sPolicy := range *k8sAutoscalePolicies {
		yamlContent, err := yaml.Marshal(k8sPolicy)
		if err != nil {
			spinner.Stop(false)
			return err
		}
		if _, err := f.Write(yamlContent); err != nil {
			spinner.Stop(false)
			return err
		}
		if _, err := f.Write([]byte("---\n")); err != nil {
			spinner.Stop(false)
			return err
		}
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

func generateCellAutoscalePolicies(policy *policies.CellPolicy, instance string) (*[]util.AutoscalePolicy, error) {
	k8sAutoscalePolicies := []util.AutoscalePolicy{}
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
			return nil, fmt.Errorf("Wrong target type %s in the Autoscale policy", polTargetType)
		}
		k8sAutoscalePolicies = append(k8sAutoscalePolicies, k8sAutoscalePolicy)
	}
	return &k8sAutoscalePolicies, nil
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
	k8sMetrics := []util.Metric{}
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

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

package policies

import "fmt"

const K8sScaleTargetApiVersion = "apps/v1"
const K8sScaleTargetKind = "Deployment"
const CelleryApiVersion = "mesh.cellery.io/v1alpha1"
const CelleryAutoscalePolicyKind = "AutoscalePolicy"
const CellComponentTargetType = "component"
const CellGatewayTargetType = "gateway"

func GetComponentAutoscalePolicyName(instance string, component string) string {
	return fmt.Sprintf("%s--%s-autoscalepolicy", instance, component)
}

func GetGatewayAutoscalePolicyName(instance string) string {
	return fmt.Sprintf("%s--gateway-autoscalepolicy", instance)
}

func GetTargetComponentDeploymentName(instance string, component string) string {
	return fmt.Sprintf("%s--%s-deployment", instance, component)
}

func GetTargetGatewayeploymentName(instance string) string {
	return fmt.Sprintf("%s--gateway-deployment", instance)
}

func BuildAutoscalePolicyNonExistErrorMatcher(name string) string {
	return fmt.Sprintf("autoscalepolicies.mesh.cellery.io(\\s)?\"%s\"(\\s)?not found", name)
}

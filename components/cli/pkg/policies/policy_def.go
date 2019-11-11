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

package policies

const PolicyTypeAutoscale = "AutoscalePolicy"

type CellPolicy struct {
	Type  string `json:"type"`
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Overridable bool   `json:"overridable"`
	Target      Target `json:"target,omitempty"`
	Policy      Policy `json:"policy,omitempty"`
}

type Target struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

type Policy struct {
	MinReplicas string   `json:"minReplicas"`
	MaxReplicas int      `json:"maxReplicas"`
	Metrics     []Metric `json:"metrics,omitempty"`
}

type Metric struct {
	Type     string   `json:"type,omitempty"`
	Resource Resource `json:"resource,omitempty"`
}

type Resource struct {
	Name                     string `json:"name"`
	TargetAverageUtilization int    `json:"targetAverageUtilization,omitempty"`
	TargetAverageValue       string `json:"targetAverageValue,omitempty"`
}

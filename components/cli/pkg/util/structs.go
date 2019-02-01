/*
 * Copyright (c) 2018 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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

package util

type Cell struct {
	CellMetaData CellMetaData `json:"metadata"`
	CellStatus   CellStatus   `json:"status"`
}

type CellMetaData struct {
	CreationTimestamp string `json:"creationTimestamp"`
}

type CellStatus struct {
	Status string `json:"status"`
}

type CellPods struct {
	Items []Pod `json:"items"`
}

type Pod struct {
	MetaData  PodMetaData `json:"metadata"`
	PodStatus PodStatus   `json:"status"`
}

type PodMetaData struct {
	Name string `json:"name"`
}

type PodStatus struct {
	Phase      string         `json:"phase"`
	StartTime  string         `json:"startTime"`
	Conditions []PodCondition `json:"conditions"`
}

type PodCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	LastTransitionTime string `json:"lastTransitionTime"`
}

type Service struct {
	Items []ServiceItem `json:"items"`
}

type ServiceItem struct {
	Metadata ServiceMetaData `json:"metadata"`
	Spec     ServiceSpec     `json:"spec"`
}

type ServiceMetaData struct {
	Name string `json:"name"`
}

type ServiceSpec struct {
	Ports []ServicePort `json:"ports"`
}

type ServicePort struct {
	Port int `json:"port"`
}

type Gateway struct {
	GatewaySpec GatewaySpec `json:"spec"`
}

type GatewaySpec struct {
	Apis []GatewayApi `json:"apis"`
}

type GatewayApi struct {
	Backend     string              `json:"backend"`
	Context     string              `json:"context"`
	Definitions []GatewayDefinition `json:"definitions"`
}

type GatewayDefinition struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

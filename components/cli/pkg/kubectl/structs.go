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

package kubectl

type Node struct {
	Items []NodeItem `json:"items"`
}

type NodeItem struct {
	Metadata NodeMetaData `json:"metadata"`
	Status   NodeStatus   `json:"status"`
}

type NodeMetaData struct {
	Name string `json:"name"`
}

type NodeStatus struct {
	NodeInfo NodeInfo `json:"nodeInfo"`
}

type NodeInfo struct {
	KubeletVersion string `json:"kubeletVersion"`
}

type Cells struct {
	Items []Cell `json:"items"`
}

type Cell struct {
	Kind         string       `json:"kind"`
	APIVersion   string       `json:"apiVersion"`
	CellMetaData CellMetaData `json:"metadata"`
	CellSpec     CellSpec     `json:"spec"`
	CellStatus   CellStatus   `json:"status"`
}

type CellMetaData struct {
	CreationTimestamp string          `json:"creationTimestamp"`
	Annotations       CellAnnotations `json:"annotations"`
	Name              string          `json:"name"`
}

type CellSpec struct {
	ComponentTemplates []ComponentTemplate `json:"servicesTemplates"`
	GateWayTemplate    Gateway             `json:"gatewayTemplate,omitempty"`
}

type ComponentTemplate struct {
	Metadata ComponentTemplateMetadata `json:"metadata"`
	Spec     ComponentTemplateSpec     `json:"spec"`
}

type ComponentTemplateMetadata struct {
	Name string `json:"name"`
}

type ComponentTemplateSpec struct {
	Container ContainerTemplate `json:"container"`
}

type ContainerTemplate struct {
	Env   []Env  `json:"env,omitempty"`
	Image string `json:"image"`
	Ports []Port `json:"ports"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Port struct {
	ContainerPort int `json:"containerPort"`
}

type Gateway struct {
	GatewaySpec GatewaySpec `json:"spec,omitempty"`
}

type GatewaySpec struct {
	HttpApis []GatewayHttpApi `json:"http,omitempty"`
	TcpApis  []GatewayTcpApi  `json:"tcp,omitempty"`
	GrpcApis []GatewayGrpcApi `json:"grpc,omitempty"`
}

type GatewayHttpApi struct {
	Backend     string              `json:"backend"`
	Context     string              `json:"context"`
	Definitions []GatewayDefinition `json:"definitions"`
	Global      bool                `json:"global"`
	Vhost       string              `json:"vhost"`
}

type GatewayTcpApi struct {
	Backend     string              `json:"backend"`
	Context     string              `json:"context"`
	Definitions []GatewayDefinition `json:"definitions"`
	Global      bool                `json:"global"`
	Vhost       string              `json:"vhost"`
}

type GatewayGrpcApi struct {
	Backend     string              `json:"backend"`
	Context     string              `json:"context"`
	Definitions []GatewayDefinition `json:"definitions"`
	Global      bool                `json:"global"`
	Vhost       string              `json:"vhost"`
}

type GatewayDefinition struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type CellAnnotations struct {
	Organization string `json:"mesh.cellery.io/cell-image-org"`
	Name         string `json:"mesh.cellery.io/cell-image-name"`
	Version      string `json:"mesh.cellery.io/cell-image-version"`
}

type CellStatus struct {
	Status       string `json:"status"`
	Gateway      string `json:"gatewayHostname"`
	ServiceCount int    `json:"serviceCount"`
}

type Pods struct {
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

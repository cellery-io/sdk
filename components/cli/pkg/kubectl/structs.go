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

type Composites struct {
	Items []Composite `json:"items"`
}

type Cell struct {
	Kind         string      `json:"kind"`
	APIVersion   string      `json:"apiVersion"`
	CellMetaData K8SMetaData `json:"metadata"`
	CellSpec     CellSpec    `json:"spec"`
	CellStatus   CellStatus  `json:"status"`
}

type Composite struct {
	Kind              string          `json:"kind"`
	APIVersion        string          `json:"apiVersion"`
	CompositeMetaData K8SMetaData     `json:"metadata"`
	CompositeSpec     CompositeSpec   `json:"spec"`
	CompositeStatus   CompositeStatus `json:"status"`
}

type K8SMetaData struct {
	CreationTimestamp string          `json:"creationTimestamp"`
	Annotations       CellAnnotations `json:"annotations"`
	Name              string          `json:"name"`
}

type CellSpec struct {
	ComponentTemplates []ComponentTemplate `json:"components"`
	GateWayTemplate    Gateway             `json:"gateway,omitempty"`
}

type CompositeSpec struct {
	ComponentTemplates []ComponentTemplate `json:"components"`
}

type ComponentTemplate struct {
	Metadata ComponentTemplateMetadata `json:"metadata"`
	Spec     ComponentTemplateSpec     `json:"spec"`
}

type ComponentTemplateMetadata struct {
	Name string `json:"name"`
}

type PodTemplate struct {
	Containers []ContainerTemplate `json:"containers"`
}

type ComponentTemplateSpec struct {
	PodTemplate PodTemplate `json:"template"`
	Ports       []Port      `json:"ports"`
}

type ContainerTemplate struct {
	Env   []Env  `json:"env,omitempty"`
	Image string `json:"image"`
	Name  string `json:"name"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Port struct {
	Name            string `json:"name,omitempty"`
	Protocol        string `json:"protocol"`
	Port            int32  `json:"port"`
	TargetContainer string `json:"targetContainer,omitempty"`
	TargetPort      int32  `json:"targetPort"`
}

type Gateway struct {
	GatewaySpec GatewaySpec `json:"spec,omitempty"`
}

type GatewaySpec struct {
	Ingress Ingress `json:"ingress,omitempty"`
}

type Ingress struct {
	Extensions interface{}      `json:"extensions,omitempty"`
	HttpApis   []GatewayHttpApi `json:"http,omitempty"`
	TcpApis    []GatewayTcpApi  `json:"tcp,omitempty"`
	GrpcApis   []GatewayGrpcApi `json:"grpc,omitempty"`
}

type GatewayHttpApi struct {
	Context      string          `json:"context"`
	Version      string          `json:"version"`
	Definitions  []APIDefinition `json:"definitions"`
	Global       bool            `json:"global"`
	Authenticate bool            `json:"authenticate"`
	Port         uint32          `json:"port"`
	Destination  Destination     `json:"destination,omitempty"`
	ZeroScale    bool            `json:"zeroScale,omitempty"`
}

type APIDefinition struct {
	Path   string `json:"path"`
	Method string `json:"method"`
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
	Organization                        string `json:"mesh.cellery.io/cell-image-org"`
	Name                                string `json:"mesh.cellery.io/cell-image-name"`
	Version                             string `json:"mesh.cellery.io/cell-image-version"`
	Dependencies                        string `json:"mesh.cellery.io/cell-dependencies"`
	ApiVersion                          string `json:"mesh.cellery.io/apiVersion"`
	OriginalDependencyComponentServices string `json:"mesh.cellery.io/original-component-svcs,omitempty"`
}

type CellStatus struct {
	Status       string `json:"status"`
	Gateway      string `json:"gatewayServiceName"`
	ServiceCount int    `json:"componentCount"`
}

type CompositeStatus struct {
	Status       string `json:"status"`
	ServiceCount int    `json:"componentCount"`
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

type Services struct {
	Items []Service `json:"items"`
}

type Service struct {
	Metadata ServiceMetaData `json:"metadata"`
	Spec     ServiceSpec     `json:"spec"`
}

type ServiceMetaData struct {
	Name string `json:"name"`
}

type ServiceSpec struct {
	Ports []ServicePort `json:"ports"`
	ExternalIPs []string `json:"externalIPs"`
}

type ServicePort struct {
	Port int `json:"port"`
}

type VirtualService struct {
	Kind       string     `json:"kind"`
	APIVersion string     `json:"apiVersion"`
	VsMetaData VsMetaData `json:"metadata"`
	VsSpec     VsSpec     `json:"spec"`
}

type VsMetaData struct {
	Name string `json:"name"`
}

type VsSpec struct {
	Hosts []string `json:"hosts"`
	HTTP  []HTTP   `json:"http,omitempty"`
	TCP   []TCP    `json:"tcp,omitempty"`
}

type TCP struct {
	Match []TCPMatch `json:"match"`
	Route []TCPRoute `json:"route"`
}

type TCPMatch struct {
	Port         int               `json:"port"`
	SourceLabels map[string]string `json:"sourceLabels"`
}

type TCPRoute struct {
	Destination TCPDestination `json:"destination"`
	Weight      int            `json:"weight,omitempty"`
}

type TCPDestination struct {
	Host string  `json:"host"`
	Port TCPPort `json:"port"`
}

type TCPPort struct {
	Number int `json:"number"`
}

type HTTP struct {
	Match []HTTPMatch `json:"match"`
	Route []HTTPRoute `json:"route"`
}

type HTTPMatch struct {
	Authority    Authority               `json:"authority"`
	SourceLabels map[string]string       `json:"sourceLabels"`
	Headers      map[string]*StringMatch `json:"headers,omitempty"`
}

type StringMatch struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regex  string `json:"regex,omitempty"`
}

type Authority struct {
	Regex string `json:"regex"`
}

type HTTPRoute struct {
	Destination Destination `json:"destination"`
	Weight      int         `json:"weight,omitempty"`
}

type Destination struct {
	Host string `json:"host"`
}

type AutoscalePolicy struct {
	Kind       string                  `json:"kind"`
	APIVersion string                  `json:"apiVersion"`
	Metadata   AutoscalePolicyMetadata `json:"metadata"`
	Spec       AutoscalePolicySpec     `json:"spec"`
}

type AutoscalePolicyMetadata struct {
	Name string `json:"name"`
}

type AutoscalePolicySpec struct {
	Overridable bool   `json:"overridable"`
	Policy      Policy `json:"policy"`
}

type Policy struct {
	MinReplicas    string         `json:"minReplicas"`
	MaxReplicas    int            `json:"maxReplicas"`
	ScaleTargetRef ScaleTargetRef `json:"scaleTargetRef"`
	Metrics        []Metric       `json:"metrics"`
}

type Metric struct {
	Type     string   `json:"type"`
	Resource Resource `json:"resource"`
}

type ScaleTargetRef struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type Resource struct {
	Name                     string `json:"name"`
	TargetAverageUtilization int    `json:"targetAverageUtilization,omitempty"`
	TargetAverageValue       string `json:"targetAverageValue,omitempty"`
}

type ScaleResource struct {
	Spec struct {
		Components []struct {
			Metadata struct {
				Name string `json:"name,omitempty"`
			} `json:"metadata,omitempty"`
			Spec struct {
				ScalingPolicy interface{} `json:"scalingPolicy,omitempty"`
			} `json:"spec,omitempty"`
		} `json:"components,omitempty"`
		Gateway struct {
			Metadata struct {
				Name string `json:"name,omitempty"`
			} `json:"metadata,omitempty"`
			Spec struct {
				ScalingPolicy interface{} `json:"scalingPolicy,omitempty"`
			} `json:"spec,omitempty"`
		} `json:"gateway,omitempty"`
	} `json:"spec,omitempty"`
}

type AutoScalingPolicy struct {
	Components []ComponentScalePolicy `json:"components,omitempty"`
	Gateway    GwScalePolicy          `json:"gateway,omitempty"`
}

type ScalingPolicy struct {
	Hpa struct {
		Overridable *bool `json:"overridable,omitempty"`
	} `json:"hpa,omitempty"`
}

type ComponentScalePolicy struct {
	Name          string      `json:"name,omitempty"`
	ScalingPolicy interface{} `json:"scalingPolicy,omitempty"`
}

type GwScalePolicy struct {
	ScalingPolicy interface{} `json:"scalingPolicy,omitempty"`
}

type InstanceKind string

const (
	InstanceKindCell      InstanceKind = "cells.mesh.cellery.io"
	InstanceKindComposite InstanceKind = "composites.mesh.cellery.io"
)

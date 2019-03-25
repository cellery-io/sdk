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

import (
	"sync"

	"github.com/tj/go-spin"
)

type CellImage struct {
	Registry     string
	Organization string
	ImageName    string
	ImageVersion string
}

type CellList struct {
	Items []Cell `json:"items"`
}

type Cell struct {
	CellMetaData CellMetaData `json:"metadata"`
	CellSpec     CellSpec     `json:"spec"`
	CellStatus   CellStatus   `json:"status"`
}

type CellMetaData struct {
	CreationTimestamp string          `json:"creationTimestamp"`
	Annotations       CellAnnotations `json:"annotations"`
	Name              string          `json:"name"`
}

type CellAnnotations struct {
	Organization string `json:"mesh.cellery.io/cell-image-org"`
	Name         string `json:"mesh.cellery.io/cell-image-name"`
	Version      string `json:"mesh.cellery.io/cell-image-version"`
}

type CellSpec struct {
	ComponentTemplates []ComponentTemplate `json:"servicesTemplates"`
}

type ComponentTemplate struct {
	Metadata ComponentTemplateMetadata `json:"metadata"`
}

type ComponentTemplateMetadata struct {
	Name string `json:"name"`
}

type CellStatus struct {
	Status       string `json:"status"`
	Gateway      string `json:"gatewayHostname"`
	ServiceCount int    `json:"serviceCount"`
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
	HttpApis []GatewayHttpApi `json:"http"`
}

type GatewayHttpApi struct {
	Backend     string              `json:"backend"`
	Context     string              `json:"context"`
	Definitions []GatewayDefinition `json:"definitions"`
}

type GatewayDefinition struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type Spinner struct {
	mux            sync.Mutex
	core           *spin.Spinner
	action         string
	previousAction string
	isSpinning     bool
	error          bool
}

type Gcp struct {
	Compute GcpCompute `json:"compute"`
	Core    GcpCore    `json:"core"`
}

type GcpCompute struct {
	Region string `json:"region"`
	Zone   string `json:"zone"`
}

type GcpCore struct {
	Account string `json:"account"`
	Project string `json:"project"`
}

type RegistryCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CellImageMetaData struct {
	Organization string                        `json:"org"`
	Name         string                        `json:"name"`
	Version      string                        `json:"version"`
	Components   []string                      `json:"components"`
	Dependencies map[string]*CellImageMetaData `json:"dependencies"`
}

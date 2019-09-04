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
 *
 */

package image

type Cell struct {
	Kind         string       `json:"kind"`
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
	GateWayTemplate    Gateway             `json:"gatewayTemplate"`
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

type Gateway struct {
	GatewaySpec GatewaySpec `json:"spec"`
}

type GatewaySpec struct {
	HttpApis []GatewayHttpApi `json:"http"`
	TcpApis  []GatewayTcpApi  `json:"tcp"`
	GrpcApis []GatewayGrpcApi `json:"grpc"`
	Host     string           `json:"host"`
}

type GatewayHttpApi struct {
	Backend      string              `json:"backend"`
	Context      string              `json:"context"`
	Definitions  []GatewayDefinition `json:"definitions"`
	Global       bool                `json:"global"`
	Authenticate string              `json:"authenticate"`
}

type GatewayTcpApi struct {
	Backend     string `json:"backendHost"`
	BackendPort uint16 `json:"backendPort"`
	Port        uint16 `json:"port"`
}

type GatewayGrpcApi struct {
	Backend     string `json:"backendHost"`
	BackendPort uint16 `json:"backendPort"`
	Port        uint16 `json:"port"`
}

type GatewayDefinition struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type CellImage struct {
	Registry     string
	Organization string
	ImageName    string
	ImageVersion string
}

type CellImageName struct {
	Organization string `json:"org"`
	Name         string `json:"name"`
	Version      string `json:"ver"`
}

type MetaData struct {
	CellImageName
	SchemaVersion       string                        `json:"schemaVersion"`
	Kind                string                        `json:"kind"`
	Components          map[string]*ComponentMetaData `json:"components"`
	BuildTimestamp      int64                         `json:"buildTimestamp"`
	BuildCelleryVersion string                        `json:"buildCelleryVersion"`
	ZeroScalingRequired bool                          `json:"zeroScalingRequired"`
	AutoScalingRequired bool                          `json:"autoScalingRequired"`
}

type ComponentMetaData struct {
	DockerImage          string                 `json:"dockerImage"`
	IsDockerPushRequired bool                   `json:"isDockerPushRequired"`
	Labels               map[string]string      `json:"labels"`
	IngressTypes         []string               `json:"ingressTypes"`
	Dependencies         *ComponentDependencies `json:"dependencies"`
}

type ComponentDependencies struct {
	Cells      map[string]*MetaData `json:"cells"`
	Composites map[string]*MetaData `json:"composites"`
	Components []string             `json:"components"`
}

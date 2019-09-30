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
	Kind     string       `json:"kind"`
	MetaData CellMetaData `json:"metadata"`
	Spec     CellSpec     `json:"spec"`
	Status   CellStatus   `json:"status"`
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
	Components []Component `json:"components"`
	Gateway    Gateway     `json:"gateway,omitempty"`
}

type Component struct {
	Metadata ComponentMetadata `json:"metadata"`
	Spec     ComponentSpec     `json:"spec"`
}

type ComponentMetadata struct {
	Name string `json:"name"`
}

type ComponentSpec struct {
	Ports []ComponentPort `json:"ports"`
}

type ComponentPort struct {
	Protocol string `json:"protocol"`
}

type Gateway struct {
	Spec GatewaySpec `json:"spec"`
}

type GatewaySpec struct {
	Ingress Ingress `json:"ingress"`
}

type Ingress struct {
	Extensions Extensions    `json:"extensions"`
	HTTP       []HttpIngress `json:"http"`
	GRPC       []GrpcIngress `json:"grpc"`
	TCP        []TcpIngress  `json:"tcp"`
}

type Extensions struct {
	ClusterIngress ClusterIngress `json:"clusterIngress"`
}

type ClusterIngress struct {
	Host string `json:"host"`
}

type HttpIngress struct {
	Context     string      `json:"context"`
	Version     string      `json:"version"`
	Global      bool        `json:"global"`
	Port        int         `json:"port"`
	Destination Destination `json:"destination"`
}

type GrpcIngress struct {
	Port        int         `json:"port"`
	Destination Destination `json:"destination"`
}

type TcpIngress struct {
	Port        int         `json:"port"`
	Destination Destination `json:"destination"`
}

type Destination struct {
	Host string `json:"host"`
}

type CellStatus struct {
	Status         string `json:"status"`
	Gateway        string `json:"gatewayServiceName"`
	ComponentCount int    `json:"componentCount"`
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

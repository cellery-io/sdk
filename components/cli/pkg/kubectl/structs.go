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
	CellMetaData CellMetaData `json:"metadata"`
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

type CellStatus struct {
	Status       string `json:"status"`
	Gateway      string `json:"gatewayHostname"`
	ServiceCount int    `json:"serviceCount"`
}

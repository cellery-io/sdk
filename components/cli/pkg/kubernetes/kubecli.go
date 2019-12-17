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

package kubernetes

const kubectl = "kubectl"

// KubeCli represents kubernetes client.
type KubeCli interface {
	GetCells() ([]Cell, error)
	SetVerboseMode(enable bool)
	DeleteResource(kind, instance string) (string, error)
	GetComposites() ([]Composite, error)
	GetInstancesNames() ([]string, error)
	GetCell(cellName string) (Cell, error)
	GetComposite(compositeName string) (Composite, error)
	GetInstanceBytes(instanceKind, InstanceName string) ([]byte, error)
	DescribeCell(cellName string) error
	Version() (string, string, error)
	GetServices(cellName string) (Services, error)
	StreamCellLogsUserComponents(instanceName string, follow bool) error
	StreamCellLogsAllComponents(instanceName string, follow bool) error
	StreamComponentLogs(instanceName, componentName string, follow bool) error
	JsonPatch(kind, instance, jsonPatch string) error
	ApplyFile(file string) error
	GetCellInstanceAsMapInterface(cell string) (map[string]interface{}, error)
	GetCompositeInstanceAsMapInterface(composite string) (map[string]interface{}, error)
	GetPodsForCell(cellName string) (Pods, error)
	GetPodsForComposite(compName string) (Pods, error)
	GetVirtualService(vs string) (VirtualService, error)
	IsInstanceAvailable(instanceName string) error
	IsComponentAvailable(instanceName, componentName string) error
	GetContext() (string, error)
	GetContexts() ([]byte, error)
	UseContext(context string) error
	GetMasterNodeName() (string, error)
	ApplyLabel(itemType, itemName, labelName string, overWrite bool) error
	DeletePersistedVolume(persistedVolume string) error
	DeleteAllCells() error
	SetNamespace(namespace string) error
	DeleteNameSpace(nameSpace string) error
	CreateNamespace(namespace string) error
}

type CelleryKubeCli struct {
}

// NewCelleryCli returns a CelleryCli instance.
func NewCelleryKubeCli() *CelleryKubeCli {
	kubeCli := &CelleryKubeCli{}
	return kubeCli
}

func (kubeCli *CelleryKubeCli) SetVerboseMode(enable bool) {
	verboseMode = enable
}

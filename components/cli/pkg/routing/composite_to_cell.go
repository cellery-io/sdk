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

package routing

import (
	"fmt"
	"os"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/kubernetes"
)

type CompositeToCellRoute struct {
	Src           kubernetes.Composite
	CurrentTarget kubernetes.Cell
	NewTarget     kubernetes.Cell
}

func (router *CompositeToCellRoute) Check() error {
	// if the target instance has only TCP components exposed, will not work.
	// TODO: remove this once TCP is supported
	if len(router.NewTarget.CellSpec.GateWayTemplate.GatewaySpec.Ingress.HttpApis) == 0 &&
		len(router.NewTarget.CellSpec.GateWayTemplate.GatewaySpec.Ingress.TcpApis) > 0 {
		return fmt.Errorf("traffic switching to TCP cells not supported")
	}
	err := checkForMatchingApis(&router.CurrentTarget, &router.NewTarget)
	if err != nil {
		return err
	}
	return nil
}

func (router *CompositeToCellRoute) Build(percentage int, isSessionAware bool, routesFile string) error {
	modfiedVss, err := buildRoutesForCellTarget(&router.NewTarget, router.Src.CompositeMetaData.Name,
		router.CurrentTarget.CellMetaData.Name, percentage, isSessionAware)
	if err != nil {
		return err
	}
	// if the percentage is 100, the running cell instance now fully depends on the new instance,
	// hence update the dependency annotation
	// additionally, if the percentage is 100, include the original gateway service name as an annotation.
	var modifiedSrcCompositeInst *kubernetes.Composite
	var gw []byte
	if percentage == 100 {
		modifiedSrcCompositeInst, err = getModifiedCompositeSrcInstance(&router.Src,
			router.CurrentTarget.CellMetaData.Name, router.NewTarget.CellMetaData.Name,
			router.NewTarget.CellMetaData.Annotations.Name, router.NewTarget.CellMetaData.Annotations.Version,
			router.NewTarget.CellMetaData.Annotations.Organization, cellDependencyKind)
		if err != nil {
			return err
		}
		// get the modified gw
		gw, err = getModifiedGateway(router.NewTarget.CellMetaData.Name, router.CurrentTarget.CellMetaData.Name)
		if err != nil {
			return err
		}
	}
	// create k8s artifacts
	err = writeCompositeToCellArtifactsToFile(routesFile, modfiedVss, modifiedSrcCompositeInst, gw)
	if err != nil {
		return err
	}

	return nil
}

func writeCompositeToCellArtifactsToFile(policiesFile string, vs *kubernetes.VirtualService, compositeInstance *kubernetes.Composite, gw []byte) error {
	f, err := os.OpenFile(policiesFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	// virtual services
	vsYamlContent, err := yaml.Marshal(vs)
	if err != nil {
		return err
	}
	if _, err := f.Write(vsYamlContent); err != nil {
		return err
	}
	if _, err := f.Write([]byte("---\n")); err != nil {
		return err
	}
	// composite
	compYamlContent, err := yaml.Marshal(compositeInstance)
	if err != nil {
		return err
	}
	if _, err := f.Write(compYamlContent); err != nil {
		return err
	}
	if _, err := f.Write([]byte("---\n")); err != nil {
		return err
	}
	// gateway
	gwYaml, err := yaml.JSONToYAML(gw)
	if err != nil {
		return err
	}
	if _, err := f.Write(gwYaml); err != nil {
		return err
	}
	return nil
}

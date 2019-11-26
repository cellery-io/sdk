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

package routing

import (
	"os"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
)

type CompositeToCompositeRoute struct {
	Src           kubernetes.Composite
	CurrentTarget kubernetes.Composite
	NewTarget     kubernetes.Composite
}

func (router *CompositeToCompositeRoute) Check() error {
	return nil
}

func (router *CompositeToCompositeRoute) Build(cli cli.Cli, percentage int, isSessionAware bool, routesFile string) error {

	modfiedVs, err := buildRoutesForCompositeTarget(cli, router.Src.CompositeMetaData.Name, &router.NewTarget,
		&router.CurrentTarget, percentage)
	if err != nil {
		return err
	}
	// if the percentage is 100, the cell instance now fully depends on the new composite instance,
	// hence update the dependency annotation.
	var modifiedTargetCompInst *kubernetes.Composite
	var modifiedSrcCompositeInst *kubernetes.Composite
	if percentage == 100 {
		modifiedSrcCompositeInst, err = getModifiedCompositeSrcInstance(&router.Src,
			router.CurrentTarget.CompositeMetaData.Name, router.NewTarget.CompositeMetaData.Name,
			router.NewTarget.CompositeMetaData.Annotations.Name, router.NewTarget.CompositeMetaData.Annotations.Version,
			router.NewTarget.CompositeMetaData.Annotations.Organization, compositeDependencyKind)
		if err != nil {
			return err
		}
		// additionally, update the target composite with service names of the very first dependency.
		// this is to re-create those from the controller side in case the relevant cell is deleted.
		modifiedTargetCompInst, err = getModifiedCompositeTargetInstance(&router.CurrentTarget, &router.NewTarget)
		if err != nil {
			return err
		}
	}
	// create k8s artifacts
	err = writeCompositeToCompositeArtifactsToFile(routesFile, modfiedVs, modifiedSrcCompositeInst, modifiedTargetCompInst)
	if err != nil {
		return err
	}
	return nil
}

func writeCompositeToCompositeArtifactsToFile(policiesFile string, vs *kubernetes.VirtualService, compositeSrcInstance *kubernetes.Composite, compositeTargetInstance *kubernetes.Composite) error {
	f, err := os.OpenFile(policiesFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	// virtual services
	yamlContent, err := yaml.Marshal(vs)
	if err != nil {
		return err
	}
	if _, err := f.Write(yamlContent); err != nil {
		return err
	}
	if _, err := f.Write([]byte("---\n")); err != nil {
		return err
	}
	// cell
	cellYamlContent, err := yaml.Marshal(compositeSrcInstance)
	if err != nil {
		return err
	}
	if _, err := f.Write(cellYamlContent); err != nil {
		return err
	}
	if _, err := f.Write([]byte("---\n")); err != nil {
		return err
	}
	// composite
	compYamlContent, err := yaml.Marshal(compositeTargetInstance)
	if err != nil {
		return err
	}
	if _, err := f.Write(compYamlContent); err != nil {
		return err
	}
	if _, err := f.Write([]byte("---\n")); err != nil {
		return err
	}
	return nil
}

/*
 * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this FileName to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this FileName except
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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
)

func buildRoutesForCompositeTarget(srcInst string, targetInst kubectl.Composite,
	dependencyInstance string, percentage int) (*kubectl.VirtualService, error) {
	// get current dependency composite instance
	currentDepCompInst, err := kubectl.GetComposite(dependencyInstance)
	if err != nil {
		return nil, err
	}
	// check if components in previous dependency and this dependency matches
	if !doComponentsMatch(&currentDepCompInst.CompositeSpec.ComponentTemplates,
		&targetInst.CompositeSpec.ComponentTemplates) {
		return nil, fmt.Errorf("all components do not match in current and target composite instances")
	}
	vs, err := kubectl.GetVirtualService(getVsName(srcInst))
	if err != nil {
		return nil, err
	}
	// modify the vs to include new route information.
	modifiedVs, err := getModifiedVsForCompositeTarget(&vs, dependencyInstance, targetInst.CompositeMetaData.Name, percentage,
		&targetInst.CompositeSpec.ComponentTemplates)
	if err != nil {
		return nil, err
	}
	return modifiedVs, nil
}

func getModifiedVsForCompositeTarget(vs *kubectl.VirtualService, dependencyInst string, targetInst string,
	percentageForTarget int, componentTemplates *[]kubectl.ComponentTemplate) (*kubectl.VirtualService, error) {
	// http
	for i, httpRule := range vs.VsSpec.HTTP {
		for _, route := range httpRule.Route {
			// check whether the destination is either for previous dependency or the new dependency (target)
			if strings.HasPrefix(route.Destination.Host, dependencyInst) || strings.HasPrefix(route.Destination.Host,
				targetInst) {
				for _, compTemplate := range *componentTemplates {
					if strings.Contains(route.Destination.Host, "--"+compTemplate.Metadata.Name) {
						// for each component in target composite inst, modify the rules
						httpRule.Route = *buildPercentageBasedHttpRoutesForCompositeInstance(dependencyInst, targetInst,
							&compTemplate, percentageForTarget)
						goto outermostloop
					}
				}
			}
		}
	outermostloop:
		vs.VsSpec.HTTP[i] = httpRule
	}
	return vs, nil
}

func doComponentsMatch(currentDepComponents *[]kubectl.ComponentTemplate, newDepComponents *[]kubectl.ComponentTemplate) bool {
	var matchCount int
	for _, currentDep := range *currentDepComponents {
		for _, newDep := range *newDepComponents {
			if currentDep.Metadata.Name == newDep.Metadata.Name {
				matchCount++
				break
			}
		}
	}
	if matchCount == len(*currentDepComponents) {
		return true
	}
	return false
}

func getModifiedCompositeSrcInstance(name string, existingDependencyInstance string, targetInstance string, newCellImage string,
	newVersion string, newOrg string, srcDependencyKind string) (*kubectl.Composite, error) {
	compositeInst, err := kubectl.GetComposite(name)
	if err != nil {
		return nil, err
	}
	newDepStr, err := getModifiedDependencies(compositeInst.CompositeMetaData.Annotations.Dependencies, existingDependencyInstance,
		targetInstance, newCellImage, newVersion, newOrg, srcDependencyKind)
	if err != nil {
		return nil, err
	}
	compositeInst.CompositeMetaData.Annotations.Dependencies = newDepStr
	return &compositeInst, nil
}

func getModifiedCompositeTargetInstance(existingDependencyInstance string, targetInstance string) (*kubectl.Composite, error) {
	// set the original compositeInst service names as an annotation to the updated compositeInst instance
	dependencyComposite, err := kubectl.GetComposite(existingDependencyInstance)
	if err != nil {
		return nil, err
	}
	var originalDependencyCompositeServicesAnnotation string
	if dependencyComposite.CompositeMetaData.Annotations.OriginalDependencyComponentServices != "" {
		originalDependencyCompositeServicesAnnotation, err = AppendToDependencyCompositeServiceAnnotaion(existingDependencyInstance,
			dependencyComposite.CompositeMetaData.Annotations.OriginalDependencyComponentServices, &dependencyComposite)
		if err != nil {
			return nil, err
		}
	} else {
		originalDependencyCompositeServicesAnnotation, err = buildDependencyCompositeServiceAnnotaion(existingDependencyInstance, &dependencyComposite)
		if err != nil {
			return nil, err
		}
	}
	targetComposite, err := kubectl.GetComposite(targetInstance)
	if err != nil {
		return nil, err
	}
	// set to annotations of the new composite instance
	targetComposite.CompositeMetaData.Annotations.OriginalDependencyComponentServices = originalDependencyCompositeServicesAnnotation
	return &targetComposite, nil
}

func AppendToDependencyCompositeServiceAnnotaion(instance string, existingValue string, composite *kubectl.Composite) (string, error) {
	var serviceNames []string
	err := json.Unmarshal([]byte(existingValue), serviceNames)
	if err != nil {
		return "", err
	}
	for _, componentTemplate := range composite.CompositeSpec.ComponentTemplates {
		compositeHost := getCompositeServiceHost(instance, componentTemplate.Metadata.Name)
		for _, existingHost := range serviceNames {
			if existingHost != compositeHost {
				serviceNames = append(serviceNames, compositeHost)
				break
			}
		}
	}
	svcNames, err := json.Marshal(serviceNames)
	if err != nil {
		return "", err
	}
	return string(svcNames), nil
}

func buildDependencyCompositeServiceAnnotaion(instance string, composite *kubectl.Composite) (string, error) {
	var serviceNames []string
	for _, componentTemplate := range composite.CompositeSpec.ComponentTemplates {
		serviceNames = append(serviceNames, getCompositeServiceHost(instance, componentTemplate.Metadata.Name))
	}
	svcNames, err := json.Marshal(serviceNames)
	if err != nil {
		return "", err
	}
	return string(svcNames), nil
}

func buildPercentageBasedHttpRoutesForCompositeInstance(dependencyInst string, targetInst string,
	compTemplate *kubectl.ComponentTemplate, percentageForTarget int) *[]kubectl.HTTPRoute {
	var routes []kubectl.HTTPRoute
	if percentageForTarget == 100 {
		// full traffic switch to target, need only one route
		routes = append(routes, kubectl.HTTPRoute{
			Destination: kubectl.Destination{
				Host: getCompositeServiceHost(targetInst, compTemplate.Metadata.Name),
			},
			Weight: 100,
		})
	} else {
		// modify the existing Route's weight
		existingRoute := kubectl.HTTPRoute{
			Destination: kubectl.Destination{
				Host: getCompositeServiceHost(dependencyInst, compTemplate.Metadata.Name),
			},
			Weight: 100 - percentageForTarget,
		}
		// add the new route
		newRoute := kubectl.HTTPRoute{
			Destination: kubectl.Destination{
				Host: getCompositeServiceHost(targetInst, compTemplate.Metadata.Name),
			},
			Weight: percentageForTarget,
		}
		routes = append(routes, existingRoute)
		routes = append(routes, newRoute)
	}
	return &routes
}

func getModifiedDependencies(depJson string, existingDependency string, newDependency string, newCellImage string,
	newVersion string, newOrg string, srcDependencyKind string) (string, error) {
	dependencies, err := ExtractDependencies(depJson)
	if err != nil {
		return "", err
	}
	// copy all except previous dependency
	var newDependencies []map[string]string
	for _, dependency := range dependencies {
		if dependency[instance] != existingDependency {
			newDependencies = append(newDependencies, dependency)
		}
	}
	// create & add the new dependency
	newDepMap := make(map[string]string)
	newDepMap[instance] = newDependency
	newDepMap[imageOrg] = newOrg
	newDepMap[imageName] = newCellImage
	newDepMap[imageVersion] = newVersion
	newDepMap[dependencyKind] = srcDependencyKind
	newDependencies = append(newDependencies, newDepMap)
	// set the new dependencies to Cell
	newDepByteArr, err := json.Marshal(newDependencies)
	if err != nil {
		return "", err
	}
	return string(newDepByteArr), nil
}

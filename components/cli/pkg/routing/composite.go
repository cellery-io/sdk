/*
 * Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this FileName to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this FileName except
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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cellery-io/sdk/components/cli/kubernetes"
)

type origComponentData struct {
	ComponentName  string  `json:"componentName"`
	ContainerPorts []int32 `json:"containerPorts"`
}

func buildRoutesForCompositeTarget(src string, newTarget *kubernetes.Composite, currentTarget *kubernetes.Composite,
	percentage int) (*kubernetes.VirtualService, error) {
	// check if components in previous dependency and this dependency matches
	if !doComponentsMatch(&currentTarget.CompositeSpec.ComponentTemplates,
		&newTarget.CompositeSpec.ComponentTemplates) {
		return nil, fmt.Errorf("all components do not match in current and target composite instances")
	}
	vs, err := kubernetes.GetVirtualService(getVsName(src))
	if err != nil {
		return nil, err
	}
	// modify the vs to include new route information.
	modifiedVs, err := getModifiedVsForCompositeTarget(&vs, currentTarget.CompositeMetaData.Name,
		newTarget.CompositeMetaData.Name, percentage, &newTarget.CompositeSpec.ComponentTemplates)
	if err != nil {
		return nil, err
	}
	return modifiedVs, nil
}

func getModifiedVsForCompositeTarget(vs *kubernetes.VirtualService, dependencyInst string, targetInst string,
	percentageForTarget int, componentTemplates *[]kubernetes.ComponentTemplate) (*kubernetes.VirtualService, error) {
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

func doComponentsMatch(currentDepComponents *[]kubernetes.ComponentTemplate, newDepComponents *[]kubernetes.ComponentTemplate) bool {
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

func getModifiedCompositeSrcInstance(src *kubernetes.Composite, currentTarget string, newTarget string, newCellImage string,
	newVersion string, newOrg string, srcDependencyKind string) (*kubernetes.Composite, error) {
	newDepStr, err := getModifiedDependencies(src.CompositeMetaData.Annotations.Dependencies, currentTarget,
		newTarget, newCellImage, newVersion, newOrg, srcDependencyKind)
	if err != nil {
		return nil, err
	}
	src.CompositeMetaData.Annotations.Dependencies = newDepStr
	return src, nil
}

func getModifiedCompositeTargetInstance(currentTarget *kubernetes.Composite,
	newTarget *kubernetes.Composite) (*kubernetes.Composite, error) {
	// set the original compositeInst service names as an annotation to the updated compositeInst instance
	var originalDependencyCompositeServicesAnnotation string
	var err error
	if currentTarget.CompositeMetaData.Annotations.OriginalDependencyComponentServices != "" {
		originalDependencyCompositeServicesAnnotation, err =
			appendToDependencyCompositeServiceAnnotaion(currentTarget.CompositeMetaData.Name,
				currentTarget.CompositeMetaData.Annotations.OriginalDependencyComponentServices, currentTarget)
		if err != nil {
			return nil, err
		}
	} else {
		originalDependencyCompositeServicesAnnotation, err = buildDependencyCompositeServiceAnnotaion(currentTarget)
		if err != nil {
			return nil, err
		}
	}
	// set to annotations of the new composite instance
	newTarget.CompositeMetaData.Annotations.OriginalDependencyComponentServices = originalDependencyCompositeServicesAnnotation
	return newTarget, nil
}

func appendToDependencyCompositeServiceAnnotaion(instance string, existingValue string,
	composite *kubernetes.Composite) (string, error) {
	var origCompData []origComponentData
	err := json.Unmarshal([]byte(existingValue), &origCompData)
	if err != nil {
		return "", err
	}
	for _, componentTemplate := range composite.CompositeSpec.ComponentTemplates {
		var matchFound bool
		compositeName := getCompositeName(instance, componentTemplate.Metadata.Name)
		for _, compData := range origCompData {
			if compData.ComponentName == compositeName {
				matchFound = true
				break
			}
		}
		if !matchFound {
			origCompData = append(origCompData, origComponentData{
				ComponentName:  compositeName,
				ContainerPorts: getPortsAsIntArray(&componentTemplate.Spec.Ports),
			})
		}
	}
	svcNames, err := json.Marshal(origCompData)
	if err != nil {
		return "", err
	}
	return string(svcNames), nil
}

func getPortsAsIntArray(ports *[]kubernetes.Port) []int32 {
	var intPorts []int32
	for _, port := range *ports {
		intPorts = append(intPorts, port.Port)
	}

	return intPorts
}

func buildDependencyCompositeServiceAnnotaion(composite *kubernetes.Composite) (string, error) {
	var origCompData []origComponentData
	for _, componentTemplate := range composite.CompositeSpec.ComponentTemplates {
		origCompData = append(origCompData, origComponentData{
			ComponentName:  getCompositeName(composite.CompositeMetaData.Name, componentTemplate.Metadata.Name),
			ContainerPorts: getPortsAsIntArray(&componentTemplate.Spec.Ports),
		})
	}
	data, err := json.Marshal(origCompData)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func buildPercentageBasedHttpRoutesForCompositeInstance(dependencyInst string, targetInst string,
	compTemplate *kubernetes.ComponentTemplate, percentageForTarget int) *[]kubernetes.HTTPRoute {
	var routes []kubernetes.HTTPRoute
	if percentageForTarget == 100 {
		// full traffic switch to target, need only one route
		routes = append(routes, kubernetes.HTTPRoute{
			Destination: kubernetes.Destination{
				Host: getCompositeServiceHost(targetInst, compTemplate.Metadata.Name),
			},
			Weight: 100,
		})
	} else {
		// modify the existing Route's weight
		existingRoute := kubernetes.HTTPRoute{
			Destination: kubernetes.Destination{
				Host: getCompositeServiceHost(dependencyInst, compTemplate.Metadata.Name),
			},
			Weight: 100 - percentageForTarget,
		}
		// add the new route
		newRoute := kubernetes.HTTPRoute{
			Destination: kubernetes.Destination{
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

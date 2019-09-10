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

	errors "github.com/cellery-io/sdk/components/cli/pkg/error"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
)

const compositeDependencyKind = "Composite"
const cellDependencyKind = "Cell"
const instance = "instance"
const imageOrg = "org"
const imageName = "name"
const imageVersion = "version"
const dependencyKind = "kind"

type Route interface {
	Build(percentage int, isSessionAware bool, routesFile string) error
}

func ExtractDependencies(depJson string) ([]map[string]string, error) {
	var dependencies []map[string]string
	if depJson == "" {
		// no dependencies
		return dependencies, nil
	}
	err := json.Unmarshal([]byte(depJson), &dependencies)
	if err != nil {
		return dependencies, err
	}
	return dependencies, nil
}

func GetRoutes(sourceInstances []string, currentTarget string, newTarget string) ([]Route, error) {
	var routes []Route
	//	var instancesDependingOnCells []string
	//	var instancesDependingOnComposites []string
	if len(sourceInstances) > 0 {
		for _, srcInst := range sourceInstances {
			var dependencies []map[string]string
			inst, err := kubectl.GetCell(srcInst)
			if err != nil {
				if notFound, _ := errors.IsCellInstanceNotFoundError(srcInst, err); notFound {
					// might be a composite, check whether the srcInst is a composite
					compInst, err := kubectl.GetComposite(srcInst)
					if err != nil {
						return nil, err
					}
					dependencies, err = ExtractDependencies(compInst.CompositeMetaData.Annotations.Dependencies)
					if err != nil {
						return nil, err
					}
					var route Route
					for _, dependency := range dependencies {
						if dependency[instance] == currentTarget {
							route = buildRouteForDependency(dependency, inst.CellMetaData.Name, currentTarget, newTarget)
							if route != nil {
								routes = append(routes, route)
								break
							}
						}
					}
				} else {
					return nil, err
				}
			} else {
				dependencies, err = ExtractDependencies(inst.CellMetaData.Annotations.Dependencies)
				if err != nil {
					return nil, err
				}
				var route Route
				for _, dependency := range dependencies {
					if dependency[instance] == currentTarget {
						route = buildRouteForDependency(dependency, inst.CellMetaData.Name, currentTarget, newTarget)
						if route != nil {
							routes = append(routes, route)
							break
						}
					}
				}
			}
		}
	} else {
		// need to get all cell instances, then check if there are instances which depend on the `newTarget`
		cellInstances, err := kubectl.GetCells()
		if err != nil {
			return nil, err
		}
		for _, cellInst := range cellInstances.Items {
			dependencies, err := ExtractDependencies(cellInst.CellMetaData.Annotations.Dependencies)
			if err != nil {
				return nil, err
			}
			var route Route
			for _, dependency := range dependencies {
				if dependency[instance] == currentTarget {
					route = buildRouteForDependency(dependency, cellInst.CellMetaData.Name, currentTarget, newTarget)
					if route != nil {
						routes = append(routes, route)
						break
					}
				}
			}
		}
		// also take all composites, then check if there are instances which depend on the `newTarget`
		compositeIntsance, err := kubectl.GetComposites()
		if err != nil {
			return nil, err
		}
		var route Route
		for _, compositeInst := range compositeIntsance.Items {
			dependencies, err := ExtractDependencies(compositeInst.CompositeMetaData.Annotations.Dependencies)
			if err != nil {
				return nil, err
			}
			for _, dependency := range dependencies {
				if dependency[instance] == currentTarget {
					route = buildRouteForDependency(dependency, compositeInst.CompositeMetaData.Name, currentTarget, newTarget)
					if route != nil {
						routes = append(routes, route)
						break
					}
				}
			}
		}
	}
	return routes, nil
}

func buildRouteForDependency(dependency map[string]string, compositeSrcName string, currentTarget string, newTarget string) Route {
	var route Route = nil
	if dependency[dependencyKind] == compositeDependencyKind {
		route = &CompositeToCompositeRoute{
			Src:           compositeSrcName,
			CurrentTarget: currentTarget,
			NewTarget:     newTarget,
		}
	} else if dependency[dependencyKind] == cellDependencyKind {
		route = &CellToCellRoute{
			Src:           compositeSrcName,
			CurrentTarget: currentTarget,
			NewTarget:     newTarget,
		}
	}
	return route
}

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

	"github.com/cellery-io/sdk/components/cli/cli"
	errors "github.com/cellery-io/sdk/components/cli/pkg/error"
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
)

const compositeDependencyKind = "Composite"
const cellDependencyKind = "Cell"
const instance = "instance"
const imageOrg = "org"
const imageName = "name"
const imageVersion = "version"
const dependencyKind = "kind"

type Route interface {
	Check() error
	Build(cli cli.Cli, percentage int, isSessionAware bool, routesFile string) error
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

func GetRoutes(cli cli.Cli, sourceInstances []string, currentTarget string, newTarget string) ([]Route, error) {
	var routes []Route
	if len(sourceInstances) > 0 {
		for _, srcInst := range sourceInstances {
			var dependencies []map[string]string
			inst, err := cli.KubeCli().GetCell(srcInst)
			if err != nil {
				if notFound, _ := errors.IsCellInstanceNotFoundError(srcInst, err); notFound {
					// might be a composite, check whether the srcInst is a composite
					compInst, err := cli.KubeCli().GetComposite(srcInst)
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
							route, err = buildRouteForDependencyOfAComposite(cli, dependency, &compInst, currentTarget, newTarget)
							if err != nil {
								return nil, err
							}
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
						route, err = buildRouteForDependencyOfACell(cli, dependency, &inst, currentTarget, newTarget)
						if err != nil {
							return nil, err
						}
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
		cellInstances, err := cli.KubeCli().GetCells()
		if err != nil {
			return nil, err
		}
		for _, cellInst := range cellInstances {
			dependencies, err := ExtractDependencies(cellInst.CellMetaData.Annotations.Dependencies)
			if err != nil {
				return nil, err
			}
			var route Route
			for _, dependency := range dependencies {
				if dependency[instance] == currentTarget {
					route, err = buildRouteForDependencyOfACell(cli, dependency, &cellInst, currentTarget, newTarget)
					if err != nil {
						return nil, err
					}
					if route != nil {
						routes = append(routes, route)
						break
					}
				}
			}
		}
		// also take all composites, then check if there are instances which depend on the `newTarget`
		compositeIntsance, err := cli.KubeCli().GetComposites()
		if err != nil {
			return nil, err
		}
		var route Route
		for _, compositeInst := range compositeIntsance {
			dependencies, err := ExtractDependencies(compositeInst.CompositeMetaData.Annotations.Dependencies)
			if err != nil {
				return nil, err
			}
			for _, dependency := range dependencies {
				if dependency[instance] == currentTarget {
					route, err = buildRouteForDependencyOfAComposite(cli, dependency, &compositeInst, currentTarget, newTarget)
					if err != nil {
						return nil, err
					}
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

func buildRouteForDependencyOfACell(cli cli.Cli, dependency map[string]string, src *kubernetes.Cell, currentTarget string, newTarget string) (Route, error) {
	var route Route = nil
	if dependency[dependencyKind] == compositeDependencyKind {
		currentTargetComp, newTargetComp, err := getTargetComposites(cli, currentTarget, newTarget)
		if err != nil {
			return nil, err
		}
		route = &CellToCompositeRoute{
			Src:           *src,
			CurrentTarget: *currentTargetComp,
			NewTarget:     *newTargetComp,
		}
	} else if dependency[dependencyKind] == cellDependencyKind {
		currentTargetCell, err := cli.KubeCli().GetCell(currentTarget)
		if err != nil {
			return nil, err
		}
		newTargetCell, err := cli.KubeCli().GetCell(newTarget)
		if err != nil {
			return nil, err
		}
		route = &CellToCellRoute{
			Src:           *src,
			CurrentTarget: currentTargetCell,
			NewTarget:     newTargetCell,
		}
	}
	return route, nil
}

func buildRouteForDependencyOfAComposite(cli cli.Cli, dependency map[string]string, src *kubernetes.Composite, currentTarget string, newTarget string) (Route, error) {
	var route Route = nil
	if dependency[dependencyKind] == compositeDependencyKind {
		currentTargetComp, newTargetComp, err := getTargetComposites(cli, currentTarget, newTarget)
		if err != nil {
			return nil, err
		}
		route = &CompositeToCompositeRoute{
			Src:           *src,
			CurrentTarget: *currentTargetComp,
			NewTarget:     *newTargetComp,
		}
	} else if dependency[dependencyKind] == cellDependencyKind {
		currentTargetCell, newTargetCell, err := getTargetCells(cli, currentTarget, newTarget)
		if err != nil {
			return nil, err
		}
		route = &CompositeToCellRoute{
			Src:           *src,
			CurrentTarget: *currentTargetCell,
			NewTarget:     *newTargetCell,
		}
	}
	return route, nil
}

func getTargetComposites(cli cli.Cli, currentTarget string, newTarget string) (*kubernetes.Composite, *kubernetes.Composite, error) {
	currentTargetComp, err := cli.KubeCli().GetComposite(currentTarget)
	if err != nil {
		return nil, nil, err
	}
	newTargetComp, err := cli.KubeCli().GetComposite(newTarget)
	if err != nil {
		return nil, nil, err
	}
	return &currentTargetComp, &newTargetComp, nil
}

func getTargetCells(cli cli.Cli, currentTarget string, newTarget string) (*kubernetes.Cell, *kubernetes.Cell, error) {
	currentTargetCell, err := cli.KubeCli().GetCell(currentTarget)
	if err != nil {
		return nil, nil, err
	}
	newTargetCell, err := cli.KubeCli().GetCell(newTarget)
	if err != nil {
		return nil, nil, err
	}
	return &currentTargetCell, &newTargetCell, nil
}

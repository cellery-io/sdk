///*
// * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
// *
// * WSO2 Inc. licenses this file to you under the Apache License,
// * Version 2.0 (the "License"); you may not use this file except
// * in compliance with the License.
// * You may obtain a copy of the License at
// *
// * http:www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing,
// * software distributed under the License is distributed on an
// * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// * KIND, either express or implied.  See the License for the
// * specific language governing permissions and limitations
// * under the License.
// */

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/util"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/routing"
)

const instance = "instance"

func RunRouteTrafficCommand(sourceInstances []string, dependencyInstance string, targetInstance string, percentage int) error {
	spinner := util.StartNewSpinner(fmt.Sprintf("Starting to route %d%% of traffic to instance %s", percentage, targetInstance))
	// check if the target instance exists
	_, err := kubectl.GetCell(targetInstance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// check the source instance and see if the dependency exists in the source
	dependingSourceInstances, err := getDependingSrcInstances(sourceInstances, dependencyInstance)
	// now we have the source instance list which actually depend on the given dependency instance.
	// get the virtual services corresponding to the given source instances and modify accordingly.
	if len(*dependingSourceInstances) == 0 {
		// no depending instances
		spinner.Stop(false)
		return fmt.Errorf("dependency cell instance %s not found among the source cell instances", dependencyInstance)
	}
	var modfiedVss []kubectl.VirtualService
	spinner.SetNewAction("Modifying routing rules")
	for _, depSrcInst := range *dependingSourceInstances {
		vs, err := kubectl.GetVirtualService(routing.GetCellVsName(depSrcInst))
		if err != nil {
			spinner.Stop(false)
			return err
		}
		// modify the vs to include new route information.
		modfiedVss = append(modfiedVss, *getModifiedVs(vs, dependencyInstance, targetInstance, percentage))
	}
	vsFile := fmt.Sprintf("./%s-vs.yaml", dependencyInstance)
	err = writeVirtualSvcsToFile(vsFile, &modfiedVss)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	defer func() {
		_ = os.Remove(vsFile)
	}()
	// perform kubectl apply
	err = kubectl.ApplyFile(vsFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully routed %d%% of traffic to instance %s", percentage, targetInstance))
	return nil
}

func getDependingSrcInstances(sourceInstances []string, dependencyInstance string) (*[]string, error) {
	var dependingSourceInstances []string
	if len(sourceInstances) > 0 {
		for _, srcInst := range sourceInstances {
			cellInst, err := kubectl.GetCell(srcInst)
			if err != nil {
				return nil, err
			}
			dependencies, err := extractDependencies(cellInst.CellMetaData.Annotations.Dependencies)
			if err != nil {
				return nil, err
			}
			for _, dependency := range dependencies {
				if dependency[instance] == dependencyInstance {
					dependingSourceInstances = append(dependingSourceInstances, srcInst)
				}
			}
		}
	} else {
		// need to get all cell instances, then check if there are instances which depend on the `dependencyInstance`
		cellInstances, err := kubectl.GetCells()
		if err != nil {
			return nil, err
		}
		for _, cellInst := range cellInstances.Items {
			dependencies, err := extractDependencies(cellInst.CellMetaData.Annotations.Dependencies)
			if err != nil {
				return nil, err
			}
			for _, dependency := range dependencies {
				if dependency[instance] == dependencyInstance {
					dependingSourceInstances = append(dependingSourceInstances, cellInst.CellMetaData.Name)
				}
			}
		}
	}
	return &dependingSourceInstances, nil
}

func extractDependencies(depJson string) ([]map[string]string, error) {
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

func getModifiedVs(vs kubectl.VirtualService, dependencyInst string, targetInst string, percentageForTarget int) *kubectl.VirtualService {
	var routesCollection []kubectl.Route
	for i, httpRule := range vs.VsSpec.HTTP {
		for _, match := range httpRule.Match {
			if strings.HasPrefix(match.Authority.Regex, fmt.Sprintf("^(%s)", dependencyInst)) {
				routesCollection = *buildRoutes(dependencyInst, targetInst, percentageForTarget)
				httpRule.Route = routesCollection
			}
		}
		vs.VsSpec.HTTP[i] = httpRule
	}
	return &vs
}

func buildRoutes(dependencyInst string, targetInst string, percentageForTarget int) *[]kubectl.Route {
	var routes []kubectl.Route
	if percentageForTarget == 100 {
		// full traffic switch to target, need only one route
		routes = append(routes, kubectl.Route{
			Destination: kubectl.Destination{
				Host: routing.GetCellGatewayHost(targetInst),
			},
			Weight: 100,
		})
	} else {
		// modify the existing Route's weight
		existingRoute := kubectl.Route{
			Destination: kubectl.Destination{
				Host: routing.GetCellGatewayHost(dependencyInst),
			},
			Weight: (100 - percentageForTarget),
		}
		// add the new route
		newRoute := kubectl.Route{
			Destination: kubectl.Destination{
				Host: routing.GetCellGatewayHost(targetInst),
			},
			Weight: percentageForTarget,
		}
		routes = append(routes, existingRoute)
		routes = append(routes, newRoute)
	}
	return &routes
}

func writeVirtualSvcsToFile(policiesFile string, vss *[]kubectl.VirtualService) error {
	f, err := os.OpenFile(policiesFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	for _, vs := range *vss {
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
	}
	return nil
}

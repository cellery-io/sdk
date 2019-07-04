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
const imageOrg = "org"
const imageName = "name"
const imageVersion = "version"
const k8sMetadata = "metadata"
const k8sAnnotations = "annotations"
const cellOriginalGatewaySvcAnnKey = "mesh.cellery.io/original-gw-svc"

func RunRouteTrafficCommand(sourceInstances []string, dependencyInstance string, targetInstance string, percentage int) error {
	spinner := util.StartNewSpinner(fmt.Sprintf("Starting to route %d%% of traffic to instance %s", percentage, targetInstance))
	// check if the target instance exists
	targetInst, err := kubectl.GetCell(targetInstance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// if the target instance has only TCP components exposed, will not work.
	// TODO: remove this once TCP is supported
	if len(targetInst.CellSpec.GateWayTemplate.GatewaySpec.HttpApis) == 0 &&
		len(targetInst.CellSpec.GateWayTemplate.GatewaySpec.TcpApis) > 0 {
		spinner.Stop(false)
		return fmt.Errorf("traffic switching to TCP cells not supported")
	}
	// check the source instance and see if the dependency exists in the source
	dependingSourceInstances, err := getDependingSrcInstances(sourceInstances, dependencyInstance)
	// now we have the source instance list which actually depend on the given dependency instance.
	// get the virtual services corresponding to the given source instances and modify accordingly.
	if len(*dependingSourceInstances) == 0 {
		// no depending instances
		spinner.Stop(false)
		return fmt.Errorf("cell instance %s not found among dependencies of source cell instance(s)", dependencyInstance)
	}
	var modfiedVss []kubectl.VirtualService
	var modifiedCellInstances []kubectl.Cell
	var gw []byte
	spinner.SetNewAction("Modifying routing rules")
	for _, depSrcInst := range *dependingSourceInstances {
		vs, err := kubectl.GetVirtualService(routing.GetCellVsName(depSrcInst))
		if err != nil {
			spinner.Stop(false)
			return err
		}
		// modify the vs to include new route information.
		vss, err := getModifiedVs(vs, dependencyInstance, targetInstance, percentage)
		if err != nil {
			spinner.Stop(false)
			return err
		}
		modfiedVss = append(modfiedVss, *vss)
		// if the percentage is 100, the running cell instance now fully depends on the new instance,
		// hence update the dependency annotation
		// additionally, if the percentage is 100, include the original gateway service name as an annotation.
		if percentage == 100 {
			// get the modified instance
			modifiedCellInst, err := getModifiedCellInstances(depSrcInst, dependencyInstance, targetInstance,
				targetInst.CellMetaData.Annotations.Name, targetInst.CellMetaData.Annotations.Version,
				targetInst.CellMetaData.Annotations.Organization)
			if err != nil {
				spinner.Stop(false)
				return err
			}
			modifiedCellInstances = append(modifiedCellInstances, *modifiedCellInst)
			// get the modified gw
			gw, err = getModifiedGateway(targetInstance, dependencyInstance)
			if err != nil {
				spinner.Stop(false)
				return err
			}
		}
	}

	// create and apply kubectl artifacts
	vsFile := fmt.Sprintf("./%s-routing-artifacts.yaml", dependencyInstance)
	err = writeArtifactsToFile(vsFile, &modfiedVss, &modifiedCellInstances, gw)
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

func getModifiedGateway(targetInstance string, dependencyInstance string) ([]byte, error) {
	// check if the annotation for previous gw service name annotation exists in the gateway of the dependency instance.
	// this means that this annotation has been set previously, when doing a full traffic shift to the dependency instance.
	// if so, copy that and use it in the target instance's annotation. this is done because even if there are
	// series of traffic shifts, the original hostname used in the client cell for the dependency cell is still the same.
	depGw, err := kubectl.GetGatewayAsMapInterface(routing.GetGatewayName(dependencyInstance))
	if err != nil {
		return nil, err
	}
	annotations, err := getGwAnnotations(depGw)
	if err != nil {
		return nil, err
	}
	var originalGwAnnotation string
	if annotations[cellOriginalGatewaySvcAnnKey] != "" {
		originalGwAnnotation = annotations[cellOriginalGatewaySvcAnnKey]
	} else {
		originalGwAnnotation = routing.GetCellGatewayHost(dependencyInstance)
	}

	targetGw, err := kubectl.GetGatewayAsMapInterface(routing.GetGatewayName(targetInstance))
	if err != nil {
		return nil, err
	}
	if targetGw == nil {
		return nil, fmt.Errorf("gateway of instance %s does not exist", targetInstance)
	}
	modifiedGw, err := addOriginalGwK8sServiceName(targetGw, originalGwAnnotation)
	if err != nil {
		return nil, err
	}
	gw, err := json.Marshal(modifiedGw)
	if err != nil {
		return nil, err
	}
	return gw, nil
}

func getModifiedCellInstances(name string, existingDependency string, newDependency string, newCellImage string, newVersion string, newOrg string) (*kubectl.Cell, error) {
	cellInst, err := kubectl.GetCell(name)
	if err != nil {
		return nil, err
	}
	dependencies, err := extractDependencies(cellInst.CellMetaData.Annotations.Dependencies)
	if err != nil {
		return nil, err
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
	newDependencies = append(newDependencies, newDepMap)
	// set the new dependencies to Cell
	newDepByteArr, err := json.Marshal(newDependencies)
	if err != nil {
		return nil, err
	}
	cellInst.CellMetaData.Annotations.Dependencies = string(newDepByteArr)
	return &cellInst, nil
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

func getModifiedVs(vs kubectl.VirtualService, dependencyInst string, targetInst string, percentageForTarget int) (*kubectl.VirtualService, error) {
	// http
	for i, httpRule := range vs.VsSpec.HTTP {
		for _, route := range httpRule.Route {
			if strings.HasPrefix(route.Destination.Host, dependencyInst) {
				httpRule.Route = *buildHttpRoutes(dependencyInst, targetInst, percentageForTarget)
			}
		}
		vs.VsSpec.HTTP[i] = httpRule
	}
	// not supported atm
	// TODO: support TCP
	// TCP
	//for i, tcpRule := range vs.VsSpec.TCP {
	//	for _, route := range tcpRule.Route {
	//		if strings.HasPrefix(route.Destination.Host, dependencyInst) {
	//			tcpRule.Route = *buildTcpRoutes(dependencyInst, targetInst, route.Destination.Port, percentageForTarget)
	//		}
	//	}
	//	vs.VsSpec.TCP[i] = tcpRule
	//}
	return &vs, nil
}

func buildHttpRoutes(dependencyInst string, targetInst string, percentageForTarget int) *[]kubectl.HTTPRoute {
	var routes []kubectl.HTTPRoute
	if percentageForTarget == 100 {
		// full traffic switch to target, need only one route
		routes = append(routes, kubectl.HTTPRoute{
			Destination: kubectl.Destination{
				Host: routing.GetCellGatewayHost(targetInst),
			},
			Weight: 100,
		})
	} else {
		// modify the existing Route's weight
		existingRoute := kubectl.HTTPRoute{
			Destination: kubectl.Destination{
				Host: routing.GetCellGatewayHost(dependencyInst),
			},
			Weight: 100 - percentageForTarget,
		}
		// add the new route
		newRoute := kubectl.HTTPRoute{
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

func buildTcpRoutes(dependencyInst string, targetInst string, port kubectl.TCPPort, percentageForTarget int) *[]kubectl.TCPRoute {
	var routes []kubectl.TCPRoute
	if percentageForTarget == 100 {
		// full traffic switch to target, need only one route
		routes = append(routes, kubectl.TCPRoute{
			Destination: kubectl.TCPDestination{
				Host: routing.GetCellGatewayHost(targetInst),
				Port: port,
			},
		})
	} else {
		// modify the existing Route's weight
		existingRoute := kubectl.TCPRoute{
			Destination: kubectl.TCPDestination{
				Host: routing.GetCellGatewayHost(dependencyInst),
				Port: port,
			},
			Weight: 100 - percentageForTarget,
		}
		// add the new route
		newRoute := kubectl.TCPRoute{
			Destination: kubectl.TCPDestination{
				Host: routing.GetCellGatewayHost(targetInst),
				Port: port,
			},
			Weight: percentageForTarget,
		}
		routes = append(routes, existingRoute)
		routes = append(routes, newRoute)
	}
	return &routes
}

func getGwMetadata(gw map[string]interface{}) (map[string]interface{}, error) {
	gwBytes, err := json.Marshal(gw[k8sMetadata])
	if err != nil {
		return nil, err
	}
	var metadata map[string]interface{}
	err = json.Unmarshal(gwBytes, &metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func getGwAnnotations(gw map[string]interface{}) (map[string]string, error) {
	// get metadata
	metadata, err := getGwMetadata(gw)
	if err != nil {
		return nil, err
	}
	// get annotations
	annotationBytes, err := json.Marshal(metadata[k8sAnnotations])
	if err != nil {
		return nil, err
	}
	var annMap map[string]string
	err = json.Unmarshal(annotationBytes, &annMap)
	if err != nil {
		return nil, err
	}
	return annMap, nil
}

func addOriginalGwK8sServiceName(gw map[string]interface{}, originalGwK8sSvsName string) (map[string]interface{}, error) {
	// get metadata
	metadata, err := getGwMetadata(gw)
	if err != nil {
		return nil, err
	}
	annMap, err := getGwAnnotations(gw)
	if err != nil {
		return nil, err
	}
	// if there are existing annotations, add the original gw k8s svc name.
	// else, create and set annotations
	if len(annMap) == 0 {
		ann := map[string]string{
			cellOriginalGatewaySvcAnnKey: originalGwK8sSvsName,
		}
		metadata[k8sAnnotations] = ann
		gw[k8sMetadata] = metadata
	} else {
		anns := make(map[string]string, len(annMap)+1)
		for k, v := range annMap {
			anns[k] = v
		}
		anns[cellOriginalGatewaySvcAnnKey] = originalGwK8sSvsName
		metadata[k8sAnnotations] = anns
		gw[k8sMetadata] = metadata
	}
	return gw, nil
}

func writeArtifactsToFile(policiesFile string, vss *[]kubectl.VirtualService, cellInstances *[]kubectl.Cell, gw []byte) error {
	f, err := os.OpenFile(policiesFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	// virtual services
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
	if len(*cellInstances) > 0 {
		for _, inst := range *cellInstances {
			yamlContent, err := yaml.Marshal(inst)
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
	}
	// gateway
	if len(gw) > 0 {
		gwYaml, err := yaml.JSONToYAML(gw)
		if err != nil {
			return err
		}
		if _, err := f.Write(gwYaml); err != nil {
			return err
		}
	}
	return nil
}

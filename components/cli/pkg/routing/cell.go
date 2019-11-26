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

	"github.com/cellery-io/sdk/components/cli/cli"
	errorpkg "github.com/cellery-io/sdk/components/cli/pkg/error"
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
)

const cellOriginalGatewaySvcAnnKey = "mesh.cellery.io/original-gw-svc"
const k8sMetadata = "metadata"
const k8sAnnotations = "annotations"
const instanceIdHeaderName = "x-instance-id"

func buildRoutesForCellTarget(cli cli.Cli, newTarget *kubernetes.Cell, src string, currentTarget string, percentage int, isSessionAware bool) (*kubernetes.VirtualService, error) {
	vs, err := cli.KubeCli().GetVirtualService(getVsName(src))
	if err != nil {
		return nil, err
	}
	// modify the vs to include new route information.
	modfiedVss, err := getModifiedVsForCellTarget(vs, currentTarget, newTarget.CellMetaData.Name, percentage,
		isSessionAware)
	if err != nil {
		return nil, err
	}
	return modfiedVss, nil
}

func getModifiedGateway(newTarget string, currentTarget string) ([]byte, error) {
	// check if the annotation for previous gw service name annotation exists in the gateway of the dependency instance.
	// this means that this annotation has been set previously, when doing a full traffic shift to the dependency instance.
	// if so, copy that and use it in the target instance's annotation. this is done because even if there are
	// series of traffic shifts, the original hostname used in the client cell for the dependency cell is still the same.
	depGw, err := kubernetes.GetGatewayAsMapInterface(getGatewayName(currentTarget))
	if err != nil {
		return nil, err
	}
	annotations, err := getAnnotations(depGw)
	if err != nil {
		return nil, err
	}
	var originalGwAnnotation string
	if annotations[cellOriginalGatewaySvcAnnKey] != "" {
		originalGwAnnotation = annotations[cellOriginalGatewaySvcAnnKey]
	} else {
		originalGwAnnotation = getCellGatewayHost(currentTarget)
	}

	targetGw, err := kubernetes.GetGatewayAsMapInterface(getGatewayName(newTarget))
	if err != nil {
		return nil, err
	}
	if targetGw == nil {
		return nil, fmt.Errorf("gateway of instance %s does not exist", newTarget)
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

func addOriginalGwK8sServiceName(gw map[string]interface{}, originalGwK8sSvsName string) (map[string]interface{}, error) {
	// get metadata
	metadata, err := getK8sMetadata(gw)
	if err != nil {
		return nil, err
	}
	annMap, err := getAnnotations(gw)
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

func getK8sMetadata(ifs map[string]interface{}) (map[string]interface{}, error) {
	metadataBytes, err := json.Marshal(ifs[k8sMetadata])
	if err != nil {
		return nil, err
	}
	var metadata map[string]interface{}
	err = json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func getAnnotations(ifs map[string]interface{}) (map[string]string, error) {
	// get metadata
	metadata, err := getK8sMetadata(ifs)
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

func getModifiedCellInstance(cellInst *kubernetes.Cell, currentTarget string, newTarget string, newCellImage string,
	newVersion string, newOrg string, srcDependencyKind string) (*kubernetes.Cell, error) {
	newDepStr, err := getModifiedDependencies(cellInst.CellMetaData.Annotations.Dependencies, currentTarget,
		newTarget, newCellImage, newVersion, newOrg, srcDependencyKind)
	if err != nil {
		return nil, err
	}
	cellInst.CellMetaData.Annotations.Dependencies = newDepStr
	return cellInst, nil
}

func getModifiedVsForCellTarget(vs kubernetes.VirtualService, dependencyInst string, targetInst string,
	percentageForTarget int, enableUserBasedSessionAwareness bool) (*kubernetes.VirtualService, error) {
	// http
	for i, httpRule := range vs.VsSpec.HTTP {
		for _, route := range httpRule.Route {
			// check whether the destination is either for previous dependency or the new dependency (target)
			if strings.HasPrefix(route.Destination.Host, dependencyInst) ||
				strings.HasPrefix(route.Destination.Host, targetInst) {
				// if this is a session based rule, should be modified with normal percentage rules only if the
				// enableUserBasedSessionAwareness flag is false
				if isSessionHeaderBasedRule(&httpRule, instanceIdHeaderName) {
					if enableUserBasedSessionAwareness {
						// need to modify and add the routes to previous dependency cell and new dependency (target) cell instances.
						// if the 'x-instance-id' header is 1, set destination to previous dependency instance gateway, and if its '2',
						// set destination to new dependency instance gateway.
						if percentageForTarget == 100 {
							dependencyInst = targetInst
						}
						route, err := getHttRouteBasedOnInstanceId(&httpRule, instanceIdHeaderName, dependencyInst, targetInst)
						if err != nil {
							return nil, err
						}
						httpRule.Route = *route

					} else {
						httpRule.Route = *buildPercentageBasedHttpRoutesForCellInstance(dependencyInst, targetInst,
							percentageForTarget)
					}
					//goto outermostloop
				} else {
					httpRule.Route = *buildPercentageBasedHttpRoutesForCellInstance(dependencyInst, targetInst,
						percentageForTarget)
					//goto outermostloop
				}
			}
		}
		//outermostloop:
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

func getHttRouteBasedOnInstanceId(httpRule *kubernetes.HTTP, sessionHeader string, dependencyInstance string,
	targetInstance string) (*[]kubernetes.HTTPRoute, error) {
	for _, match := range httpRule.Match {
		if match.Headers != nil && match.Headers[sessionHeader] != nil {
			if match.Headers[sessionHeader].Exact == "1" {
				return &[]kubernetes.HTTPRoute{
					{
						Destination: kubernetes.Destination{
							Host: getCellGatewayHost(dependencyInstance),
						},
					},
				}, nil
			} else if match.Headers[sessionHeader].Exact == "2" {
				return &[]kubernetes.HTTPRoute{
					{
						Destination: kubernetes.Destination{
							Host: getCellGatewayHost(targetInstance),
						},
					},
				}, nil
			} else {
				// should not happen
				return nil, fmt.Errorf("unable to find accepted value match for %s header, "+
					"expected either 1 or 2 but found %s", instanceIdHeaderName, match.Headers[sessionHeader].Exact)
			}
		}
	}
	// should not happen
	return nil, fmt.Errorf("unable to find accepted value match for %s header", instanceIdHeaderName)
}

func isSessionHeaderBasedRule(httpRule *kubernetes.HTTP, sessionHeader string) bool {
	for _, match := range httpRule.Match {
		if match.Headers != nil && match.Headers[sessionHeader] != nil {
			// this is a rule based on session header
			return true
		}
	}
	return false
}

func buildPercentageBasedHttpRoutesForCellInstance(dependencyInst string, targetInst string,
	percentageForTarget int) *[]kubernetes.HTTPRoute {
	var routes []kubernetes.HTTPRoute
	if percentageForTarget == 100 {
		// full traffic switch to target, need only one route
		routes = append(routes, kubernetes.HTTPRoute{
			Destination: kubernetes.Destination{
				Host: getCellGatewayHost(targetInst),
			},
			Weight: 100,
		})
	} else {
		// modify the existing Route's weight
		existingRoute := kubernetes.HTTPRoute{
			Destination: kubernetes.Destination{
				Host: getCellGatewayHost(dependencyInst),
			},
			Weight: 100 - percentageForTarget,
		}
		// add the new route
		newRoute := kubernetes.HTTPRoute{
			Destination: kubernetes.Destination{
				Host: getCellGatewayHost(targetInst),
			},
			Weight: percentageForTarget,
		}
		routes = append(routes, existingRoute)
		routes = append(routes, newRoute)
	}
	return &routes
}

func buildTcpRoutes(dependencyInst string, targetInst string, port kubernetes.TCPPort, percentageForTarget int) *[]kubernetes.TCPRoute {
	var routes []kubernetes.TCPRoute
	if percentageForTarget == 100 {
		// full traffic switch to target, need only one route
		routes = append(routes, kubernetes.TCPRoute{
			Destination: kubernetes.TCPDestination{
				Host: getCellGatewayHost(targetInst),
				Port: port,
			},
		})
	} else {
		// modify the existing Route's weight
		existingRoute := kubernetes.TCPRoute{
			Destination: kubernetes.TCPDestination{
				Host: getCellGatewayHost(dependencyInst),
				Port: port,
			},
			Weight: 100 - percentageForTarget,
		}
		// add the new route
		newRoute := kubernetes.TCPRoute{
			Destination: kubernetes.TCPDestination{
				Host: getCellGatewayHost(targetInst),
				Port: port,
			},
			Weight: percentageForTarget,
		}
		routes = append(routes, existingRoute)
		routes = append(routes, newRoute)
	}
	return &routes
}

func checkForMatchingApis(currentTarget *kubernetes.Cell, newTarget *kubernetes.Cell) error {
outer:
	for _, currTargetGwApi := range currentTarget.CellSpec.GateWayTemplate.GatewaySpec.Ingress.HttpApis {
		for _, newTargetGwApi := range newTarget.CellSpec.GateWayTemplate.GatewaySpec.Ingress.HttpApis {
			if currTargetGwApi.Context == newTargetGwApi.Context {
				if doApisVersionsMatch(&newTargetGwApi, &currTargetGwApi) {
					// if matches, continue the outer loop to check for other APIs
					continue outer
				} else {
					// versions does not match
					return errorpkg.CellGwApiVersionMismatchError{
						currentTarget.CellMetaData.Name, newTarget.CellMetaData.Name,
						currTargetGwApi.Context, currTargetGwApi.Version, newTargetGwApi.Version,
					}
				}
			}
		}
		// no match, error
		return errorpkg.CellGwApiVersionMismatchError{
			currentTarget.CellMetaData.Name, newTarget.CellMetaData.Name,
			currTargetGwApi.Context, currTargetGwApi.Version, "",
		}
	}
	return nil
}

func doApisVersionsMatch(api1 *kubernetes.GatewayHttpApi, api2 *kubernetes.GatewayHttpApi) bool {
	return api1.Version == api2.Version
}

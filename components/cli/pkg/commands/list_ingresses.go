/*
 * Copyright (c) 2018 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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

package commands

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	errorpkg "github.com/cellery-io/sdk/components/cli/pkg/error"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/util"

	"github.com/olekukonko/tablewriter"

	"github.com/ghodss/yaml"
)

func RunListIngresses(name string) {
	instancePattern, _ := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), name)
	if instancePattern {
		displayInstanceApisTable(name)
	} else {
		displayCellImageApisTable(name)
	}
}

func displayInstanceApisTable(instanceName string) {
	var canBeComposite bool
	cell, err := kubectl.GetCell(instanceName)
	if err != nil {
		if cellNotFound, _ := errorpkg.IsCellInstanceNotFoundError(instanceName, err); cellNotFound {
			canBeComposite = true
		} else {
			util.ExitWithErrorMessage("Failed to check available Cells", err)
		}
	} else {
		displayCellInstanceApisTable(cell, instanceName)
	}

	if canBeComposite {
		composite, err := kubectl.GetComposite(instanceName)
		if err != nil {
			if compositeNotFound, _ := errorpkg.IsCompositeInstanceNotFoundError(instanceName, err); compositeNotFound {
				util.ExitWithErrorMessage("Failed to retrieve ingresses of "+instanceName,
					errors.New(instanceName+" instance not available in the runtime"))
			} else {
				util.ExitWithErrorMessage("Failed to check available Composites", err)
			}
		} else {
			displayCompositeInstanceApisTable(composite)
		}
	}
}

func displayCompositeInstanceApisTable(composite kubectl.Composite) {
	components := composite.CompositeSpec.ComponentTemplates
	var tableData [][]string
	for i := 0; i < len(components); i++ {
		component := components[i]
		ports := component.Spec.Ports
		for j := 0; j < len(ports); j++ {
			port := ports[j]
			tableRecord := []string{component.Metadata.Name, fmt.Sprint(port.Port), port.Protocol}
			tableData = append(tableData, tableRecord)
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"COMPONENT", "PORT", "PROTOCOL"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{})
	table.AppendBulk(tableData)
	table.Render()
}

func displayCellInstanceApisTable(cell kubectl.Cell, cellInstanceName string) {
	apiArray := cell.CellSpec.GateWayTemplate.GatewaySpec.Ingress.HttpApis
	var tableData [][]string
	for i := 0; i < len(apiArray); i++ {
		for j := 0; j < len(apiArray[i].Definitions); j++ {
			url := cellInstanceName + "--gateway-service"
			path := apiArray[i].Definitions[j].Path
			context := apiArray[i].Context
			version := apiArray[i].Version
			method := apiArray[i].Definitions[j].Method
			// Add the context of the Cell
			if !strings.HasPrefix(context, "/") {
				url += "/"
			}
			url += context
			// Add the path of the API definition
			if path != "/" {
				if !strings.HasSuffix(url, "/") {
					if !strings.HasPrefix(path, "/") {
						url += "/"
					}
				} else {
					if strings.HasPrefix(path, "/") {
						url = strings.TrimSuffix(url, "/")
					}
				}
				url += path
			}
			// Add the global api url if globally exposed
			globalUrl := ""
			if apiArray[i].Global {
				if path != "/" {
					globalUrl = constants.WSO2_APIM_HOST + "/" + cellInstanceName + "/" + context + path
				} else {
					globalUrl = constants.WSO2_APIM_HOST + "/" + cellInstanceName + "/" + context
				}
			}
			tableRecord := []string{context, version, method, url, globalUrl}
			tableData = append(tableData, tableRecord)
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"CONTEXT", "VERSION", "METHOD", "LOCAL CELL GATEWAY", "GLOBAL API URL"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{})
	table.AppendBulk(tableData)
	table.Render()
}

func displayCellImageApisTable(cellImageName string) {
	cellYamlContent := image.ReadCellImageYaml(cellImageName)
	cellImageContent := &image.Cell{}
	err := yaml.Unmarshal(cellYamlContent, cellImageContent)
	if err != nil {
		util.ExitWithErrorMessage("Error while reading cell image content", err)
	}
	var tableData [][]string
	for _, component := range cellImageContent.Spec.Components {
		componentName := component.Metadata.Name
		// Iterate HTTP and Web ingresses
		for _, ingress := range cellImageContent.Spec.Gateway.Spec.Ingress.HTTP {
			backend := ingress.Destination.Host
			if componentName == backend {
				var ingressData []string
				ingressData = append(ingressData, componentName)

				var ingressType = "web"
				if len(cellImageContent.Spec.Gateway.Spec.Ingress.Extensions.ClusterIngress.Host) == 0 {
					ingressType = "http"
				}
				ingressData = append(ingressData, ingressType)
				ingressData = append(ingressData, ingress.Context)
				ingressData = append(ingressData, ingress.Version)
				ingressData = append(ingressData, strconv.Itoa(ingress.Port))
				if ingressType == "web" || ingress.Global {
					ingressData = append(ingressData, "True")
				} else {
					ingressData = append(ingressData, "False")
				}
				tableData = append(tableData, ingressData)
			}
		}
		// Iterate TCP ingresses
		for _, ingress := range cellImageContent.Spec.Gateway.Spec.Ingress.TCP {
			backend := ingress.Destination.Host
			if componentName == backend {
				var ingressData []string
				ingressData = append(ingressData, componentName)
				ingressData = append(ingressData, "tcp")
				ingressData = append(ingressData, "N/A")
				ingressData = append(ingressData, "N/A")
				ingressData = append(ingressData, strconv.Itoa(ingress.Port))
				ingressData = append(ingressData, "False")
				tableData = append(tableData, ingressData)
			}
		}
		// Iterate GRPC ingresses
		for _, ingress := range cellImageContent.Spec.Gateway.Spec.Ingress.GRPC {
			backend := ingress.Destination.Host
			if componentName == backend {
				var ingressData []string
				ingressData = append(ingressData, componentName)
				ingressData = append(ingressData, "grpc")
				ingressData = append(ingressData, "N/A")
				ingressData = append(ingressData, "N/A")
				ingressData = append(ingressData, strconv.Itoa(ingress.Port))
				ingressData = append(ingressData, "False")
				tableData = append(tableData, ingressData)
			}
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"COMPONENT", "INGRESS TYPE", "INGRESS CONTEXT", "INGRESS_VERSION", "INGRESS PORT", "GLOBALLY EXPOSED"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetAlignment(3)
	table.SetRowSeparator("-")
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold})
	table.SetColumnColor(
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{})
	table.AppendBulk(tableData)
	table.Render()
}

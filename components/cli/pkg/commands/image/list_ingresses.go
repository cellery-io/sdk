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

package image

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	errorpkg "github.com/cellery-io/sdk/components/cli/pkg/error"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
)

func RunListIngresses(cli cli.Cli, name string) error {
	instancePattern, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), name)
	if err != nil {
		return fmt.Errorf("%s is neither instance nor an image", name)
	}
	if instancePattern {
		return displayInstanceApisTable(cli, name)
	} else {
		return displayImageApisTable(cli, name)
	}
}

func displayInstanceApisTable(cli cli.Cli, instanceName string) error {
	var canBeComposite bool
	cell, err := cli.KubeCli().GetCell(instanceName)
	if err != nil {
		if cellNotFound, _ := errorpkg.IsCellInstanceNotFoundError(instanceName, err); cellNotFound {
			canBeComposite = true
		} else {
			return fmt.Errorf("failed to check available Cells, %v", err)
		}
	} else {
		displayCellInstanceApisTable(cli, cell, instanceName)
	}

	if canBeComposite {
		composite, err := cli.KubeCli().GetComposite(instanceName)
		if err != nil {
			if compositeNotFound, _ := errorpkg.IsCompositeInstanceNotFoundError(instanceName, err); compositeNotFound {
				return fmt.Errorf(fmt.Sprintf("failed to retrieve ingresses of %s, %v ", instanceName, errors.New(instanceName+" instance not available in the runtime")))
			} else {
				return fmt.Errorf("failed to check available Composites, %v", err)
			}
		} else {
			displayCompositeInstanceApisTable(cli, composite, instanceName)
		}
	}
	return nil
}

func displayCompositeInstanceApisTable(cli cli.Cli, composite kubernetes.Composite, compositeInstance string) {
	var tableData [][]string
	for _, component := range composite.CompositeSpec.ComponentTemplates {
		for _, port := range component.Spec.Ports {
			tableRecord := []string{component.Metadata.Name, port.Protocol, fmt.Sprint(port.Port)}
			tableData = append(tableData, tableRecord)
		}
	}
	if len(tableData) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"COMPONENT", "INGRESS TYPE", "INGRESS PORT"})
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
	} else {
		fmt.Fprintln(cli.Out(), fmt.Sprintf("No ingresses found for composite instance %s", compositeInstance))
	}
}

func displayCellInstanceApisTable(cli cli.Cli, cell kubernetes.Cell, cellInstanceName string) {
	apiArray := cell.CellSpec.GateWayTemplate.GatewaySpec.Ingress.HttpApis
	var ingressType = "web"
	globalContext := ""
	globalVersion := ""
	if len(cell.CellSpec.GateWayTemplate.GatewaySpec.Ingress.Extensions.ClusterIngress.Host) == 0 {
		ingressType = "http"
	}
	if cell.CellSpec.GateWayTemplate.GatewaySpec.Ingress.Extensions.ApiPublisher.Context != "" {
		globalContext = cell.CellSpec.GateWayTemplate.GatewaySpec.Ingress.Extensions.ApiPublisher.Context
	}
	if cell.CellSpec.GateWayTemplate.GatewaySpec.Ingress.Extensions.ApiPublisher.Version != "" {
		globalVersion = cell.CellSpec.GateWayTemplate.GatewaySpec.Ingress.Extensions.ApiPublisher.Version
	}
	var tableData [][]string
	for i := 0; i < len(apiArray); i++ {
		url := cellInstanceName + "--gateway-service"
		context := apiArray[i].Context
		version := apiArray[i].Version
		// Add the context of the Cell
		if !strings.HasPrefix(context, "/") {
			url += "/"
		}
		url += context
		if len(apiArray[i].Definitions) == 0 {
			tableRecord := []string{context, ingressType, version, "", "", url, cell.CellSpec.GateWayTemplate.GatewaySpec.Ingress.Extensions.ClusterIngress.Host}
			tableData = append(tableData, tableRecord)
		}
		for j := 0; j < len(apiArray[i].Definitions); j++ {
			gatewayUrl := url
			path := apiArray[i].Definitions[j].Path
			method := apiArray[i].Definitions[j].Method
			// Add the path of the API definition
			if path != "/" {
				if !strings.HasSuffix(gatewayUrl, "/") {
					if !strings.HasPrefix(path, "/") {
						gatewayUrl += "/"
					}
				} else {
					if strings.HasPrefix(path, "/") {
						gatewayUrl = strings.TrimSuffix(gatewayUrl, "/")
					}
				}
				gatewayUrl += path
			}
			// Add the global api url if globally exposed
			globalUrl := ""
			globalUrlContext := getGlobalUrlContext(globalContext, cellInstanceName)
			globalUrlVersion := getGlobalUrlVersion(globalVersion, version)
			if apiArray[i].Global {
				if path != "/" {
					globalUrl = constants.Wso2ApimHost + strings.Replace("/"+globalUrlContext+"/"+context+path+"/"+globalUrlVersion, "//", "/", -1)
				} else {
					globalUrl = constants.Wso2ApimHost + strings.Replace("/"+globalUrlContext+"/"+context+"/"+globalUrlVersion, "//", "/", -1)
				}
			}
			tableRecord := []string{context, ingressType, version, method, path, gatewayUrl, globalUrl}
			tableData = append(tableData, tableRecord)
		}
	}
	if len(tableData) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		if ingressType == "http" {
			table.SetHeader([]string{"CONTEXT", "INGRESS TYPE", "VERSION", "METHOD", "RESOURCE", "LOCAL CELL GATEWAY", "GLOBAL API URL"})
		} else {
			table.SetHeader([]string{"CONTEXT", "INGRESS TYPE", "VERSION", "METHOD", "RESOURCE", "LOCAL CELL GATEWAY", "VHOST"})
		}
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
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold})
		table.SetColumnColor(
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{})
		table.AppendBulk(tableData)
		table.Render()
	} else {
		fmt.Fprintln(cli.Out(), fmt.Sprintf("No ingresses found for cell instance, %s", cellInstanceName))
	}
}

func displayImageApisTable(cli cli.Cli, imageName string) error {
	cellYamlContent, err := image.ReadCellImageYaml(cli.FileSystem().Repository(), imageName)
	if err != nil {
		return fmt.Errorf("error while reading cell image content, %v", err)
	}
	cellImageContent := &image.Cell{}
	err = yaml.Unmarshal(cellYamlContent, cellImageContent)
	if err != nil {
		return fmt.Errorf("error while unmarshalling cell image content, %v", err)
	}

	if cellImageContent.Kind == "Cell" {
		if err := displayCellImageApisTable(cli, imageName); err != nil {
			return fmt.Errorf("error displaying cell image apis table, %v", err)
		}
	} else if cellImageContent.Kind == "Composite" {
		if err := displayCompositeImageApisTable(cli, imageName); err != nil {
			return fmt.Errorf("error displaying composite image apis table, %v", err)
		}
	}
	return nil
}

func displayCompositeImageApisTable(cli cli.Cli, compositeImageContent string) error {
	cell, err := getIngressValues(cli, compositeImageContent)
	if err != nil {
		return fmt.Errorf("error occurred while displaying composite image ingress, %v", err)
	}
	var tableData [][]string
	for _, componentDetail := range cell.Component {
		for _, ingressInfo := range componentDetail.Ingress {
			var ingressData []string
			ingressData = append(ingressData, componentDetail.ComponentName)
			if ingressInfo.IngressTypeTCP == constants.TcpIngress {
				ingressData = append(ingressData, ingressInfo.IngressTypeTCP)
			} else {
				ingressData = append(ingressData, ingressInfo.IngressType)
			}
			if (int(ingressInfo.Port)) == 0 {
				ingressData = append(ingressData, "--")
			} else {
				ingressData = append(ingressData, strconv.Itoa(int(ingressInfo.Port)))
			}
			if ingressInfo.IngressTypeTCP == constants.TcpIngress {
				ingressData = append(ingressData, fmt.Sprintf("%s_%s, %s_tcp_%s", componentDetail.ComponentName,
					constants.HOST, componentDetail.ComponentName, constants.PORT))
			} else {
				ingressData = append(ingressData, fmt.Sprintf("%s_%s, %s_%s", componentDetail.ComponentName,
					constants.HOST, componentDetail.ComponentName, constants.PORT))
			}
			tableData = append(tableData, ingressData)
		}
	}
	if len(tableData) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"COMPONENT", "INGRESS TYPE", "INGRESS PORT", "INGRESS_KEY"})
		table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
		table.SetAlignment(3)
		table.SetRowSeparator("-")
		table.SetCenterSeparator(" ")
		table.SetColumnSeparator(" ")
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold})
		table.SetColumnColor(
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{})
		table.AppendBulk(tableData)
		table.Render()
	} else {
		fmt.Fprintln(cli.Out(), fmt.Sprintf("No ingresses found for composite image, %s", compositeImageContent))
	}
	return nil
}

func displayCellImageApisTable(cli cli.Cli, cellImageContent string) error {
	cell, err := getIngressValues(cli, cellImageContent)
	if err != nil {
		return fmt.Errorf("error occurred while displaying cell image ingress, %v", err)
	}
	var tableData [][]string
	for _, componentDetail := range cell.Component {
		for ingress, ingressInfo := range componentDetail.Ingress {
			var ingressData []string
			if ingressInfo.Expose == "global" {
				ingressInfo.Expose = "true"
			} else {
				ingressInfo.Expose = "false"
			}
			if ingressInfo.IngressType == constants.HttpApiIngress && ingressInfo.Context != "" {
				for _, resourcesValue := range ingressInfo.Definition {
					for _, resource := range resourcesValue {
						ingressData = []string{componentDetail.ComponentName, ingressInfo.IngressType,
							ingressInfo.Context, ingressInfo.ApiVersion, strconv.Itoa(int(ingressInfo.Port)),
							resource.Path, resource.Method, ingressInfo.Expose, constants.NA,
							fmt.Sprintf("%s_%s_%s", componentDetail.ComponentName, ingress, "api_url")}
						tableData = append(tableData, ingressData)
					}
				}
			} else if ingressInfo.IngressType == constants.WebIngress {
				ingressData = []string{componentDetail.ComponentName, ingressInfo.IngressType,
					ingressInfo.GatewayConfig.Context, ingressInfo.ApiVersion, strconv.Itoa(int(ingressInfo.Port)),
					constants.NA, constants.NA, ingressInfo.Expose, ingressInfo.GatewayConfig.Vhost, constants.NA}
				tableData = append(tableData, ingressData)

			} else if ingressInfo.IngressType == constants.GrpcIngress {
				ingressData = []string{componentDetail.ComponentName, ingressInfo.IngressType, constants.NA,
					constants.NA, strconv.Itoa(ingressInfo.GatewayPort), constants.NA, constants.NA, constants.NA,
					ingressInfo.GatewayConfig.Vhost,
					fmt.Sprintf("%s, %s_%s", constants.GatewayHost, componentDetail.ComponentName, "grpc_port")}
				tableData = append(tableData, ingressData)

			} else if ingressInfo.IngressTypeTCP == constants.TcpIngress {
				if (int(ingressInfo.Port)) == 0 {
					ingressData = []string{componentDetail.ComponentName, ingressInfo.IngressTypeTCP, constants.NA,
						constants.NA, "--", constants.NA, constants.NA, ingressInfo.Expose,
						ingressInfo.GatewayConfig.Vhost, fmt.Sprintf("%s, %s_tcp_%s", constants.GatewayHost,
							componentDetail.ComponentName, constants.PORT)}
				} else {
					ingressData = []string{componentDetail.ComponentName, ingressInfo.IngressTypeTCP, constants.NA,
						constants.NA, strconv.Itoa(int(ingressInfo.Port)), constants.NA, constants.NA,
						ingressInfo.Expose, ingressInfo.GatewayConfig.Vhost,
						fmt.Sprintf("%s, %s_tcp_%s", constants.GatewayHost, componentDetail.ComponentName, constants.PORT)}
				}
				tableData = append(tableData, ingressData)
			}
		}
	}
	if len(tableData) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"COMPONENT", "INGRESS TYPE", "INGRESS CONTEXT", "INGRESS_VERSION", "INGRESS PORT", "RESOURCE", "METHOD", "GLOBALLY EXPOSED", "VHOST", "INGRESS_KEY"})
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
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{})
		table.AppendBulk(tableData)
		table.Render()
	} else {
		fmt.Fprintln(cli.Out(), fmt.Sprintf("No ingresses found for cell image, %s", cellImageContent))
	}
	return nil
}

func getIngressValues(cli cli.Cli, cellImageContent string) (kubernetes.Cell, error) {
	parsedCellImage, err := image.ParseImageTag(cellImageContent)
	imageDir, err := ExtractImage(cli, parsedCellImage, false)
	if err != nil {
		return kubernetes.Cell{}, fmt.Errorf("error occurred while extracting image: %s", err)
	}

	jsonFile, err := os.Open(fmt.Sprintf("%s/%s/%s/%s%s%s", imageDir, constants.ZipArtifacts, constants.CELLERY,
		parsedCellImage.ImageName, constants.ZipMetaSuffix, constants.JsonExt))
	if err != nil {
		return kubernetes.Cell{}, fmt.Errorf("error occurred while reading image_meta file: %s", err)
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return kubernetes.Cell{}, fmt.Errorf("error occurred while reading json data from image_meta file: %s", err)
	}
	cell := kubernetes.Cell{}
	err = json.Unmarshal(byteValue, &cell)
	if err != nil {
		return kubernetes.Cell{}, fmt.Errorf("error occurred while unmarshalling json data "+
			"from image_meta file: %s", err)
	}
	return cell, nil
}

func getGlobalUrlContext(globalContext string, cellInstanceName string) string {
	if globalContext != "" {
		return globalContext
	} else {
		return cellInstanceName
	}
}

func getGlobalUrlVersion(globalVersion string, apiVersion string) string {
	if globalVersion != "" {
		return globalVersion
	} else {
		return apiVersion
	}
}

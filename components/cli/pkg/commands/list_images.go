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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

type Component struct {
	name, dockerImage   string
	ports               []int
	deployment, service string
}
type ImageData struct {
	name    string
	size    string
	created string
}

func RunImage() {
	var data [][]string
	images := getImagesArray()
	for _, image := range images {
		data = append(data, []string{image.name, image.size, image.created})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IMAGE", "SIZE", "CREATED"})
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

	table.AppendBulk(data)
	table.Render()
}

func RunImageInformation(input string) {
	componentArrays := [][]Component{
		{
			{"API-GW", "cellery.io/micro-gw:v1.1", []int{443, 80}, "API_GW_deployment", "API_GW_service"},
			{"STS", "cellery.io/sts:v1.0", []int{443}, "STS_deployment", "STS_service"},
			{"hrApp", "hrApp:v1.0", []int{9443}, "hrApp_deployment", "hrApp_service"},
			{"employeeApp", "employee:v.10", []int{8080}, "employeeApp_deployment", "employeeApp_service"},
			{"stockApp", "stockApp:v2.0", []int{8085}, "stockApp_deployment", "stockApp_service"},
		},
		{
			{"API-GW", "cellery.io/micro-gw:v1.1", []int{443, 80}, "API_GW_deployment", "API_GW_service"},
			{"STS", "cellery.io/sts:v1.0", []int{443}, "STS_deployment", "STS_service"},
			{"helloApp", "helloApp:v1.0", []int{9444}, "helloApp_deployment", "helloApp_service"},
		},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"COMPONENT", "DOCKER-IMAGE", "PORTS", "DEPLOYMENT", "SERVICE"})
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
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{})

	if strings.Contains(input, ":") {
		table.AppendBulk(componentArrayToStringArray(getComponentByNameAndVersion(input, componentArrays)))
	} else {
		table.AppendBulk(componentArrayToStringArray(getComponentByUniqueId(input, componentArrays)))
	}
	table.Render()
}

func getComponentByNameAndVersion(nameAndVersion string, componentArrays [][]Component) []Component {
	components := map[string][]Component{
		"abc/hr_app_cell:v1.0.0": componentArrays[0],
		"xyz/Hello_cell:v1.0.0":  componentArrays[1],
	}

	return components[nameAndVersion]
}

func getComponentByUniqueId(uniqueId string, componentArrays [][]Component) []Component {
	components := map[string][]Component{
		"a70ad572a50f": componentArrays[0],
		"6efa497099d9": componentArrays[1],
	}

	return components[uniqueId]
}

func componentArrayToStringArray(components []Component) [][]string {
	var convertedArray [][]string

	for i := 0; i < len(components); i++ {
		component := make([]string, 5)
		component[0] = components[i].name
		component[1] = components[i].dockerImage
		component[2] = intArrayToString(components[i].ports)
		component[3] = components[i].deployment
		component[4] = components[i].service

		convertedArray = append(convertedArray, component)
	}

	return convertedArray
}

func intArrayToString(intArray []int) string {
	return strings.Trim(strings.Replace(fmt.Sprint(intArray), " ", ", ", -1), "[]")
}

func getImagesArray() []ImageData {
	var images []ImageData
	organizations, err := util.GetSubDirectoryNames(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo"))
	if err != nil {
		log.Fatal(err)
	}
	for _, organization := range organizations {
		projects, err := util.GetSubDirectoryNames(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo", organization))
		if err != nil {
			log.Fatal(err)
		}
		for _, project := range projects {
			versions, err := util.GetSubDirectoryNames(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo",
				organization, project))
			if err != nil {
				log.Fatal(err)
			}
			for _, version := range versions {
				zipFile := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo", organization, project,
					version, project+".zip")
				zipFileExists, err := util.FileExists(zipFile)
				if err != nil {
					util.ExitWithErrorMessage("Error checking if zip file exists", err)
				}
				if zipFileExists {
					size, err := util.GetFileSize(zipFile)
					if err != nil {
						log.Fatal(err)
					}
					meta, err := image.ReadMetaData(organization, project, version)
					if err != nil {
						util.ExitWithErrorMessage("Error while listing images", err)
					}
					images = append(images, ImageData{
						fmt.Sprintf("%s/%s:%s", organization, project, version),
						units.HumanSize(float64(size)),
						fmt.Sprintf("%s ago", units.HumanDuration(time.Since(time.Unix(meta.BuildTimestamp, 0)))),
					})
				}
			}
		}
	}
	return images
}

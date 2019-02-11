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
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

type Component struct {
	name, dockerImage   string
	ports               []int
	deployment, service string
}

func RunImage() error {
	data := getImagesArray()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"CELL", "VERSION", "SIZE"})
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

	return nil
}

func RunImageInformation(input string) error {
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

	return nil
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

func getImagesArray() [][]string {
	var images [][]string
	organizations, err := util.GetSubDirectoryNames(filepath.Join(util.UserHomeDir(), ".cellery", "repo"))
	if err != nil {
		log.Fatal(err)
	}
	for _, organization := range organizations {
		projects, err := util.GetSubDirectoryNames(filepath.Join(util.UserHomeDir(), ".cellery", "repo", organization))
		if err != nil {
			log.Fatal(err)
		}
		for _, project := range projects {
			versions, err := util.GetSubDirectoryNames(filepath.Join(util.UserHomeDir(), ".cellery", "repo",
				organization, project))
			if err != nil {
				log.Fatal(err)
			}
			for _, version := range versions {
				size, err := util.GetFileSize(filepath.Join(util.UserHomeDir(), ".cellery", "repo",
					organization, project, version, project+".zip"))
				if err != nil {
					log.Fatal(err)
				}
				if size > 1024 {
					images = append(images, []string{organization + "/" + project, version,
						strconv.FormatFloat(float64(size/1024), 'f', -1, 64) + "KB"})
				} else {
					images = append(images, []string{organization + "/" + project, version,
						strconv.FormatInt(size, 10) + "B"})
				}

			}
		}
	}
	return images
}

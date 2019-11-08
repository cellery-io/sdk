/*
 * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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
	"time"

	"github.com/docker/go-units"
	"github.com/olekukonko/tablewriter"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

type imageData struct {
	name    string
	size    string
	created string
	kind    string
}

func RunImage(cli cli.Cli) error {
	var data [][]string
	images, err := getImagesArray(cli)
	if err != nil {
		return fmt.Errorf("error getting images arrays, %v", err)
	}
	for _, i := range images {
		data = append(data, []string{i.name, i.size, i.created, i.kind})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IMAGE", "SIZE", "CREATED", "KIND"})
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
	table.AppendBulk(data)
	table.Render()
	return nil
}

func getImagesArray(cli cli.Cli) ([]imageData, error) {
	var images []imageData
	repoLocation := cli.FileSystem().Repository()
	organizations, err := util.GetSubDirectoryNames(repoLocation)
	if err != nil {
		log.Fatal(err)
	}
	for _, organization := range organizations {
		projects, err := util.GetSubDirectoryNames(filepath.Join(repoLocation, organization))
		if err != nil {
			log.Fatal(err)
		}
		for _, project := range projects {
			versions, err := util.GetSubDirectoryNames(filepath.Join(repoLocation, organization, project))
			if err != nil {
				log.Fatal(err)
			}
			for _, version := range versions {
				zipFile := filepath.Join(repoLocation, organization, project, version, project+cellImageExt)
				zipFileExists, err := util.FileExists(zipFile)
				if err != nil {
					return nil, fmt.Errorf("error checking if zip file exists, %v", err)
				}
				if zipFileExists {
					size, err := util.GetFileSize(zipFile)
					if err != nil {
						log.Fatal(err)
					}
					meta, err := image.ReadMetaData(organization, project, version)
					if err != nil {
						return nil, fmt.Errorf("error while listing images, %v", err)
					}
					images = append(images, imageData{
						fmt.Sprintf("%s/%s:%s", organization, project, version),
						units.HumanSize(float64(size)),
						fmt.Sprintf("%s ago", units.HumanDuration(time.Since(time.Unix(meta.BuildTimestamp, 0)))),
						fmt.Sprintf("%s", meta.Kind),
					})
				}
			}
		}
	}
	return images, nil
}

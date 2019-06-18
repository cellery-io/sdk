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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	//"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunUpdateCellComponents(instance string, image string) error {
	spinner := util.StartNewSpinner("Updating cell components")
	// parse the new image
	parsedCellImage, err := util.ParseImageTag(image)
	// Read image name from metadata.json
	imageDir, err := ExtractImage(parsedCellImage, spinner)
	defer func() {
		_ = os.RemoveAll(imageDir)
	}()
	// read metadata.json to get the image name
	metadataFile := filepath.Join(imageDir, constants.ZIP_ARTIFACTS, "cellery",
		"metadata.json")
	metadataFileContent, err := ioutil.ReadFile(metadataFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	cellImageMetadata := &util.CellImageMetaData{}
	err = json.Unmarshal(metadataFileContent, cellImageMetadata)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	imageName := cellImageMetadata.Name
	if imageName == "" {
		spinner.Stop(false)
		return fmt.Errorf("unable to extract image name from metadata file %s", metadataFile)
	}
	// now read the image and parse it
	imageFile := filepath.Join(imageDir, constants.ZIP_ARTIFACTS, "cellery", imageName+".yaml")
	imageFileContent, err := ioutil.ReadFile(imageFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	cellImage := kubectl.Cell{}
	err = yaml.Unmarshal(imageFileContent, cellImage)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// check if the cell instance to be updated exists
	cellInst, err := kubectl.GetCell(instance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	buildUpdatedInstance(&cellImage.CellSpec.ComponentTemplates, &cellInst)
	// remove the gateway related information as we are not updating them
	cellInst.CellSpec.GateWayTemplate = kubectl.Gateway{}

	updatedCellInstContents, err := yaml.Marshal(cellInst)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	cellInstFile := filepath.Join("./", fmt.Sprintf("%s.yaml", instance))
	defer func() {
		_ = os.Remove(cellInstFile)
	}()
	err = writeToFile(updatedCellInstContents, cellInstFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// kubectl apply
	err = kubectl.ApplyFile(cellInstFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully updated the instance %s with new component images in %s", instance, image))
	return nil
}

func buildUpdatedInstance(newComponentTemplates *[]kubectl.ComponentTemplate, cellInst *kubectl.Cell) {
	// updates the image names using the new component templates, if the component template name matches
	for _, componentTemplate := range *newComponentTemplates {
		for i, cellComponent := range cellInst.CellSpec.ComponentTemplates {
			if componentTemplate.Metadata.Name == cellComponent.Metadata.Name {
				cellInst.CellSpec.ComponentTemplates[i].Spec.Container = componentTemplate.Spec.Container
			}
		}
	}
}

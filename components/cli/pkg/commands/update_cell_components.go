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
	img "github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

const k8sYamlSpec = "spec"
const k8sYamlServicetemplates = "servicesTemplates"
const k8sYamlMetadata = "metadata"
const k8sYamlName = "name"
const k8sYamlContainer = "container"
const k8sYamlImage = "image"

func RunUpdateCellComponents(instance string, image string) error {
	spinner := util.StartNewSpinner("Updating cell components")
	// parse the new image
	parsedCellImage, err := img.ParseImageTag(image)
	// Read image name from metadata.json
	imageDir, err := ExtractImage(parsedCellImage, true, spinner)
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error occurred while extracting image", err)
	}
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
	cellImageMetadata := &img.CellImageMetaData{}
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
	cellImage := &kubectl.Cell{}
	err = yaml.Unmarshal(imageFileContent, cellImage)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	cellInst, err := getUpdatedCellInstance(instance, cellImage.CellSpec.ComponentTemplates)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	updatedCellInstContents, err := yaml.Marshal(cellInst)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	cellInstFile := filepath.Join("./", fmt.Sprintf("%s.yaml", instance))
	err = writeToFile(updatedCellInstContents, cellInstFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	defer func() {
		_ = os.Remove(cellInstFile)
	}()
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

func getUpdatedCellInstance(instance string, newSvcTemplates []kubectl.ComponentTemplate) (map[string]interface{}, error) {
	cellInstance, err := kubectl.GetCellInstanceAsMapInterface(instance)
	if err != nil {
		return cellInstance, err
	}
	cellSpecByteArr, err := yaml.Marshal(cellInstance[k8sYamlSpec])
	if err != nil {
		return cellInstance, err
	}
	var cellSpec map[string]interface{}
	err = yaml.Unmarshal(cellSpecByteArr, &cellSpec)
	if err != nil {
		return cellInstance, err
	}
	svcTempalateBytes, err := yaml.Marshal(cellSpec[k8sYamlServicetemplates])
	if err != nil {
		return cellInstance, err
	}
	var svcTemplates []map[string]interface{}
	err = yaml.Unmarshal(svcTempalateBytes, &svcTemplates)
	if err != nil {
		return cellInstance, err
	}

	for _, svcTemplate := range svcTemplates {
		// metadata
		svcTemplateMetadataBytes, err := yaml.Marshal(svcTemplate[k8sYamlMetadata])
		if err != nil {
			return cellInstance, err
		}
		var svcTemplateMetadata map[string]string
		err = yaml.Unmarshal(svcTemplateMetadataBytes, &svcTemplateMetadata)
		if err != nil {
			return cellInstance, err
		}
		// for each new component template, update the relevant container image
		for i, newSvcTemplate := range newSvcTemplates {
			if newSvcTemplate.Metadata.Name == svcTemplateMetadata[k8sYamlName] {
				svcTemplateSpecBytes, err := yaml.Marshal(svcTemplate[k8sYamlSpec])
				if err != nil {
					return cellInstance, err
				}
				var svcTemplateSpec map[string]interface{}
				err = yaml.Unmarshal(svcTemplateSpecBytes, &svcTemplateSpec)
				if err != nil {
					return cellInstance, err
				}
				containerSpecBytes, err := yaml.Marshal(svcTemplateSpec[k8sYamlContainer])
				if err != nil {
					return cellInstance, err
				}
				var containerSpec map[string]interface{}
				err = yaml.Unmarshal(containerSpecBytes, &containerSpec)
				if err != nil {
					return cellInstance, err
				}
				containerSpec[k8sYamlImage] = newSvcTemplate.Spec.Container.Image
				svcTemplateSpec[k8sYamlContainer] = containerSpec
				svcTemplate[k8sYamlSpec] = svcTemplateSpec
				svcTemplates[i] = svcTemplate
			}
		}
	}
	cellSpec[k8sYamlServicetemplates] = svcTemplates
	cellInstance[k8sYamlSpec] = cellSpec
	return cellInstance, nil
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

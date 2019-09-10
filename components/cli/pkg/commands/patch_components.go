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
	"strings"

	errorpkg "github.com/cellery-io/sdk/components/cli/pkg/error"

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
const k8sYamlImageEnvVars = "env"

func RunPatchComponents(instance string, fqImage string) error {
	spinner := util.StartNewSpinner(fmt.Sprintf("Patching components in instance %s", instance))
	// parse the fully qualified image name
	parsedImage, err := img.ParseImageTag(fqImage)
	imageDir, err := ExtractImage(parsedImage, true, spinner)
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
	imageMetadata := &img.MetaData{}
	err = json.Unmarshal(metadataFileContent, imageMetadata)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	imageName := imageMetadata.Name
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
	// need to decide if the file should be parsed as a Composite or a Cell
	var canBeComposite bool
	_, err = kubectl.GetCell(instance)
	if err != nil {
		if notFound, _ := errorpkg.IsCellInstanceNotFoundError(instance, err); notFound {
			// could be a composite
			canBeComposite = true
		} else {
			spinner.Stop(false)
			return err
		}
	}

	artifacttFile := filepath.Join("./", fmt.Sprintf("%s.yaml", instance))
	defer func() {
		_ = os.Remove(artifacttFile)
	}()
	// check whether this is actually a Cell or a Composite
	if canBeComposite {
		err := patchCompositeInstace(imageFileContent, instance, artifacttFile)
		if err != nil {
			spinner.Stop(false)
			return err
		}
	} else {
		err := patchCellInstance(imageFileContent, instance, artifacttFile)
		if err != nil {
			spinner.Stop(false)
			return err
		}
	}

	// kubectl apply
	err = kubectl.ApplyFile(artifacttFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully patched the instance %s with new component images in %s", instance, fqImage))
	return nil
}

func patchCompositeInstace(imageFileContent []byte, instance string, artifactFile string) error {
	image := &kubectl.Composite{}
	err := yaml.Unmarshal(imageFileContent, image)
	if err != nil {
		return err
	}
	compositeInst, err := getPatchedCompositeInstance(instance, image.CompositeSpec.ComponentTemplates)
	if err != nil {
		return err
	}

	patchedCompositeInstContents, err := yaml.Marshal(compositeInst)
	if err != nil {
		return err
	}
	err = writeToFile(patchedCompositeInstContents, artifactFile)
	if err != nil {
		return err
	}
	return nil
}

func patchCellInstance(imageFileContent []byte, instance string, artifactFile string) error {
	image := &kubectl.Cell{}
	err := yaml.Unmarshal(imageFileContent, image)
	if err != nil {
		return err
	}
	cellInst, err := getPatchedCellInstance(instance, image.CellSpec.ComponentTemplates)
	if err != nil {
		return err
	}
	patchedCellInstContents, err := yaml.Marshal(cellInst)
	if err != nil {
		return err
	}
	err = writeToFile(patchedCellInstContents, artifactFile)
	if err != nil {
		return err
	}
	return nil
}

func RunPatchForSingleComponent(instance string, component string, containerImage string, envVars []string) error {
	spinner := util.StartNewSpinner(fmt.Sprintf("Patching component %s in instance %s", component, instance))
	var canBeComposite bool
	_, err := kubectl.GetCell(instance)
	if err != nil {
		if notFound, _ := errorpkg.IsCellInstanceNotFoundError(instance, err); notFound {
			// could be a composite
			canBeComposite = true
		} else {
			spinner.Stop(false)
			return err
		}
	}
	artifactFile := filepath.Join("./", fmt.Sprintf("%s.yaml", instance))
	defer func() {
		_ = os.Remove(artifactFile)
	}()

	if canBeComposite {
		err := patchSingleComponentinComposite(instance, component, containerImage, envVars, artifactFile)
		if err != nil {
			spinner.Stop(false)
			return err
		}
	} else {
		err := patchSingleComponentinCell(instance, component, containerImage, envVars, artifactFile)
		if err != nil {
			spinner.Stop(false)
			return err
		}
	}

	// kubectl apply
	err = kubectl.ApplyFile(artifactFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully patched the component %s in instance %s with container image %s", component, instance, containerImage))
	return nil
}

func patchSingleComponentinComposite(instance string, component string, containerImage string, envVars []string, artifactFile string) error {
	compositeInst, err := getPatchedCompositeInstanceForSingleComponent(instance, component, containerImage, envVars)
	if err != nil {
		return err
	}
	patchedCompInstContents, err := yaml.Marshal(compositeInst)
	if err != nil {
		return err
	}
	err = writeToFile(patchedCompInstContents, artifactFile)
	if err != nil {
		return err
	}
	return nil
}

func patchSingleComponentinCell(instance string, component string, containerImage string, envVars []string, artifactFile string) error {
	cellInst, err := getPatchededCellInstanceForSingleComponent(instance, component, containerImage, envVars)
	if err != nil {
		return err
	}
	patchedCellInstContents, err := yaml.Marshal(cellInst)
	if err != nil {
		return err
	}
	err = writeToFile(patchedCellInstContents, artifactFile)
	if err != nil {
		return err
	}
	return nil
}

func getPatchededCellInstanceForSingleComponent(instance string, componentName string, containerImage string, newEnvVars []string) (map[string]interface{}, error) {
	err := validateEnvVars(newEnvVars)
	if err != nil {
		return nil, err
	}
	cellInstance, err := kubectl.GetCellInstanceAsMapInterface(instance)
	if err != nil {
		return cellInstance, err
	}
	cellSpec, err := getModifiedSpecForComponent(cellInstance, componentName, containerImage, newEnvVars)
	cellInstance[k8sYamlSpec] = cellSpec
	return cellInstance, nil
}

func getPatchedCompositeInstanceForSingleComponent(instance string, componentName string, containerImage string, newEnvVars []string) (map[string]interface{}, error) {
	err := validateEnvVars(newEnvVars)
	if err != nil {
		return nil, err
	}
	compositeInst, err := kubectl.GetCompositeInstanceAsMapInterface(instance)
	if err != nil {
		return compositeInst, err
	}
	compSpec, err := getModifiedSpecForComponent(compositeInst, componentName, containerImage, newEnvVars)
	compositeInst[k8sYamlSpec] = compSpec
	return compositeInst, nil
}

func convertToPodSpecEnvVars(envVars []string) []kubectl.Env {
	var podSpecEnvVars []kubectl.Env
	for _, envVar := range envVars {
		key, value := getEnvVarKeyValue(envVar)
		podSpecEnvVars = append(podSpecEnvVars, kubectl.Env{
			Name:  key,
			Value: value,
		})
	}
	return podSpecEnvVars
}

func validateEnvVars(envVars []string) error {
	for _, envVar := range envVars {
		if len(strings.Split(envVar, "=")) < 2 {
			return fmt.Errorf("expected env variables as <key>=<value> tuples, got %s", envVar)
		}
	}
	return nil
}

func getMergedEnvVars(existingEnvVars []kubectl.Env, newEnvVars []kubectl.Env) []kubectl.Env {
	mergedEnvVarsMap := make(map[string]string)
	// add existing env vars
	for _, value := range existingEnvVars {
		mergedEnvVarsMap[value.Name] = value.Value
	}
	// add new vars
	for _, newEnvVar := range newEnvVars {
		// will override any existing env var with the same name
		mergedEnvVarsMap[newEnvVar.Name] = newEnvVar.Value
	}
	// now create an array of kind []kubectl.Env from the map
	var mergedEnvVars []kubectl.Env
	for mapKey, mapVal := range mergedEnvVarsMap {
		mergedEnvVars = append(mergedEnvVars, kubectl.Env{
			Name:  mapKey,
			Value: mapVal,
		})
	}
	return mergedEnvVars
}

func getEnvVarKeyValue(tuple string) (string, string) {
	keyValue := strings.Split(tuple, "=")
	return keyValue[0], keyValue[1]
}

func getPatchedCellInstance(instance string, newSvcTemplates []kubectl.ComponentTemplate) (map[string]interface{}, error) {
	cellInstance, err := kubectl.GetCellInstanceAsMapInterface(instance)
	if err != nil {
		return cellInstance, err
	}
	cellSpec, err := getModifedSpec(cellInstance, newSvcTemplates)
	if err != nil {
		return cellInstance, err
	}
	cellInstance[k8sYamlSpec] = cellSpec
	return cellInstance, nil
}

func getPatchedCompositeInstance(instance string, newSvcTemplates []kubectl.ComponentTemplate) (map[string]interface{}, error) {
	compositeInstance, err := kubectl.GetCompositeInstanceAsMapInterface(instance)
	if err != nil {
		return compositeInstance, err
	}
	compositeSpec, err := getModifedSpec(compositeInstance, newSvcTemplates)
	if err != nil {
		return compositeInstance, err
	}
	compositeInstance[k8sYamlSpec] = compositeSpec
	return compositeInstance, nil
}

func getModifiedSpecForComponent(instance map[string]interface{}, componentName string, containerImage string,
	newEnvVars []string) (map[string]interface{}, error) {
	cellSpecByteArr, err := yaml.Marshal(instance[k8sYamlSpec])
	if err != nil {
		return instance, err
	}
	var cellSpec map[string]interface{}
	err = yaml.Unmarshal(cellSpecByteArr, &cellSpec)
	if err != nil {
		return instance, err
	}
	svcTempalateBytes, err := yaml.Marshal(cellSpec[k8sYamlServicetemplates])
	if err != nil {
		return instance, err
	}
	var svcTemplates []map[string]interface{}
	err = yaml.Unmarshal(svcTempalateBytes, &svcTemplates)
	if err != nil {
		return instance, err
	}

	var matchFound bool
	for i, svcTemplate := range svcTemplates {
		// metadata
		svcTemplateMetadataBytes, err := yaml.Marshal(svcTemplate[k8sYamlMetadata])
		if err != nil {
			return instance, err
		}
		var svcTemplateMetadata map[string]string
		err = yaml.Unmarshal(svcTemplateMetadataBytes, &svcTemplateMetadata)
		if err != nil {
			return instance, err
		}
		// if the component name matches, update the container
		if componentName == svcTemplateMetadata[k8sYamlName] {
			matchFound = true
			svcTemplateSpecBytes, err := yaml.Marshal(svcTemplate[k8sYamlSpec])
			if err != nil {
				return instance, err
			}
			var svcTemplateSpec map[string]interface{}
			err = yaml.Unmarshal(svcTemplateSpecBytes, &svcTemplateSpec)
			if err != nil {
				return instance, err
			}
			containerSpecBytes, err := yaml.Marshal(svcTemplateSpec[k8sYamlContainer])
			if err != nil {
				return instance, err
			}
			var containerSpec map[string]interface{}
			err = yaml.Unmarshal(containerSpecBytes, &containerSpec)
			if err != nil {
				return instance, err
			}
			if len(newEnvVars) > 0 {
				envVarBytes, err := yaml.Marshal(containerSpec[k8sYamlImageEnvVars])
				if err != nil {
					return instance, err
				}
				var envVars []kubectl.Env
				err = yaml.Unmarshal(envVarBytes, &envVars)
				if err != nil {
					return instance, err
				}
				containerSpec[k8sYamlImageEnvVars] = getMergedEnvVars(envVars, convertToPodSpecEnvVars(newEnvVars))
			}
			containerSpec[k8sYamlImage] = containerImage
			svcTemplateSpec[k8sYamlContainer] = containerSpec
			svcTemplate[k8sYamlSpec] = svcTemplateSpec
			svcTemplates[i] = svcTemplate
			break
		}
	}
	if !matchFound {
		return instance, fmt.Errorf("no component with name %s found in instance %s", componentName, instance)
	}
	cellSpec[k8sYamlServicetemplates] = svcTemplates
	return cellSpec, nil
}

func getModifedSpec(instance map[string]interface{}, newSvcTemplates []kubectl.ComponentTemplate) (map[string]interface{}, error) {
	cellSpecByteArr, err := yaml.Marshal(instance[k8sYamlSpec])
	if err != nil {
		return instance, err
	}
	var cellSpec map[string]interface{}
	err = yaml.Unmarshal(cellSpecByteArr, &cellSpec)
	if err != nil {
		return instance, err
	}
	svcTempalateBytes, err := yaml.Marshal(cellSpec[k8sYamlServicetemplates])
	if err != nil {
		return instance, err
	}
	var svcTemplates []map[string]interface{}
	err = yaml.Unmarshal(svcTempalateBytes, &svcTemplates)
	if err != nil {
		return instance, err
	}

	for _, svcTemplate := range svcTemplates {
		// metadata
		svcTemplateMetadataBytes, err := yaml.Marshal(svcTemplate[k8sYamlMetadata])
		if err != nil {
			return instance, err
		}
		var svcTemplateMetadata map[string]string
		err = yaml.Unmarshal(svcTemplateMetadataBytes, &svcTemplateMetadata)
		if err != nil {
			return instance, err
		}
		// for each new component template, update the relevant container image
		for i, newSvcTemplate := range newSvcTemplates {
			if newSvcTemplate.Metadata.Name == svcTemplateMetadata[k8sYamlName] {
				svcTemplateSpecBytes, err := yaml.Marshal(svcTemplate[k8sYamlSpec])
				if err != nil {
					return instance, err
				}
				var svcTemplateSpec map[string]interface{}
				err = yaml.Unmarshal(svcTemplateSpecBytes, &svcTemplateSpec)
				if err != nil {
					return instance, err
				}
				containerSpecBytes, err := yaml.Marshal(svcTemplateSpec[k8sYamlContainer])
				if err != nil {
					return instance, err
				}
				var containerSpec map[string]interface{}
				err = yaml.Unmarshal(containerSpecBytes, &containerSpec)
				if err != nil {
					return instance, err
				}
				containerSpec[k8sYamlImage] = newSvcTemplate.Spec.Container.Image
				svcTemplateSpec[k8sYamlContainer] = containerSpec
				svcTemplate[k8sYamlSpec] = svcTemplateSpec
				svcTemplates[i] = svcTemplate
			}
		}
	}
	cellSpec[k8sYamlServicetemplates] = svcTemplates
	return cellSpec, nil
}

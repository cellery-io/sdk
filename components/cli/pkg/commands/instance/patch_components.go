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

package instance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/cellery-io/sdk/components/cli/cli"
	errorpkg "github.com/cellery-io/sdk/components/cli/pkg/error"
	"github.com/cellery-io/sdk/components/cli/pkg/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

const k8sYamlSpec = "spec"
const k8sYamlComponents = "components"
const k8sYamlMetadata = "metadata"
const k8sYamlName = "name"
const k8sYamlContainers = "containers"
const k8sYamlImage = "image"
const k8sYamlImageEnvVars = "env"
const k8sYamlContainerName = "name"
const k8sContainerTemplate = "template"

func RunPatchForSingleComponent(cli cli.Cli, instance string, component string, containerImage string, containerName string, envVars []string) error {
	//spinner := util.StartNewSpinner(fmt.Sprintf("Patching component %s in instance %s", component, instance))
	var canBeComposite bool
	_, err := cli.KubeCli().GetCell(instance)
	if err != nil {
		if notFound, _ := errorpkg.IsCellInstanceNotFoundError(instance, err); notFound {
			// could be a composite
			canBeComposite = true
		} else {
			return err
		}
	}
	artifactFile := filepath.Join("./", fmt.Sprintf("%s.yaml", instance))
	defer func() error {
		return os.Remove(artifactFile)
	}()

	if canBeComposite {
		_, err = cli.KubeCli().GetComposite(instance)
		if err != nil {
			if notFound, _ := errorpkg.IsCompositeInstanceNotFoundError(instance, err); notFound {
				// the given instance is neither a cell or a composite
				return fmt.Errorf("unable to find a running instance with name: %s", instance)
			} else {
				return err
			}
		}
		if err = patchSingleComponentinComposite(cli, instance, component, containerImage, containerName, envVars, artifactFile); err != nil {
			return fmt.Errorf("error patching single component in composite")
		}
	} else {
		if err = patchSingleComponentinCell(cli, instance, component, containerImage, containerName, envVars, artifactFile); err != nil {
			return fmt.Errorf("error patching single component in cell, %v", err)
		}
	}
	if err = cli.KubeCli().ApplyFile(artifactFile); err != nil {
		return fmt.Errorf("error applying yaml %s", artifactFile)
	}
	util.PrintSuccessMessage(fmt.Sprintf("Successfully patched the component %s in instance %s with container image %s", component, instance, containerImage))
	return nil
}

func patchSingleComponentinComposite(cli cli.Cli, instance string, component string, containerImage string, containerName string, envVars []string, artifactFile string) error {
	compositeInst, err := getPatchedCompositeInstanceForSingleComponent(cli, instance, component, containerImage, containerName, envVars)
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

func patchSingleComponentinCell(cli cli.Cli, instance string, component string, containerImage string, containerName string, envVars []string, artifactFile string) error {
	cellInst, err := getPatchededCellInstanceForSingleComponent(cli, instance, component, containerImage, containerName, envVars)
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

func getPatchededCellInstanceForSingleComponent(cli cli.Cli, instance string, componentName string, containerImage string, containerName string, newEnvVars []string) (map[string]interface{}, error) {
	err := validateEnvVars(newEnvVars)
	if err != nil {
		return nil, err
	}
	cellInstance, err := cli.KubeCli().GetCellInstanceAsMapInterface(instance)
	fmt.Println(cellInstance)
	if err != nil {
		return nil, err
	}
	cellSpec, err := getModifiedSpecForComponent(cellInstance, componentName, containerImage, containerName, newEnvVars)
	if err != nil {
		return nil, err
	}
	cellInstance[k8sYamlSpec] = cellSpec
	return cellInstance, nil
}

func getPatchedCompositeInstanceForSingleComponent(cli cli.Cli, instance string, componentName string, containerImage string, containerName string, newEnvVars []string) (map[string]interface{}, error) {
	err := validateEnvVars(newEnvVars)
	if err != nil {
		return nil, err
	}
	compositeInst, err := cli.KubeCli().GetCompositeInstanceAsMapInterface(instance)
	if err != nil {
		return nil, err
	}
	compSpec, err := getModifiedSpecForComponent(compositeInst, componentName, containerImage, containerName, newEnvVars)
	if err != nil {
		return nil, err
	}
	compositeInst[k8sYamlSpec] = compSpec
	return compositeInst, nil
}

func convertToPodSpecEnvVars(envVars []string) []kubernetes.Env {
	var podSpecEnvVars []kubernetes.Env
	for _, envVar := range envVars {
		key, value := getEnvVarKeyValue(envVar)
		podSpecEnvVars = append(podSpecEnvVars, kubernetes.Env{
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

func getMergedEnvVars(currentContainerSpec interface{}, newEnvVars []kubernetes.Env) ([]kubernetes.Env, error) {
	mergedEnvVarsMap := make(map[string]string)
	envVarBytes, err := yaml.Marshal(currentContainerSpec)
	if err != nil {
		return nil, err
	}
	var envVars []kubernetes.Env
	err = yaml.Unmarshal(envVarBytes, &envVars)
	if err != nil {
		return nil, err
	}
	// add existing env vars
	for _, value := range envVars {
		mergedEnvVarsMap[value.Name] = value.Value
	}
	// add new vars
	for _, newEnvVar := range newEnvVars {
		// will override any existing env var with the same name
		mergedEnvVarsMap[newEnvVar.Name] = newEnvVar.Value
	}
	// now create an array of kind []kubectl.Env from the map
	var mergedEnvVars []kubernetes.Env
	for mapKey, mapVal := range mergedEnvVarsMap {
		mergedEnvVars = append(mergedEnvVars, kubernetes.Env{
			Name:  mapKey,
			Value: mapVal,
		})
	}
	return mergedEnvVars, nil
}

func getEnvVarKeyValue(tuple string) (string, string) {
	keyValue := strings.Split(tuple, "=")
	return keyValue[0], keyValue[1]
}

func getModifiedSpecForComponent(instance map[string]interface{}, componentName string, containerImage string,
	containerName string, newEnvVars []string) (map[string]interface{}, error) {
	cellSpecByteArr, err := yaml.Marshal(instance[k8sYamlSpec])
	if err != nil {
		return nil, err
	}
	var cellSpec map[string]interface{}
	err = yaml.Unmarshal(cellSpecByteArr, &cellSpec)
	if err != nil {
		return nil, err
	}
	components, err := getComponentSpecs(cellSpec)
	if err != nil {
		return nil, err
	}
	var componentFound bool
	for _, component := range components {
		// metadata
		compMetadata, err := getComponentMetadate(component)
		if err != nil {
			return nil, err
		}
		// if the component name matches, update the container
		if componentName == compMetadata[k8sYamlName] {
			componentFound = true
			componentSpec, err := getComponentSpec(component)
			if err != nil {
				return nil, err
			}
			template, err := getTemplate(componentSpec)
			if err != nil {
				return nil, err
			}
			containerSpecs, err := getContainerSpecs(template)
			if err != nil {
				return nil, err
			}
			if containerName == "" {
				// container name not provided, use the first container and update it.
				if len(newEnvVars) > 0 {
					mergedEnvVars, err := getMergedEnvVars(containerSpecs[0][k8sYamlImageEnvVars], convertToPodSpecEnvVars(newEnvVars))
					if err != nil {
						return nil, err
					}
					containerSpecs[0][k8sYamlImageEnvVars] = mergedEnvVars
				}
				containerSpecs[0][k8sYamlImage] = containerImage
			} else {
				containerFound := false
				for i, spec := range containerSpecs {
					containerSpec, err := getContainerSpec(spec)
					if err != nil {
						return nil, err
					}
					if containerName == containerSpec[k8sYamlContainerName] {
						containerFound = true
						if len(newEnvVars) > 0 {
							mergedEnvVars, err := getMergedEnvVars(containerSpec[k8sYamlImageEnvVars], convertToPodSpecEnvVars(newEnvVars))
							if err != nil {
								return nil, err
							}
							containerSpec[k8sYamlImageEnvVars] = mergedEnvVars
						}
						containerSpec[k8sYamlImage] = containerImage
						containerSpecs[i] = containerSpec
						break
					}
				}
				if !containerFound {
					return nil, fmt.Errorf("no container with name %s found in component %s", containerName, componentName)
				}
			}

			template[k8sYamlContainers] = containerSpecs
			componentSpec[k8sContainerTemplate] = template
			component[k8sYamlSpec] = componentSpec
			break
		}
	}
	if !componentFound {
		return instance, fmt.Errorf("no component with name %s found in instance %s", componentName, instance)
	}
	cellSpec[k8sYamlComponents] = components
	return cellSpec, nil
}

func getComponentSpecs(cellSpec map[string]interface{}) ([]map[string]interface{}, error) {

	componentBytes, err := yaml.Marshal(cellSpec[k8sYamlComponents])
	if err != nil {
		return nil, err
	}
	var components []map[string]interface{}
	err = yaml.Unmarshal(componentBytes, &components)
	if err != nil {
		return nil, err
	}
	return components, nil
}

func getComponentSpec(component map[string]interface{}) (map[string]interface{}, error) {
	compSpecBytes, err := yaml.Marshal(component[k8sYamlSpec])
	if err != nil {
		return nil, err
	}
	var componentSpec map[string]interface{}
	err = yaml.Unmarshal(compSpecBytes, &componentSpec)
	if err != nil {
		return nil, err
	}
	return componentSpec, nil
}

func getContainerSpecs(template map[string]interface{}) ([]map[string]interface{}, error) {
	containerSpecBytes, err := yaml.Marshal(template[k8sYamlContainers])
	if err != nil {
		return nil, err
	}
	var containerSpecs []map[string]interface{}
	err = yaml.Unmarshal(containerSpecBytes, &containerSpecs)
	if err != nil {
		return nil, err
	}
	return containerSpecs, nil
}

func getContainerSpec(o interface{}) (map[string]interface{}, error) {
	var containerSpec map[string]interface{}
	containerYamlBytes, err := yaml.Marshal(o)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(containerYamlBytes, &containerSpec)
	if err != nil {
		return nil, err
	}
	return containerSpec, err
}

func getComponentMetadate(component map[string]interface{}) (map[string]interface{}, error) {
	compMetadataBytes, err := yaml.Marshal(component[k8sYamlMetadata])
	if err != nil {
		return nil, err
	}
	var compMetadata map[string]interface{}
	err = yaml.Unmarshal(compMetadataBytes, &compMetadata)
	if err != nil {
		return nil, err
	}
	return compMetadata, nil
}

func getTemplate(componentSpec map[string]interface{}) (map[string]interface{}, error) {
	templateBytes, err := yaml.Marshal(componentSpec[k8sContainerTemplate])
	if err != nil {
		return nil, err
	}
	var template map[string]interface{}
	err = yaml.Unmarshal(templateBytes, &template)
	if err != nil {
		return nil, err
	}
	return template, nil
}

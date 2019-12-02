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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"cellery.io/cellery/components/cli/pkg/constants"

	"github.com/ghodss/yaml"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/image"
	"cellery.io/cellery/components/cli/pkg/util"
	"cellery.io/cellery/components/cli/pkg/version"
)

// RunBuild executes the cell's build life cycle method and saves the generated cell image to the local repo.
// This also copies the relevant ballerina files to the ballerina repo directory.
func RunBuild(cli cli.Cli, tag string, balSource string) error {
	var err error
	var tmpProjectDir string
	var tmpCellSource string
	var parsedCellImage *image.CellImage
	var iName []byte
	currentTime := time.Now()
	tmpImageDirName := "cellery-cell-image" + currentTime.Format("27065102350415")

	if parsedCellImage, err = image.ParseImageTag(tag); err != nil {
		return fmt.Errorf("error occurred while parsing image, %v", err)
	}
	var imageName = &image.CellImageName{
		Organization: parsedCellImage.Organization,
		Name:         parsedCellImage.ImageName,
		Version:      parsedCellImage.ImageVersion,
	}
	if iName, err = json.Marshal(imageName); err != nil {
		return fmt.Errorf("error in generating cellery:ImageName construct, %v", err)
	}

	cellProjectInfo, err := os.Stat(balSource)
	if err != nil {
		return fmt.Errorf("error occured while getting fileInfo of cell project, %v", err)
	}
	// If the cell project is a Ballerina project, create a main.bal file in a temp project location
	if cellProjectInfo.IsDir() {
		// Validate that the project has only one module
		modules, _ := ioutil.ReadDir(filepath.Join(balSource, "src"))
		if len(modules) > 1 {
			return fmt.Errorf("cell project cannot contain more than one module. Found %s modules", string(len(modules)))
		}

		// Create a temporary project location to execute bal files and to generate artifacts
		tmpProjectDir = filepath.Join(cli.FileSystem().TempDir(), tmpImageDirName, balSource)
		util.CreateDir(tmpProjectDir)
		tmpCellSource = filepath.Join(tmpProjectDir, "src", modules[0].Name())
		balModuleDirPath := filepath.Join(tmpProjectDir, "src", modules[0].Name())

		// Copy project source files to temp project location
		util.CopyDir(balSource, tmpProjectDir)

		// Create a main.bal with a main function within the ballerina module in temp project directory
		if err = cli.ExecuteTask("Creating temporary executable main bal file", "Failed to create temporary main bal file",
			"", func() error {
				err = util.CreateTempMainBalFile(balModuleDirPath)
				return err
			}); err != nil {
			return err
		}
	} else {
		// Validate that the file exists
		var fileExist bool
		if fileExist, err = util.FileExists(balSource); err != nil {
			return fmt.Errorf("failed to check if file '%s' exists", util.Bold(balSource))
		}
		if !fileExist {
			return fmt.Errorf("file '%s' does not exist", util.Bold(balSource))
		}

		// Create a temporary project location to execute bal files and to generate artifacts
		tmpProjectDir = filepath.Join(cli.FileSystem().TempDir(), tmpImageDirName)
		util.CreateDir(tmpProjectDir)

		// Create a temp cell file appending the main function in temp project directory
		if err = cli.ExecuteTask("Creating temporary executable bal file", "Failed to create temporary bal file",
			"", func() error {
				tmpCellSource, err = createTempBalFile(balSource, tmpProjectDir)
				return err
			}); err != nil {
			return err
		}

	}

	// Execute ballerina build in temporary executable bal file.
	if err = cli.ExecuteTask("Executing ballerina build", "Failed to execute ballerina build",
		"", func() error {
			if err := cli.BalExecutor().Build(filepath.Base(tmpCellSource), []string{string(iName)}, tmpProjectDir); err != nil {
				return err
			}
			return err
		}); err != nil {
		return err
	}
	// Generate metadata.
	if err = cli.ExecuteTask("Generating metadata", "Failed to generate metadata",
		"", func() error {
			err := generateMetaData(cli, parsedCellImage, tmpProjectDir)
			return err
		}); err != nil {
		return err
	}

	// Create the image zip
	artifactsZip := parsedCellImage.ImageName + cellImageExt
	var zipSrc string
	if err = cli.ExecuteTask("Creating the image zip file", "Failed to create the image zip",
		"", func() error {
			zipSrc, err = createArtifactsZip(artifactsZip, tmpProjectDir, balSource)
			return err
		}); err != nil {
		return err
	}

	repoLocation := filepath.Join(cli.FileSystem().Repository(), parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion)
	var hasOldImage bool
	if hasOldImage, err = util.FileExists(repoLocation); err != nil {
		return fmt.Errorf("error occurred while removing the old image, %v", err)
	}
	if hasOldImage {
		if err = os.RemoveAll(repoLocation); err != nil {
			return fmt.Errorf("error occurred while cleaning up, %v", err)
		}
	}
	if err = util.CreateDir(repoLocation); err != nil {
		return fmt.Errorf("error occurred while creating image location, %v", err)
	}
	zipDst := filepath.Join(repoLocation, artifactsZip)

	if err = util.CopyFile(zipSrc, zipDst); err != nil {
		return fmt.Errorf("error occurred while saving image to local repo, %v", err)
	}
	if err = os.Remove(zipSrc); err != nil {
		return fmt.Errorf("error occurred while removing zipSrc dir, %v", err)
	}
	util.PrintSuccessMessage(fmt.Sprintf("Successfully built image: %s", util.Bold(tag)))
	util.PrintWhatsNextMessage("run the image", "cellery run "+tag)
	return nil
}

// generateMetaData generates the metadata file for cellery
func generateMetaData(cli cli.Cli, cellImage *image.CellImage, projectDir string) error {
	targetDir := filepath.Join(projectDir, "target")
	var err error
	var metadataJSON []byte
	var cellYamlContent []byte
	metadataFile := filepath.Join(targetDir, constants.CELLERY, "metadata.json")
	if metadataJSON, err = ioutil.ReadFile(metadataFile); err != nil {
		return fmt.Errorf("error occurred while reading metadata %s, %v", metadataFile, err)
	}
	if cellYamlContent, err = ioutil.ReadFile(filepath.Join(targetDir, constants.CELLERY, cellImage.ImageName+".yaml")); err != nil {
		return fmt.Errorf("error reading cell yaml content, %v", err)
	}
	k8sCell := &image.Cell{}
	if err = yaml.Unmarshal(cellYamlContent, k8sCell); err != nil {
		return fmt.Errorf("error unmarshalling cell yaml content, %v", err)
	}
	metadata := &image.MetaData{
		SchemaVersion: "0.1.0",
		CellImageName: image.CellImageName{
			Organization: cellImage.Organization,
			Name:         cellImage.ImageName,
			Version:      cellImage.ImageVersion,
		},
		Kind:                k8sCell.Kind,
		Components:          map[string]*image.ComponentMetaData{},
		BuildTimestamp:      time.Now().Unix(),
		BuildCelleryVersion: version.BuildVersion(),
		ZeroScalingRequired: false,
		AutoScalingRequired: false,
	}
	if err = json.Unmarshal(metadataJSON, metadata); err != nil {
		return fmt.Errorf("error unmarshalling metadata json, %v", err)
	}
	for componentName, componentMetadata := range metadata.Components {
		for alias, dependencyMetadata := range componentMetadata.Dependencies.Cells {
			if dependencyMetadata, err = extractDependenciesFromMetaData(cli, dependencyMetadata, cellImage); err != nil {
				return fmt.Errorf("error extracting cell dependencies from meta of image %s", cellImage)
			}
			metadata.Components[componentName].Dependencies.Cells[alias] = dependencyMetadata
		}

		for alias, dependencyMetadata := range componentMetadata.Dependencies.Composites {
			if dependencyMetadata, err = extractDependenciesFromMetaData(cli, dependencyMetadata, cellImage); err != nil {
				return fmt.Errorf("error extracting composite dependencies from meta of image %s", cellImage)
			}
			metadata.Components[componentName].Dependencies.Composites[alias] = dependencyMetadata

		}
		componentMetadata.IngressTypes = []string{}
	}

	// Getting the Ingress Types
	appendIfNotPresent := func(ingressTypesArray []string, newIngress string) []string {
		hasIngressType := false
		for _, ingressType := range ingressTypesArray {
			if ingressType == newIngress {
				hasIngressType = true
				break
			}
		}
		if !hasIngressType {
			return append(ingressTypesArray, newIngress)
		}
		return ingressTypesArray
	}
	if k8sCell.Kind == "Cell" {
		for _, tcpIngress := range k8sCell.Spec.Gateway.Spec.Ingress.TCP {
			metadata.Components[tcpIngress.Destination.Host].IngressTypes =
				appendIfNotPresent(metadata.Components[tcpIngress.Destination.Host].IngressTypes, "TCP")
		}
		for _, httpIngress := range k8sCell.Spec.Gateway.Spec.Ingress.HTTP {
			var ingressType string
			if k8sCell.Spec.Gateway.Spec.Ingress.Extensions.ClusterIngress.Host == "" {
				ingressType = "HTTP"
			} else {
				ingressType = "WEB"
			}
			metadata.Components[httpIngress.Destination.Host].IngressTypes =
				appendIfNotPresent(metadata.Components[httpIngress.Destination.Host].IngressTypes, ingressType)
		}
		for _, grpcIngress := range k8sCell.Spec.Gateway.Spec.Ingress.GRPC {
			metadata.Components[grpcIngress.Destination.Host].IngressTypes =
				appendIfNotPresent(metadata.Components[grpcIngress.Destination.Host].IngressTypes, "GRPC")
		}
	} else {
		for _, component := range k8sCell.Spec.Components {
			for _, port := range component.Spec.Ports {
				if strings.ToUpper(port.Protocol) == "HTTP" {
					metadata.Components[component.Metadata.Name].IngressTypes =
						appendIfNotPresent(metadata.Components[component.Metadata.Name].IngressTypes, "HTTP")
				} else if strings.ToUpper(port.Protocol) == "TCP" {
					metadata.Components[component.Metadata.Name].IngressTypes =
						appendIfNotPresent(metadata.Components[component.Metadata.Name].IngressTypes, "TCP")
				} else if strings.ToUpper(port.Protocol) == "GRPC" {
					metadata.Components[component.Metadata.Name].IngressTypes =
						appendIfNotPresent(metadata.Components[component.Metadata.Name].IngressTypes, "GRPC")
				}
			}
		}
	}
	var metadataFileContent []byte
	if metadataFileContent, err = json.Marshal(metadata); err != nil {
		return fmt.Errorf("error unmarshalling metadata file content, %v", err)
	}
	if err = ioutil.WriteFile(metadataFile, metadataFileContent, 0666); err != nil {
		return fmt.Errorf("error writing content to metadata file, %v", err)
	}
	return nil
}

func extractDependenciesFromMetaData(cli cli.Cli, dependencyMetadata *image.MetaData, cellImage *image.CellImage) (*image.MetaData, error) {
	var err error
	cellImageZip := path.Join(cli.FileSystem().Repository(), dependencyMetadata.Organization, dependencyMetadata.Name,
		dependencyMetadata.Version, dependencyMetadata.Name+cellImageExt)
	dependencyImage := dependencyMetadata.Organization + "/" + dependencyMetadata.Name +
		":" + dependencyMetadata.Version
	if cellImage.Registry != "" {
		dependencyImage = cellImage.Registry + "/" + dependencyImage
	}
	// Pulling the dependency if not exist (This will not be executed most of the time)
	var dependencyExists bool
	if dependencyExists, err = util.FileExists(cellImageZip); err != nil {
		return nil, fmt.Errorf("error checking if dependency exists, %v", err)
	}
	if !dependencyExists {
		RunPull(cli, dependencyImage, true, "", "")
	}
	// Create temp directory
	currentTime := time.Now()
	timestamp := currentTime.Format("27065102350415")
	tempPath := filepath.Join(cli.FileSystem().TempDir(), timestamp)
	if err = util.CreateDir(tempPath); err != nil {
		return nil, fmt.Errorf("error while creating temp directory, %v", err)
	}
	// Unzipping Cellery Image
	if err = util.Unzip(cellImageZip, tempPath); err != nil {
		return nil, fmt.Errorf("error while unzipping cell image zip, %v", err)
	}
	// Reading the dependency's metadata
	var metadataJsonContent []byte
	if metadataJsonContent, err = ioutil.ReadFile(filepath.Join(tempPath, artifacts, constants.CELLERY, "metadata.json")); err != nil {
		return nil, fmt.Errorf("metadata.json file not found for dependency: %s, %v", dependencyImage,
			err)
	}
	dependencyMetadata = &image.MetaData{}
	if err = json.Unmarshal(metadataJsonContent, dependencyMetadata); err != nil {
		return nil, fmt.Errorf("error while unmarshalling metadata json content of dependency, %v", err)
	}
	// Cleaning up
	if err = os.RemoveAll(tempPath); err != nil {
		return nil, fmt.Errorf("error while cleaning up temp directory, %v", err)
	}
	return dependencyMetadata, nil
}

func createTempBalFile(fileName string, tmpProjectDir string) (string, error) {
	var err error
	var tempBuildFileName string
	// First clean target directory if exists
	targetDir := filepath.Join(tmpProjectDir, "target")
	_ = os.Remove(targetDir)
	if tempBuildFileName, err = util.CreateTempExecutableBalFile(fileName, "build", tmpProjectDir); err != nil {
		return "", fmt.Errorf("error creating executable bal file: %v", err)
	}
	return tempBuildFileName, nil
}

func createArtifactsZip(artifactsZip, projectDir, projectSrc string) (string, error) {
	var err error
	targetDir := filepath.Join(projectDir, "target")
	imgDir := filepath.Join(projectDir, "zip")

	if err = util.CopyDir(filepath.Join(targetDir, constants.CELLERY), filepath.Join(imgDir, artifacts, constants.CELLERY)); err != nil {
		return "", fmt.Errorf("error occurred copying artifacts directory, %v", err)
	}
	if err = util.CopyDir(filepath.Join(targetDir, constants.Ref), filepath.Join(imgDir, artifacts, constants.Ref)); err != nil {
		return "", fmt.Errorf("error occurred copying artifacts directory, %v", err)
	}
	if err = util.CreateDir(filepath.Join(imgDir, src)); err != nil {
		return "", fmt.Errorf("error occurred while creating src directory, %v", err)
	}
	info, err := os.Stat(projectSrc)
	if err != nil {
		return "", fmt.Errorf("error occured while getting fileInfo of cell project, %v", err)
	}
	if info.IsDir() {
		if err = util.CopyDir(projectSrc, filepath.Join(imgDir, src, filepath.Base(projectSrc))); err != nil {
			return "", fmt.Errorf("error occurred copying bal project to src directory, %v", err)
		}
	} else {
		if err = util.CopyFile(projectSrc, filepath.Join(imgDir, src, filepath.Base(projectSrc))); err != nil {
			return "", fmt.Errorf("error occurred copying bal file to src directory, %v", err)
		}
	}
	folders := []string{filepath.Join(imgDir, artifacts), filepath.Join(imgDir, src)}

	// Todo: Check if WorkingDirRelativePath could be omitted.
	// For actual scenario WorkingDirRelativePath == ""
	// However, since the current dir is different to the running location, exact path has to be provided when running unit tests.
	if err = util.RecursiveZip(nil, folders, filepath.Join(imgDir, artifactsZip)); err != nil {
		return "", fmt.Errorf("error occurred while creating the image, %v", err)
	}
	return filepath.Join(imgDir, artifactsZip), nil
}

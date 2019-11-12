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
 *
 */

package image

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

// ParseImageTag parses the given image name string and returns a CellImage struct with the relevant information.
func ParseImageTag(cellImageString string) (parsedCellImage *CellImage, err error) {
	cellImage := &CellImage{
		constants.CentralRegistryHost,
		"",
		"",
		"",
	}

	if cellImageString == "" {
		return cellImage, errors.New("no cell image specified")
	}

	const IMAGE_FORMAT_ERROR_MESSAGE = "incorrect image name format. Image name should be " +
		"[REGISTRY[:REGISTRY_PORT]/]ORGANIZATION/IMAGE_NAME:VERSION"

	// Parsing the cell image string
	strArr := strings.Split(cellImageString, "/")
	if len(strArr) == 3 {
		cellImage.Registry = strArr[0]
		cellImage.Organization = strArr[1]
		imageTag := strings.Split(strArr[2], ":")
		if len(imageTag) != 2 {
			return cellImage, errors.New(IMAGE_FORMAT_ERROR_MESSAGE)
		}
		cellImage.ImageName = imageTag[0]
		cellImage.ImageVersion = imageTag[1]
	} else if len(strArr) == 2 {
		cellImage.Organization = strArr[0]
		imageNameSplit := strings.Split(strArr[1], ":")
		if len(imageNameSplit) != 2 {
			return cellImage, errors.New(IMAGE_FORMAT_ERROR_MESSAGE)
		}
		cellImage.ImageName = imageNameSplit[0]
		cellImage.ImageVersion = imageNameSplit[1]
	} else {
		return cellImage, errors.New(IMAGE_FORMAT_ERROR_MESSAGE)
	}

	return cellImage, nil
}

// ValidateImageTag validates the image tag (without the registry in it).
func ValidateImageTag(imageTag string) error {
	r := regexp.MustCompile("^([^/:]*)/([^/:]*):([^/:]*)$")
	subMatch := r.FindStringSubmatch(imageTag)

	if subMatch == nil {
		return fmt.Errorf("expects <organization>/<cell-image>:<version> as the tag, received %s", imageTag)
	}

	organization := subMatch[1]
	isValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), organization)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid organization name (lower case letters, numbers and dashes "+
			"with only letters and numbers at the begining and end), received %s", organization)
	}

	imageName := subMatch[2]
	isValid, err = regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), imageName)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid image name (lower case letters, numbers and dashes "+
			"with only letters and numbers at the begining and end), received %s", imageName)
	}

	imageVersion := subMatch[3]
	isValid, err = regexp.MatchString(fmt.Sprintf("^%s$", constants.ImageVersionPattern), imageVersion)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid image version (lower case letters, numbers, dashes and dots "+
			"with only letters and numbers at the begining and end), received %s", imageVersion)
	}

	return nil
}

// ValidateImageTag validates the image tag (with the registry in it). The registry is an option element
// in this validation.
func ValidateImageTagWithRegistry(imageTag string) error {
	r := regexp.MustCompile("^(?:([^/]*)/)?([^/:]*)/([^/:]*):([^/:]*)$")
	subMatch := r.FindStringSubmatch(imageTag)

	if subMatch == nil {
		return fmt.Errorf("expects [<registry>]/<organization>/<cell-image>:<version> as the tag, received %s",
			imageTag)
	}

	registry := subMatch[1]
	isValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.DomainNamePattern), registry)
	if registry != "" && (err != nil || !isValid) {
		return fmt.Errorf("expects a valid URL as the registry, received %s", registry)
	}

	organization := subMatch[2]
	isValid, err = regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), organization)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid organization name (lower case letters, numbers and dashes "+
			"with only letters and numbers at the begining and end), received %s", organization)
	}

	imageName := subMatch[3]
	isValid, err = regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), imageName)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid image name (lower case letters, numbers and dashes "+
			"with only letters and numbers at the begining and end), received %s", imageName)
	}

	imageVersion := subMatch[4]
	isValid, err = regexp.MatchString(fmt.Sprintf("^%s$", constants.ImageVersionPattern), imageVersion)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid image version (lower case letters, numbers, dashes and dots "+
			"with only letters and numbers at the begining and end), received %s", imageVersion)
	}

	return nil
}

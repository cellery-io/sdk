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

package test

import (
	"bytes"
	"io"

	"github.com/cellery-io/sdk/components/cli/pkg/image"
)

type MockRegistry struct {
	out       io.Writer
	outBuffer *bytes.Buffer
	images map[string][]byte
}

func NewMockRegistry(opts ...func(*MockRegistry)) *MockRegistry {
	outBuffer := new(bytes.Buffer)
	registry := &MockRegistry{
		out:       outBuffer,
		outBuffer: outBuffer,
	}
	for _, opt := range opts {
		opt(registry)
	}
	return registry
}

func SetImages(images map[string][]byte) func(*MockRegistry) {
	return func(registry *MockRegistry) {
		registry.images = images
	}
}

func (registry *MockRegistry) Push(parsedCellImage *image.CellImage, fileBytes []byte, username, password string) error {
	return nil
}

func (registry *MockRegistry) Pull(parsedCellImage *image.CellImage, username string, password string) ([]byte, error) {
	imageName := parsedCellImage.Organization + "/" + parsedCellImage.ImageName + ":" + parsedCellImage.ImageVersion
	return registry.images[imageName], nil
}

// Out returns the mock writer used for the stdout.
func (registry *MockRegistry) Out() io.Writer {
	return registry.out
}

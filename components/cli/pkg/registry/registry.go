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

package registry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"io/ioutil"

	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/libtrust"
	registry2 "github.com/nokia/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"

	"github.com/cellery-io/sdk/components/cli/pkg/image"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

type Registry interface {
	Pull(parsedCellImage *image.CellImage, username string, password string) ([]byte, error)
	Push(parsedCellImage *image.CellImage, fileBytes []byte, username, password string) error
	Out() io.Writer
}

type CelleryRegistry struct {
	hub *registry2.Registry
}

func NewCelleryRegistry() *CelleryRegistry {
	registry := &CelleryRegistry{}
	return registry
}

func (registry *CelleryRegistry) Push(parsedCellImage *image.CellImage, fileBytes []byte, username, password string) error {
	fmt.Fprintln(registry.Out(), fmt.Sprintf("\nConnecting to %s", util.Bold(parsedCellImage.Registry)))
	// Initiating a connection to Cellery Registry
	hub, err := registry2.New("https://"+parsedCellImage.Registry, username, password)
	if err != nil {
		return fmt.Errorf("failed to initialize connection to Cellery Registry %v", err)
	}
	imageName := fmt.Sprintf("%s/%s:%s", parsedCellImage.Organization, parsedCellImage.ImageName,
		parsedCellImage.ImageVersion)
	repository := parsedCellImage.Organization + "/" + parsedCellImage.ImageName
	// Creating the Cell Image Digest (Docker file Layer digest)
	hash := sha256.New()
	hash.Write(fileBytes)
	sha256sum := hex.EncodeToString(hash.Sum(nil))
	cellImageDigest := digest.NewDigestFromEncoded(digest.SHA256, sha256sum)

	// Checking if the the Cell Image already exists in the registry
	fmt.Fprintln(registry.Out(), fmt.Sprintf("\nChecking if the image %s already exists in the Registry", util.Bold(imageName)))
	cellImageDigestExists, err := hub.HasBlob(repository, cellImageDigest)
	if err != nil {
		return err
	}

	// Pushing the cell image if it is not already uploaded
	if !cellImageDigestExists {
		fmt.Fprintln(registry.Out(), fmt.Sprintf("\nPushing image %s", util.Bold(imageName)))
		// Read stream of files
		err = hub.UploadBlob(repository, cellImageDigest, bytes.NewReader(fileBytes), nil)
		if err != nil {
			return err
		}
	} else {
		fmt.Fprintln(registry.Out(), fmt.Sprintf("\nUsing already existing image %s in %s Registry", util.Bold(imageName),
			util.Bold(parsedCellImage.Registry)))
	}

	// Creating a Docker manifest to be uploaded
	cellImageManifest := &schema1.Manifest{
		Name: repository,
		Versioned: manifest.Versioned{
			SchemaVersion: 1,
			MediaType:     schema1.MediaTypeSignedManifest,
		},
		Tag:          parsedCellImage.ImageVersion,
		Architecture: "amd64",
		FSLayers: []schema1.FSLayer{
			{BlobSum: cellImageDigest},
		},
		History: []schema1.History{
			{},
		},
	}
	// Signing the Docker Manifest
	key, err := libtrust.GenerateECP256PrivateKey()
	if err != nil {
		return fmt.Errorf("error occurred while pushing the cell image, %v", err)
	}
	signedCellImageManifest, err := schema1.Sign(cellImageManifest, key)
	if err != nil {
		return fmt.Errorf("error occurred while pushing the cell image, %v", err)
	}

	// Uploading the manifest to the Cellery Registry (Docker Registry)
	err = hub.PutManifest(repository, parsedCellImage.ImageVersion, signedCellImageManifest)
	if err != nil {
		return err
	}
	fmt.Fprintln(registry.Out(), fmt.Sprintf("\nImage Digest : %s\n", util.Bold(cellImageDigest)))
	return nil
}

func (registry *CelleryRegistry) Pull(parsedCellImage *image.CellImage, username string, password string) ([]byte, error) {
	var cellImage []byte
	repository := parsedCellImage.Organization + "/" + parsedCellImage.ImageName
	// Initiating a connection to Cellery Registry
	hub, err := registry2.New("https://"+parsedCellImage.Registry, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connection to Cellery Registry %v", err)
	}
	// Fetching the Docker Image Manifest
	cellImageManifest, err := hub.Manifest(repository, parsedCellImage.ImageVersion)
	if err != nil {
		return nil, err
	}
	var cellImageDigest digest.Digest
	if len(cellImageManifest.References()) == 1 {
		cellImageReference := cellImageManifest.References()[0]
		cellImageDigest = cellImageReference.Digest

		imageName := fmt.Sprintf("%s/%s:%s", parsedCellImage.Organization, parsedCellImage.ImageName,
			parsedCellImage.ImageVersion)
		fmt.Fprintln(registry.Out(), fmt.Sprintf("\nPulling image %s", util.Bold(imageName)))

		// Downloading the Cell Image from the repository
		reader, err := hub.DownloadBlob(repository, cellImageReference.Digest)
		if err != nil {
			return nil, err
		}
		if reader != nil {
			defer func() error {
				err = reader.Close()
				if err != nil {
					return fmt.Errorf("error occurred while cleaning up, %v", err)
				}
				return nil
			}()
		}
		cellImage, err = ioutil.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("error occurred while pulling cell image, %v", err)
		}

	} else {
		return nil, fmt.Errorf("invalid cell image, %v",
			errors.New(fmt.Sprintf("expected exactly 1 File Layer, but found %d",
				len(cellImageManifest.References()))))
	}
	fmt.Fprintln(registry.Out(), fmt.Sprintf("\nImage Digest : %s\n", util.Bold(cellImageDigest)))
	return cellImage, nil
}

// Out returns the writer used for the stdout.
func (registry *CelleryRegistry) Out() io.Writer {
	return os.Stdout
}

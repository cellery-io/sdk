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

package gcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/file/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/util"
)

type Gcp struct {
	uuid          string
	ctx           context.Context
	service       *container.Service
	accountName   string
	zone          string
	region        string
	projectName   string
	clusterName   string
	sqlAccount    string
	sqlService    *sqladmin.Service
	nfsService    *file.Service
	storageClient *storage.Client
	bucketName    string
}

// NewGcp returns a Gcp instance.
func NewGcp(opts ...func(*Gcp)) (*Gcp, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	uuid := strconv.Itoa(rand.Intn(1000))
	gcp := &Gcp{
		uuid:        uuid,
		clusterName: clusterNamePrefix + uuid,
		bucketName:  storagePrefix + uuid,
	}
	if err := gcp.authenticateCredentials(); err != nil {
		return nil, fmt.Errorf("failed to configure gcp credentials")
	}
	gcp.ctx = context.Background()
	// Create a GKE client
	gkeClient, err := google.DefaultClient(gcp.ctx, container.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("error creating gcp client: %v", err)
	}
	gcp.service, err = container.New(gkeClient)
	if err != nil {
		return nil, fmt.Errorf("error creating gcp service: %v", err)
	}
	hcSql, err := google.DefaultClient(gcp.ctx, sqladmin.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("error creating sql client: %v", err)
	}
	gcp.sqlService, err = sqladmin.New(hcSql)
	if err != nil {
		return nil, fmt.Errorf("error creating sql service")
	}
	hcNfs, err := google.DefaultClient(gcp.ctx, file.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("error creating nfs client: %v", err)
	}
	gcp.nfsService, err = file.New(hcNfs)
	if err != nil {
		return nil, fmt.Errorf("error creating nfs service: %v", err)
	}
	gcp.storageClient, err = storage.NewClient(gcp.ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating storage client: %v", err)
	}
	if err := gcp.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize gcp")
	}
	if gcp.region == "" {
		return nil, fmt.Errorf("error creating cluster, %v", fmt.Errorf("region not found in gcloud "+
			"config list. Please run `gcloud init` to set the region"))
	}
	if gcp.zone == "" {
		return nil, fmt.Errorf("error creating cluster, %v", fmt.Errorf("zone not found in gcloud "+
			"config list. Please run `gcloud init` to set the zone"))
	}
	for _, opt := range opts {
		opt(gcp)
	}
	return gcp, nil
}

func SetUuid(uuid string) func(*Gcp) {
	return func(gcp *Gcp) {
		gcp.uuid = uuid
	}
}

func SetClusterName(clusterName string) func(*Gcp) {
	return func(gcp *Gcp) {
		gcp.clusterName = clusterName
	}
}

func (gcp *Gcp) initialize() error {
	cmd := exec.Command("gcloud", "config", "list", "--format", "json")
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	output := ""
	go func() {
		for stdoutScanner.Scan() {
			output = output + stdoutScanner.Text()
		}
	}()
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)

	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error occurred while starting to get gcp data, %v", err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error occurred while getting gcp data, %v", err)
	}
	jsonOutput := &util.Gcp{}
	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson != nil {
		return fmt.Errorf("failed to unmarshall, %v", errJson)
	}
	gcp.projectName = jsonOutput.Core.Project
	gcp.accountName = jsonOutput.Core.Account
	gcp.region = jsonOutput.Compute.Region
	gcp.zone = jsonOutput.Compute.Zone
	return nil
}

func (gcp *Gcp) authenticateCredentials() error {
	jsonAuthFile := util.FindInDirectory(filepath.Join(util.UserHomeDir(), constants.CelleryHome, constants.GCP),
		".json")
	if len(jsonAuthFile) > 0 {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", jsonAuthFile[0])
	} else {
		return fmt.Errorf("failed to create gcp cluster, %v", fmt.Errorf(
			"Could not find authentication json file in : %s. Please copy GCP service account credentials"+
				" json file into this directory.\n", filepath.Join(util.UserHomeDir(),
				constants.CelleryHome, constants.GCP)))
	}
	return nil
}

func (gcp *Gcp) gcpClusterExist() (bool, error) {
	clusters, err := gcp.GetClusterList()
	if err != nil {
		return false, err
	}
	if len(clusters) == 0 {
		return false, nil
	}
	if util.ContainsInStringArray(clusters, clusterNamePrefix+gcp.uuid) {
		return true, nil
	}
	return false, nil
}

func (gcp *Gcp) GetClusterList() ([]string, error) {
	var clusters []string
	list, err := gcp.service.Projects.Zones.Clusters.List(gcp.projectName, gcp.zone).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %v", err)
	}
	for _, cluster := range list.Clusters {
		clusters = append(clusters, cluster.Name)
	}
	return clusters, nil
}

func (gcp *Gcp) ClusterName() string {
	return gcp.clusterName
}

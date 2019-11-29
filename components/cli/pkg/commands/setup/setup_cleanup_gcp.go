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

package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"

	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/gcp"
	"cellery.io/cellery/components/cli/pkg/util"
)

func manageGcp() error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select `cleanup` to remove an existing GCP cluster",
		Items:     []string{constants.CelleryManageCleanup, constants.CellerySetupBack},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select an option: %v", err)
	}

	switch value {
	case constants.CelleryManageCleanup:
		{
			cleanupGcp()
		}
	default:
		{
			manageEnvironment()
		}
	}
	return nil
}

func cleanupGcp() error {
	jsonAuthFile := util.FindInDirectory(filepath.Join(util.UserHomeDir(), constants.CelleryHome, constants.GCP),
		".json")

	if len(jsonAuthFile) > 0 {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", jsonAuthFile[0])
	} else {
		util.ExitWithErrorMessage("Failed to cleanup gcp setup", fmt.Errorf("Could not find "+
			"authentication json file in : %s. Please copy GCP service account credentials"+
			" json file into this directory.\n", filepath.Join(util.UserHomeDir(), constants.CelleryHome,
			constants.GCP)))
	}
	ctx := context.Background()
	projectName, accountName, region, zone = getGcpData()
	gkeClient, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		fmt.Printf("Failed to create gke client: %v", err)
	}

	gcpService, err := container.New(gkeClient)
	if err != nil {
		fmt.Printf("Failed to create gcp service: %v", err)
	}

	clusters, err := getClusterList(gcpService, projectName, zone)
	if err != nil {
		fmt.Printf("Failed to list clusters: %v", err)
	}
	userCreatedGcpClusters := getGcpClustersCreatedByUser(clusters)

	selectTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select a GCP cluster to delete",
		Items:     append(userCreatedGcpClusters, constants.CellerySetupBack),
		Templates: selectTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option: %v", err)
	}
	if value == constants.CellerySetupBack {
		manageGcp()
		return nil
	}
	RunCleanupGcp(value)
	return nil
}

func ValidateGcpCluster(cluster string) (bool, error) {
	valid := false
	jsonAuthFile := util.FindInDirectory(filepath.Join(util.UserHomeDir(), constants.CelleryHome, constants.GCP), ".json")

	if len(jsonAuthFile) > 0 {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", jsonAuthFile[0])
	} else {
		return false, fmt.Errorf("Could not find authentication json file in : %s. Please copy GCP service account credentials"+
			" json file into this directory.\n", filepath.Join(util.UserHomeDir(), constants.CelleryHome, constants.GCP))
	}
	ctx := context.Background()
	gkeClient, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		return false, fmt.Errorf("failed to create gke client: %v", err)
	}
	gcpService, err := container.New(gkeClient)
	if err != nil {
		return false, fmt.Errorf("failed to create gcp service: %v", err)
	}
	projectName, accountName, region, zone = getGcpData()
	clusters, err := getClusterList(gcpService, projectName, zone)
	if err != nil {
		return false, fmt.Errorf("failed to list clusters: %v", err)
	}
	clusterNameSlice := strings.Split(cluster, constants.GcpClusterName)
	if len(clusterNameSlice) > 1 {
		uniqueNumber := strings.Split(cluster, constants.GcpClusterName)[1]
		if util.ContainsInStringArray(clusters, constants.GcpClusterName+uniqueNumber) {
			valid = true
		}
	}
	return valid, nil
}

func RunCleanupGcp(value string) error {
	cleanupSpinner := util.StartNewSpinner("Removing GCP cluster")
	platform, err := gcp.NewGcp(gcp.SetClusterName(value))
	if err != nil {
		return fmt.Errorf("failed to initialize gcp platform, %v", err)
	}
	err = platform.TearDown()
	if err != nil {
		cleanupSpinner.Stop(false)
	}
	return nil
}

func getClusterList(service *container.Service, projectID, zone string) ([]string, error) {
	var clusters []string
	list, err := service.Projects.Zones.Clusters.List(projectID, zone).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %v", err)
	}

	for _, cluster := range list.Clusters {
		clusters = append(clusters, cluster.Name)
	}
	return clusters, nil
}

func getGcpClustersCreatedByUser(gcpClusters []string) []string {
	var userCreatedGcpClusters []string
	// Get the list of clusters created by the user
	userCreatedClusters := getContexts()
	for _, gcpCluster := range gcpClusters {
		for _, userCreatedCluster := range userCreatedClusters {
			// Check if the given gcp cluster is created by the user
			if strings.Contains(userCreatedCluster, gcpCluster) {
				// If the gcp cluster is created by user add it to the list of gcp clusters created by the user
				userCreatedGcpClusters = append(userCreatedGcpClusters, gcpCluster)
				break
			}
		}
	}
	return userCreatedGcpClusters
}

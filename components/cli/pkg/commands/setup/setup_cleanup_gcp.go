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
	"time"

	"cloud.google.com/go/storage"
	"github.com/manifoldco/promptui"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/file/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
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
	gkeClient, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		fmt.Printf("Failed to create gke client: %v", err)
	}
	gcpService, err := container.New(gkeClient)
	if err != nil {
		fmt.Printf("Failed to create gcp service: %v", err)
	}
	// Get the suffix - unique number of GCP cluster which is common to all infrastructures (sql instance, filestore, bucket)
	uniqueNumber := strings.TrimPrefix(value, "cellery-cluster")

	cleanupSpinner := util.StartNewSpinner("Removing GCP cluster")

	// Delete GCP cluster
	errCleanup := deleteGcpCluster(gcpService, projectName, zone, constants.GcpClusterName+uniqueNumber)
	if errCleanup != nil {
		cleanupSpinner.Stop(false)
		fmt.Printf("Failed to cleanup GCP cluster: %v", errCleanup)
	}
	for i := 0; i < 15; i++ {
		if gcpClusterExist(gcpService, projectName, zone, constants.GcpClusterName+uniqueNumber) {
			time.Sleep(60 * time.Second)
		} else {
			break
		}
	}
	// Delete sql instance
	cleanupSpinner.SetNewAction("Deleting sql instance")
	hcSql, err := google.DefaultClient(ctx, sqladmin.CloudPlatformScope)

	if err != nil {
		cleanupSpinner.Stop(false)
		fmt.Printf("Failed to cleanup GCP cluster: %v", err)
	}

	sqlService, err := sqladmin.New(hcSql)
	if err != nil {
		cleanupSpinner.Stop(false)
		fmt.Printf("Failed to cleanup GCP cluster: %v", err)
	}

	errorDeleteSql := deleteSqlInstance(sqlService, projectName, constants.GcpDbInstanceName+uniqueNumber)
	if errorDeleteSql != nil {
		cleanupSpinner.Stop(false)
		fmt.Printf("Failed to cleanup GCP cluster: %v", errorDeleteSql)
	}

	// Delete nfs file server
	cleanupSpinner.SetNewAction("Deleting File store")
	hcNfs, err := google.DefaultClient(ctx, file.CloudPlatformScope)
	if err != nil {
		cleanupSpinner.Stop(false)
		fmt.Printf("Failed to cleanup GCP cluster: %v", err)
	}

	nfsService, err := file.New(hcNfs)
	if err != nil {
		cleanupSpinner.Stop(false)
		fmt.Printf("Failed to cleanup GCP cluster: %v", err)
	}
	errDeleteFileStore := deleteFileStore(nfsService, projectName, zone, constants.GcpNfsServerInstance+uniqueNumber)
	if errDeleteFileStore != nil {
		fmt.Printf("Failed to cleanup GCP cluster: %v", errDeleteFileStore)
	}

	// Delete GCP storage
	cleanupSpinner.SetNewAction("Deleting Storage")
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("Error creating storage client: %v", err)
	}

	errorDeleteStorageObject := deleteGcpStorageObject(storageClient, constants.GcpBucketName+uniqueNumber, constants.InitSql)
	if err != nil {
		fmt.Printf("Error deleting gcp storage object: %v", errorDeleteStorageObject)
	}

	errorDeleteStorage := deleteGcpStorage(storageClient, constants.GcpBucketName+uniqueNumber)
	if err != nil {
		fmt.Printf("Error deleting gcp storage: %v", errorDeleteStorage)
	}

	cleanupSpinner.Stop(true)
	return nil
}

func deleteGcpCluster(gcpService *container.Service, projectID string, zone string, clusterID string) error {
	_, err := gcpService.Projects.Zones.Clusters.Delete(projectID, zone, clusterID).Do()
	if err != nil {
		return fmt.Errorf("failed to delete gcp cluster: %v", err)

	}
	return nil
}

func deleteSqlInstance(service *sqladmin.Service, projectId string, instanceName string) error {
	_, err := service.Instances.Delete(projectId, instanceName).Do()
	if err != nil {
		return fmt.Errorf("failed to delete the sql instance: %v", err)
	}
	return nil
}

func deleteFileStore(service *file.Service, projectName, zone, storeName string) error {
	_, err := service.Projects.Locations.Instances.Delete("projects/" + projectName + "/locations/" + zone + "/instances/" + storeName).Do()
	if err != nil {
		return fmt.Errorf("failed to delete nfs server %v", err)
	}
	return nil
}

func deleteGcpStorage(client *storage.Client, bucketName string) error {
	ctx := context.Background()
	if err := client.Bucket(bucketName).Delete(ctx); err != nil {
		return err
	}
	return nil
}

func deleteGcpStorageObject(client *storage.Client, bucketName, objectName string) error {
	ctx := context.Background()
	object := client.Bucket(bucketName).Object(objectName)
	if err := object.Delete(ctx); err != nil {
		return err
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

func gcpClusterExist(service *container.Service, projectID, zone, clusterName string) bool {
	clusters, err := getClusterList(service, projectID, zone)
	if err != nil {
		fmt.Printf("Failed to list clusters: %v", err)
	}
	if len(clusters) == 0 {
		return false
	}
	if util.ContainsInStringArray(clusters, clusterName) {
		return true
	}
	return false
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

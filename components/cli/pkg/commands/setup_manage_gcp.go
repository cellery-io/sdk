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
		Items:     []string{constants.CELLERY_MANAGE_CLEANUP, constants.CELLERY_SETUP_BACK},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("Failed to select an option: %v", err)
	}

	switch value {
	case constants.CELLERY_MANAGE_CLEANUP:
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
	jsonAuthFile := util.FindInDirectory(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP), ".json")

	if len(jsonAuthFile) > 0 {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", jsonAuthFile[0])
	} else {
		fmt.Printf("Could not find authentication json file in : %v", filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP))
		os.Exit(1)
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

	selectTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select a GCP cluster to delete",
		Items:     append(clusters, constants.CELLERY_SETUP_BACK),
		Templates: selectTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option: %v", err)
	}
	if value == constants.CELLERY_SETUP_BACK {
		manageGcp()
		return nil
	}

	// Get the suffix - unique number of GCP cluster which is common to all infrastructures (sql instance, filestore, bucket)
	uniqueNumber := strings.TrimPrefix(value, "cellery-cluster")

	cleanupSpinner := util.StartNewSpinner("Removing GCP cluster")

	// Delete GCP cluster
	errCleanup := deleteGcpCluster(gcpService, projectName, zone, constants.GCP_CLUSTER_NAME+uniqueNumber)
	if errCleanup != nil {
		cleanupSpinner.Stop(false)
		fmt.Printf("Failed to cleanup GCP cluster: %v", errCleanup)
	}
	for i := 0; i < 15; i++ {
		if gcpClusterExist(gcpService, projectName, zone, constants.GCP_CLUSTER_NAME+uniqueNumber) {
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

	errorDeleteSql := deleteSqlInstance(sqlService, projectName, constants.GCP_DB_INSTANCE_NAME+uniqueNumber)
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
	errDeleteFileStore := deleteFileStore(nfsService, projectName, zone, constants.GCP_NFS_SERVER_INSTANCE+uniqueNumber)
	if errDeleteFileStore != nil {
		fmt.Printf("Failed to cleanup GCP cluster: %v", errDeleteFileStore)
	}

	// Delete GCP storage
	cleanupSpinner.SetNewAction("Deleting Storage")
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("Error creating storage client: %v", err)
	}

	errorDeleteStorageObject := deleteGcpStorageObject(storageClient, constants.GCP_BUCKET_NAME+uniqueNumber, constants.INIT_SQL)
	if err != nil {
		fmt.Printf("Error deleting gcp storage object: %v", errorDeleteStorageObject)
	}

	errorDeleteStorage := deleteGcpStorage(storageClient, constants.GCP_BUCKET_NAME+uniqueNumber)
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

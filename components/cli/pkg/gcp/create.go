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
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/file/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
)

func (gcp *Gcp) CreateK8sCluster() error {
	gcpGKENode := &container.NodeConfig{
		ImageType:   clusterImageType,
		MachineType: clusterMachineType,
		DiskSizeGb:  clusterDiskSizeGb,
	}
	gcpCluster := &container.Cluster{
		Name:                  gcp.clusterName,
		Description:           "k8s cluster",
		InitialClusterVersion: "1.14",
		InitialNodeCount:      1,
		Location:              gcp.zone,
		NodeConfig:            gcpGKENode,
		ResourceLabels:        gcp.getCreatedByLabel(),
	}
	createClusterRequest := &container.CreateClusterRequest{
		ProjectId: gcp.projectName,
		Zone:      gcp.zone,
		Cluster:   gcpCluster,
	}

	_, err := gcp.service.Projects.Zones.Clusters.Create(gcp.projectName, gcp.zone, createClusterRequest).Do()

	// Loop until cluster status becomes RUNNING
	for i := 0; i < 30; i++ {
		resp, err := gcp.service.Projects.Zones.Clusters.Get(gcp.projectName, gcp.zone, gcp.clusterName).Do()
		if err != nil {
			time.Sleep(30 * time.Second)
		} else {
			if resp.Status == "RUNNING" {
				return nil
			}
			time.Sleep(15 * time.Second)
		}
	}
	// Give permission to the user
	if err := kubernetes.CreateClusterRoleBinding("cluster-admin", gcp.accountName); err != nil {
		return fmt.Errorf("error creating cluster role binding, %v", err)
	}
	return fmt.Errorf("failed to create clusters: %v", err)
}

func (gcp *Gcp) getCreatedByLabel() map[string]string {
	// Add a label to identify the user who created it
	labels := make(map[string]string)
	createdBy := util.ConvertToAlphanumeric(gcp.accountName, "_")
	labels["created_by"] = createdBy
	return labels
}

func (gcp *Gcp) ConfigureSqlInstance() (runtime.MysqlDb, error) {
	//Create sql instance
	_, err := gcp.createSqlInstance(gcp.sqlService, gcp.projectName, gcp.region, gcp.region, dbInstanceNamePrefix+uuid)
	if err != nil {
		return runtime.MysqlDb{}, fmt.Errorf("error creating sql instance: %v", err)
	}
	sqlIpAddress, serviceAccountEmailAddress := gcp.getSqlServiceAccount(gcp.ctx, gcp.sqlService, gcp.projectName,
		dbInstanceNamePrefix+uuid)
	gcp.sqlAccount = serviceAccountEmailAddress
	return runtime.MysqlDb{DbHostName: sqlIpAddress, DbUserName: sqlUserName, DbPassword: sqlPassword + uuid}, nil
}

func (gcp *Gcp) createSqlInstance(service *sqladmin.Service, projectId string, region string, zone string,
	instanceName string) ([]*sqladmin.DatabaseInstance, error) {
	settings := &sqladmin.Settings{
		DataDiskSizeGb: sqlDiskSizeGb,
		Tier:           sqlTier,
		UserLabels:     gcp.getCreatedByLabel(),
	}

	dbInstance := &sqladmin.DatabaseInstance{
		Name:            instanceName,
		Project:         projectId,
		Region:          region,
		GceZone:         zone,
		BackendType:     "SECOND_GEN",
		DatabaseVersion: "MYSQL_5_7",
		Settings:        settings,
	}

	if _, err := service.Instances.Insert(projectId, dbInstance).Do(); err != nil {
		return nil, fmt.Errorf("error in sql instance creation, %v", err)
	}
	return nil, nil
}

func (gcp *Gcp) getSqlServiceAccount(ctx context.Context, gcpService *sqladmin.Service, projectId string, instanceName string) (string, string) {
	for i := 0; i < 30; i++ {
		resp, err := gcpService.Instances.Get(projectId, instanceName).Context(ctx).Do()
		if err != nil {
			time.Sleep(60 * time.Second)
		} else {
			if resp.State == "RUNNABLE" {
				return resp.IpAddresses[0].IpAddress, resp.ServiceAccountEmailAddress
			}
			time.Sleep(15 * time.Second)
		}
	}
	return "", ""
}

func (gcp *Gcp) CreateStorage() error {
	// Create bucketName
	if err := gcp.createGcpStorage(); err != nil {
		return fmt.Errorf("error creating storage client: %v", err)
	}
	// Upload init file to S3 bucketName
	if err := gcp.uploadSqlFile(gcp.storageClient, gcp.bucketName, initSql); err != nil {
		return fmt.Errorf("error Uploading Sql file: %v", err)
	}
	// Update bucketName permission
	if err := gcp.updateBucketPermission(gcp.bucketName, gcp.sqlAccount); err != nil {
		return fmt.Errorf("error Updating bucketName permission: %v", err)
	}
	// Import sql script
	var uri = "gs://" + gcp.bucketName + "/init.sql"
	if err := gcp.importSqlScript(gcp.sqlService, gcp.projectName, dbInstanceNamePrefix+uuid, uri); err != nil {
		return fmt.Errorf("error Updating bucketName permission: %v", err)
	}
	time.Sleep(30 * time.Second)
	// Update sql instance
	if err := gcp.updateInstance(); err != nil {
		return fmt.Errorf("error Updating sql instance: %v", err)
	}
	return nil
}

func (gcp *Gcp) createGcpStorage() error {
	ctx := context.Background()
	attributes := &storage.BucketAttrs{
		Labels: gcp.getCreatedByLabel(),
	}
	if err := gcp.storageClient.Bucket(gcp.bucketName).Create(ctx, gcp.projectName, attributes); err != nil {
		return err
	}
	return nil
}

func (gcp *Gcp) uploadSqlFile(client *storage.Client, bucket, object string) error {
	ctx := context.Background()
	if err := updateInitSql(sqlUserName, sqlPassword+uuid); err != nil {
		return err
	}
	f, err := os.Open(filepath.Join(util.UserHomeDir(), constants.CelleryHome, constants.K8sArtifacts,
		constants.MySql, dbScripts, initSql))
	if err != nil {
		return err
	}
	defer f.Close()

	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}

func (gcp *Gcp) updateBucketPermission(bucketName string, svcAccount string) (err error) {
	scCtx := context.Background()
	c, err := storage.NewClient(scCtx)
	if err != nil {
		return err
	}

	ctx := context.Background()

	bucket := c.Bucket(bucketName)
	policy, err := bucket.IAM().Policy(ctx)
	if err != nil {
		return err
	}

	policy.Add(strings.Join([]string{"serviceAccount:", svcAccount}, ""), "roles/storage.objectViewer")
	if err := bucket.IAM().SetPolicy(ctx, policy); err != nil {
		return err
	}
	return nil
}

func (gcp *Gcp) importSqlScript(service *sqladmin.Service, projectId string, instanceName string, uri string) error {
	ic := &sqladmin.ImportContext{
		FileType: "SQL",
		Uri:      uri,
	}

	iIR := &sqladmin.InstancesImportRequest{
		ImportContext: ic,
	}

	_, err := service.Instances.Import(projectId, instanceName, iIR).Do()
	if err != nil {
		return err
	}
	return nil
}

func (gcp *Gcp) updateInstance() error {
	aclEntry := &sqladmin.AclEntry{
		Value: "0.0.0.0/0",
	}

	var aclEntryList []*sqladmin.AclEntry
	aclEntryList = append(aclEntryList, aclEntry)

	ipConfigs := &sqladmin.IpConfiguration{
		AuthorizedNetworks: aclEntryList,
	}

	newSettings := &sqladmin.Settings{
		IpConfiguration: ipConfigs,
	}

	newDbInstance := &sqladmin.DatabaseInstance{
		Settings: newSettings,
	}

	_, err := gcp.sqlService.Instances.Patch(gcp.projectName, dbInstanceNamePrefix+uuid, newDbInstance).Do()
	if err != nil {
		return fmt.Errorf("error updating instance :%v", err)
	}
	return nil
}

func (gcp *Gcp) CreateNfs() (runtime.Nfs, error) {
	// Create NFS server
	if err := gcp.createNfsServer(); err != nil {
		return runtime.Nfs{}, fmt.Errorf("error creating NFS server: %v", err)
	}
	nfsIpAddress, errIp := gcp.getNfsServerIp()
	if errIp != nil {
		return runtime.Nfs{}, fmt.Errorf("error getting NFS server IP address: %v", errIp)
	}
	return runtime.Nfs{NfsServerIp: nfsIpAddress, FileShare: "/data"}, nil
}

func (gcp *Gcp) createNfsServer() error {
	fileShare := &file.FileShareConfig{
		Name:       fileStoreFileShare,
		CapacityGb: fileStoreCapacity,
	}

	fileShares := append([]*file.FileShareConfig{}, fileShare)

	network := &file.NetworkConfig{
		Network: "default",
	}

	networks := append([]*file.NetworkConfig{}, network)

	nfsInstance := &file.Instance{
		FileShares: fileShares,
		Networks:   networks,
		Tier:       "STANDARD",
		Labels:     gcp.getCreatedByLabel(),
	}

	if _, err := gcp.nfsService.Projects.Locations.Instances.Create("projects/"+gcp.projectName+"/locations/"+gcp.zone,
		nfsInstance).InstanceId(fileStorePrefix + uuid).Do(); err != nil {
		return err
	}
	return nil
}

func (gcp *Gcp) getNfsServerIp() (string, error) {
	serverIp := ""
	for true {
		inst, err := gcp.nfsService.Projects.Locations.Instances.Get("projects/" + gcp.projectName + "/locations/" + gcp.zone +
			"/instances/" + fileStorePrefix + uuid).Do()
		if err != nil {
			return serverIp, err
		} else {
			if inst.State == "READY" {
				serverIp = inst.Networks[0].IpAddresses[0]
				return serverIp, nil
			} else {
				time.Sleep(10 * time.Second)
			}
		}
	}
	return serverIp, nil
}

func (gcp *Gcp) UpdateKubeConfig() error {
	cmd := exec.Command("bash", "-c", "gcloud container clusters get-credentials "+gcp.clusterName+" --zone "+gcp.zone+" --project "+gcp.projectName)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()

	go io.Copy(os.Stdout, stderr)
	return nil
}

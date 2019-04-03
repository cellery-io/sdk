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
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/manifoldco/promptui"
	"golang.org/x/oauth2/google"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"

	"google.golang.org/api/container/v1"
	"google.golang.org/api/file/v1"

	"context"
)

func RunSetupCreateGcp(isCompleteSetup bool) {
	if isCompleteSetup {
		createCompleteGcpRuntime()
	} else {
		createMinimalGcpRuntime()
	}
}

func createGcp() error {
	var isCompleteSelected = false
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select the type of runtime",
		Items:     []string{constants.BASIC, constants.COMPLETE},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("Failed to select an option: %v", err)
	}
	if value == constants.COMPLETE {
		isCompleteSelected = true
	}
	RunSetupCreateGcp(isCompleteSelected)
	return nil
}

func createMinimalGcpRuntime() {
	_ = configureGCPCredentials()
	ctx := context.Background()

	// Create a GKE client
	gcpSpinner := util.StartNewSpinner("Creating GKE client")
	createKubernentesClusterOnGcp(ctx, gcpSpinner)
	// Deploy cellery runtime
	gcpSpinner.SetNewAction("Deploying Cellery runtime")
	deployMinimalCelleryRuntime()
	gcpSpinner.Stop(true)
}

func createCompleteGcpRuntime() error {
	gcpBucketName := configureGCPCredentials()
	ctx := context.Background()

	// Create a GKE client
	gcpSpinner := util.StartNewSpinner("Creating GKE client")
	createKubernentesClusterOnGcp(ctx, gcpSpinner)

	sqlService, serviceAccountEmailAddress := configureMysqlOnGcp(ctx, gcpSpinner)

	configureBucketOnGcp(ctx, gcpSpinner, gcpBucketName, serviceAccountEmailAddress, sqlService)

	configureNfsOnGcp(gcpSpinner, ctx)

	// Deploy cellery runtime
	gcpSpinner.SetNewAction("Deploying Cellery runtime")
	deployCompleteCelleryRuntime()

	gcpSpinner.Stop(true)
	return nil
}

func configureNfsOnGcp(gcpSpinner *util.Spinner, ctx context.Context) {
	// Create NFS server
	gcpSpinner.SetNewAction("Creating NFS server")
	hcNfs, err := google.DefaultClient(ctx, file.CloudPlatformScope)
	if err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error getting authenticated client: %v", err)
	}
	nfsService, err := file.New(hcNfs)
	if err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error creating nfs service: %v", err)
	}
	if err := createNfsServer(nfsService); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error creating NFS server: %v", err)
	}
	nfsIpAddress, errIp := getNfsServerIp(nfsService)
	fmt.Printf("Nfs Ip address : %v", nfsIpAddress)
	if errIp != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error getting NFS server IP address: %v", errIp)
	}
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.GLOBAL_APIM, constants.ARTIFACTS_PERSISTENT_VOLUME_YAML), "NFS_SERVER_IP", nfsIpAddress, -1); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error replacing in file artifacts-persistent-volume.yaml: %v", err)
	}
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.GLOBAL_APIM, constants.ARTIFACTS_PERSISTENT_VOLUME_YAML), "NFS_SHARE_LOCATION", "/data", -1); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error replacing in file artifacts-persistent-volume.yaml: %v", err)
	}
}

func configureBucketOnGcp(ctx context.Context, gcpSpinner *util.Spinner, gcpBucketName string, serviceAccountEmailAddress string, sqlService *sqladmin.Service) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error creating storage client: %v", err)
	}
	// Create bucket
	gcpSpinner.SetNewAction("Creating gcp bucket")
	if err := createGcpStorage(client, projectName, gcpBucketName); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error creating storage client: %v", err)
	}
	// Upload init file to S3 bucket
	gcpSpinner.SetNewAction("Uploading init.sql file to dcp bucket")
	if err := uploadSqlFile(client, gcpBucketName, constants.INIT_SQL); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error Uploading Sql file: %v", err)
	}
	// Update bucket permission
	gcpSpinner.SetNewAction("Updating bucket permission")
	if err := updateBucketPermission(gcpBucketName, serviceAccountEmailAddress); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error Updating bucket permission: %v", err)
	}
	// Import sql script
	gcpSpinner.SetNewAction("Importing sql script")
	var uri = "gs://" + gcpBucketName + "/init.sql"
	if err := importSqlScript(sqlService, projectName, constants.GCP_DB_INSTANCE_NAME+uniqueNumber, uri); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error Updating bucket permission: %v", err)
	}
	time.Sleep(30 * time.Second)
	// Update sql instance
	gcpSpinner.SetNewAction("Updating sql instance")
	if err := updateInstance(sqlService, projectName, constants.GCP_DB_INSTANCE_NAME+uniqueNumber); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error Updating sql instance: %v", err)
	}
}

func configureMysqlOnGcp(ctx context.Context, gcpSpinner *util.Spinner) (*sqladmin.Service, string) {
	hcSql, err := google.DefaultClient(ctx, sqladmin.CloudPlatformScope)
	if err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error creating client: %v", err)
	}
	sqlService, err := sqladmin.New(hcSql)
	if err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error creating sql service: %v", err)
	}
	//Create sql instance
	gcpSpinner.SetNewAction("Creating sql instance")
	_, err = createSqlInstance(ctx, sqlService, projectName, region, region, constants.GCP_DB_INSTANCE_NAME+uniqueNumber)
	if err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error creating sql instance: %v", err)
	}
	sqlIpAddress, serviceAccountEmailAddress := getSqlServieAccount(ctx, sqlService, projectName, constants.GCP_DB_INSTANCE_NAME+uniqueNumber)
	fmt.Printf("Sql Ip address : %v", sqlIpAddress)
	// Replace username in /global-apim/conf/datasources/master-datasources.xml
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.GLOBAL_APIM, constants.CONF, constants.DATA_SOURCES, constants.MASTER_DATA_SOURCES_XML), constants.DATABASE_USERNAME, constants.GCP_SQL_USER_NAME, -1); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("%v: %v", constants.ERROR_REPLACING_MASTER_DATASOURCES_XML, err)
	}
	// Replace password in /global-apim/conf/datasources/master-datasources.xml
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.GLOBAL_APIM, constants.CONF, constants.DATA_SOURCES, constants.MASTER_DATA_SOURCES_XML), constants.DATABASE_PASSWORD, constants.GCP_SQL_PASSWORD+uniqueNumber, -1); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("%v: %v", constants.ERROR_REPLACING_MASTER_DATASOURCES_XML, err)
	}
	// Replace host in /global-apim/conf/datasources/master-datasources.xml
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.GLOBAL_APIM, constants.CONF, constants.DATA_SOURCES, constants.MASTER_DATA_SOURCES_XML), constants.MYSQL_DATABASE_HOST, sqlIpAddress, -1); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("%v: %v", constants.ERROR_REPLACING_MASTER_DATASOURCES_XML, err)
	}
	// Replace username in /observability/sp/conf/deployment.yaml
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.OBSERVABILITY, constants.SP, constants.CONF, constants.DEPLOYMENT_YAML), constants.DATABASE_USERNAME, constants.GCP_SQL_USER_NAME, -1); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("%V: %v", constants.ERROR_REPLACING_OBSERVABILITY_YAML, err)
	}
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.OBSERVABILITY, constants.SP, constants.CONF, constants.DEPLOYMENT_YAML), constants.DATABASE_PASSWORD, constants.GCP_SQL_PASSWORD+uniqueNumber, -1); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("%V: %v", constants.ERROR_REPLACING_OBSERVABILITY_YAML, err)
	}
	if err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.OBSERVABILITY, constants.SP, constants.CONF, constants.DEPLOYMENT_YAML), constants.MYSQL_DATABASE_HOST, sqlIpAddress, -1); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("%V: %v", constants.ERROR_REPLACING_OBSERVABILITY_YAML, err)
	}
	return sqlService, serviceAccountEmailAddress
}

func createKubernentesClusterOnGcp(ctx context.Context, gcpSpinner *util.Spinner) {
	gkeClient, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Could not get authenticated client: %v", err)
	}
	gcpService, err := container.New(gkeClient)
	if err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Could not initialize gke client: %v", err)
	}
	// Create GCP cluster
	gcpSpinner.SetNewAction("Creating GCP cluster")
	errCreate := createGcpCluster(gcpService, constants.GCP_CLUSTER_NAME+uniqueNumber)
	if errCreate != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Could not create cluster: %v", errCreate)
	}
	// Update kube config file
	gcpSpinner.SetNewAction("Updating kube config cluster")
	updateKubeConfig()
}

func configureGCPCredentials() string {
	// Backup artifacts folder
	artiFactsBackupExist, errBackupDir := util.FileExists(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS_OLD))
	if errBackupDir == nil {
		if artiFactsBackupExist {
			if err := os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS)); err != nil {
				fmt.Printf("Error replacing artifacts filel: %v", err)
			}
			if err := os.Rename(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS_OLD), filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS)); err != nil {
				fmt.Printf("Error replacing artifacts filel: %v", err)
			}
		}
	}
	util.CopyDir(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS), filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS_OLD))
	validateGcpConfigFile([]string{filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.GLOBAL_APIM, constants.CONF, constants.DATA_SOURCES, constants.MASTER_DATA_SOURCES_XML),
		filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.OBSERVABILITY, constants.SP, constants.CONF, constants.DEPLOYMENT_YAML),
		filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.GLOBAL_APIM, constants.ARTIFACTS_PERSISTENT_VOLUME_YAML),
		filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.MYSQL, constants.DB_SCRIPTS, constants.INIT_SQL)})
	// Get the GCP cluster data
	projectName, accountName, region, zone = getGcpData()
	var gcpBucketName = constants.GCP_BUCKET_NAME + uniqueNumber
	jsonAuthFile := util.FindInDirectory(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP), ".json")
	if len(jsonAuthFile) > 0 {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", jsonAuthFile[0])
	} else {
		fmt.Printf("Could not find authentication json file in : %v", filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP))
		os.Exit(1)
	}
	return gcpBucketName
}

func createGcpCluster(gcpService *container.Service, clusterName string) error {
	gcpGKENode := &container.NodeConfig{
		ImageType:   constants.GCP_CLUSTER_IMAGE_TYPE,
		MachineType: constants.GCP_CLUSTER_MACHINE_TYPE,
		DiskSizeGb:  constants.GCP_CLUSTER_DISK_SIZE_GB,
	}
	gcpCluster := &container.Cluster{
		Name:                  clusterName,
		Description:           "k8s cluster",
		InitialClusterVersion: "latest",
		EnableKubernetesAlpha: true,
		InitialNodeCount:      1,
		Location:              zone,
		NodeConfig:            gcpGKENode,
	}
	createClusterRequest := &container.CreateClusterRequest{
		ProjectId: projectName,
		Zone:      zone,
		Cluster:   gcpCluster,
	}

	k8sCluster, err := gcpService.Projects.Zones.Clusters.Create(projectName, zone, createClusterRequest).Do()

	// Loop until cluster status becomes RUNNING
	for i := 0; i < 30; i++ {
		resp, err := gcpService.Projects.Zones.Clusters.Get(projectName, zone, constants.GCP_CLUSTER_NAME+uniqueNumber).Do()
		if err != nil {
			time.Sleep(60 * time.Second)
		} else {
			if resp.Status == "RUNNING" {
				fmt.Printf("Cluster: %v created", k8sCluster.Name)
				return nil
			}
			time.Sleep(15 * time.Second)
		}
	}
	return fmt.Errorf("failed to create clusters: %v", err)
}

func createSqlInstance(ctx context.Context, service *sqladmin.Service, projectId string, region string, zone string, instanceName string) ([]*sqladmin.DatabaseInstance, error) {
	settings := &sqladmin.Settings{
		DataDiskSizeGb: constants.GCP_SQL_DISK_SIZE_GB,
		Tier:           constants.GCP_SQL_TIER,
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
		fmt.Printf("Error in sql instance creation")
	}
	return nil, nil
}

func createGcpStorage(client *storage.Client, projectID, bucketName string) error {
	ctx := context.Background()
	if err := client.Bucket(bucketName).Create(ctx, projectID, nil); err != nil {
		return err
	}
	return nil
}

func uploadSqlFile(client *storage.Client, bucket, object string) error {
	ctx := context.Background()
	err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.MYSQL, constants.DB_SCRIPTS, constants.INIT_SQL), constants.DATABASE_USERNAME, constants.GCP_SQL_USER_NAME, -1)
	err = util.ReplaceInFile(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.MYSQL, constants.DB_SCRIPTS, constants.INIT_SQL), constants.DATABASE_PASSWORD, constants.GCP_SQL_PASSWORD+uniqueNumber, -1)
	f, err := os.Open(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS, constants.K8S_ARTIFACTS, constants.MYSQL, constants.DB_SCRIPTS, constants.INIT_SQL))
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

func updateKubeConfig() {
	cmd := exec.Command("bash", "-c", "gcloud container clusters get-credentials "+constants.GCP_CLUSTER_NAME+uniqueNumber+" --zone "+zone+" --project "+projectName)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}
	defer cmd.Wait()

	go io.Copy(os.Stdout, stderr)
}

func createNfsServer(nfsService *file.Service) error {
	fileShare := &file.FileShareConfig{
		Name:       constants.GCP_NFS_CONFIG_NAME,
		CapacityGb: constants.GCP_NFS_CONFIG_CAPACITY,
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
	}

	if _, err := nfsService.Projects.Locations.Instances.Create("projects/"+projectName+"/locations/"+zone, nfsInstance).InstanceId(constants.GCP_NFS_SERVER_INSTANCE + uniqueNumber).Do(); err != nil {
		return err
	}
	return nil
}

func getNfsServerIp(nfsService *file.Service) (string, error) {
	serverIp := ""
	for true {
		inst, err := nfsService.Projects.Locations.Instances.Get("projects/" + projectName + "/locations/" + zone + "/instances/" + constants.GCP_NFS_SERVER_INSTANCE + uniqueNumber).Do()
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

func getSqlServieAccount(ctx context.Context, gcpService *sqladmin.Service, projectId string, instanceName string) (string, string) {
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

func updateBucketPermission(bucketName string, svcAccount string) (err error) {
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

func importSqlScript(service *sqladmin.Service, projectId string, instanceName string, uri string) error {
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

func updateInstance(service *sqladmin.Service, projectId string, instanceName string) error {
	aclEntry := &sqladmin.AclEntry{
		Value: "0.0.0.0/0",
	}

	aclEntryList := []*sqladmin.AclEntry{}
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

	_, err := service.Instances.Patch(projectId, instanceName, newDbInstance).Do()
	if err != nil {
		fmt.Printf("Error updating instance :%v", err)
	}
	return nil
}

func validateGcpConfigFile(configFiles []string) error {
	for i := 0; i < len(configFiles); i++ {
		fileExist, fileExistError := util.FileExists(configFiles[i])
		if fileExistError != nil || !fileExist {
			fmt.Printf("Cannot find file : %v", configFiles[i])
			os.Exit(1)
		}
	}
	return nil
}

func deployMinimalCelleryRuntime() error {
	util.CopyDir(util.CelleryInstallationDir(), filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS))
	var artifactPath = filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS)
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"
	// Give permission
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, "clusterrolebinding", "cluster-admin-binding", "--clusterrole", "cluster-admin", "--user", accountName), errorDeployingCelleryRuntime)

	// Setup Celley namespace, create service account and the docker registry credentials
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/ns-init.yaml"), errorDeployingCelleryRuntime)

	// Istio
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/istio-crds.yaml"), errorDeployingCelleryRuntime)

	// Enabling Istio injection
	util.ExecuteCommand(exec.Command(constants.KUBECTL, "label", "namespace", "default", "istio-injection=enabled"), errorDeployingCelleryRuntime)

	// Without security
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/istio-demo-cellery.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/istio-gateway.yaml"), errorDeployingCelleryRuntime)

	// Install Cellery crds
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/01-cluster-role.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/02-service-account.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/03-cluster-role-binding.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/04-crd-cell.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/05-crd-gateway.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/06-crd-token-service.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/07-crd-service.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/08-config.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/controller/09-controller.yaml"), errorDeployingCelleryRuntime)

	return nil
}

func deployCompleteCelleryRuntime() {
	var artifactPath = filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP, constants.ARTIFACTS)
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"

	deployMinimalCelleryRuntime()

	// Create apim NFS volumes and volume claims
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-apim/artifacts-persistent-volume.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-apim/artifacts-persistent-volume-claim.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Create the gw config maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "gw-conf", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "gw-conf-datasources", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/datasources/", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Create KM config maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "conf-identity", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/identity", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "apim-template", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/resources/api_templates", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "apim-tomcat", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/tomcat", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "apim-security", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/security", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	//Create gateway deployment and the service
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/global-apim/global-apim.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Wait till the gateway deployment availability
	//util.ExecuteCommand(exec.Command(constants.KUBECTL, "wait", "deployment.apps/gateway", "--for", "condition available", "--timeout", "6000s", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Create SP worker configmaps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "sp-worker-siddhi", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/identity", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "sp-worker-conf", "--from-file", artifactPath+"/k8s-artefacts/observability/sp/conf", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Create SP worker deployment
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/observability/sp/sp-worker.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Create observability portal deployment, service and ingress.
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "observability-portal-config", "--from-file", artifactPath+"/k8s-artefacts/observability/node-server/config", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/observability/portal/observability-portal.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Create K8s Metrics Config-maps
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "k8s-metrics-prometheus-conf", "--from-file", artifactPath+"/k8s-artefacts/observability/prometheus/config", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "k8s-metrics-grafana-conf", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/config", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "k8s-metrics-grafana-datasources", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/datasources", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "k8s-metrics-grafana-dashboards", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/dashboards", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.CREATE, constants.CONFIG_MAP, "k8s-metrics-grafana-dashboards-default", "--from-file", artifactPath+"/k8s-artefacts/observability/grafana/dashboards/default", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Create K8s Metrics deployment, service and ingress.
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/observability/prometheus/k8s-metrics-prometheus.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/observability/grafana/k8s-metrics-grafana.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	// Install nginx-ingress for control plane ingress
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/mandatory.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command(constants.KUBECTL, constants.APPLY, constants.KUBECTL_FLAG, artifactPath+"/k8s-artefacts/system/cloud-generic.yaml"), errorDeployingCelleryRuntime)
}

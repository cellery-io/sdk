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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/manifoldco/promptui"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/file/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
	"github.com/cellery-io/sdk/components/cli/pkg/runtime"
	"github.com/cellery-io/sdk/components/cli/pkg/runtime/gcp"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

var projectName string
var accountName string
var region string
var zone string
var uniqueNumber string

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	uniqueNumber = strconv.Itoa(rand.Intn(1000))
}

func RunSetupCreateGcp(isCompleteSetup bool) {
	projectName, accountName, region, zone = getGcpData()
	if region == "" {
		util.ExitWithErrorMessage("Error creating cluster", fmt.Errorf("region not found in gcloud "+
			"config list. Please run `gcloud init` to set the region"))
	}
	if zone == "" {
		util.ExitWithErrorMessage("Error creating cluster", fmt.Errorf("zone not found in gcloud "+
			"config list. Please run `gcloud init` to set the zone"))
	}
	if isCompleteSetup {
		createCompleteGcpRuntime()
	} else {
		createMinimalGcpRuntime()
	}
	runtime.WaitFor(true, false)
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
		Items:     []string{constants.BASIC, constants.COMPLETE, constants.CELLERY_SETUP_BACK},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select an option: %v", err)
	}
	if value == constants.CELLERY_SETUP_BACK {
		createEnvironment()
		return nil
	}
	if value == constants.COMPLETE {
		isCompleteSelected = true
	}
	RunSetupCreateGcp(isCompleteSelected)
	return nil
}

func createMinimalGcpRuntime() {
	util.CopyK8sArtifacts(util.UserHomeCelleryDir())
	gcpBucketName := configureGCPCredentials()
	_ = configureGCPCredentials()
	ctx := context.Background()

	// Create a GKE client
	gcpSpinner := util.StartNewSpinner("Creating GKE client")
	createKubernetesClusterOnGcp(ctx, gcpSpinner)

	sqlService, serviceAccountEmailAddress := configureMysqlOnGcp(ctx, gcpSpinner)

	configureBucketOnGcp(ctx, gcpSpinner, gcpBucketName, serviceAccountEmailAddress, sqlService)

	configureNfsOnGcp(gcpSpinner, ctx)

	// Deploy cellery runtime
	gcpSpinner.SetNewAction("Deploying Cellery runtime")
	deployMinimalCelleryRuntime()
	util.RemoveDir(filepath.Join(util.UserHomeCelleryDir(), constants.K8S_ARTIFACTS))
	gcpSpinner.Stop(true)
}

func createCompleteGcpRuntime() error {
	util.CopyK8sArtifacts(util.UserHomeCelleryDir())
	gcpBucketName := configureGCPCredentials()
	ctx := context.Background()

	// Create a GKE client
	gcpSpinner := util.StartNewSpinner("Creating GKE client")
	createKubernetesClusterOnGcp(ctx, gcpSpinner)

	sqlService, serviceAccountEmailAddress := configureMysqlOnGcp(ctx, gcpSpinner)

	configureBucketOnGcp(ctx, gcpSpinner, gcpBucketName, serviceAccountEmailAddress, sqlService)

	configureNfsOnGcp(gcpSpinner, ctx)

	// Deploy cellery runtime
	gcpSpinner.SetNewAction("Deploying Cellery runtime")
	deployCompleteCelleryRuntime()
	util.RemoveDir(filepath.Join(util.UserHomeCelleryDir(), constants.K8S_ARTIFACTS))
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
	if err := gcp.UpdateNfsServerDetails(nfsIpAddress, "/data"); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error replacing in file artifacts-persistent-volume.yaml: %v", err)
	}
}

func configureBucketOnGcp(ctx context.Context, gcpSpinner *util.Spinner, gcpBucketName string,
	serviceAccountEmailAddress string, sqlService *sqladmin.Service) {
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
	gcpSpinner.SetNewAction("Uploading init.sql file to GCP bucket")
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
	_, err = createSqlInstance(sqlService, projectName, region, region, constants.GCP_DB_INSTANCE_NAME+uniqueNumber)
	if err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error creating sql instance: %v", err)
	}
	sqlIpAddress, serviceAccountEmailAddress := getSqlServieAccount(ctx, sqlService, projectName,
		constants.GCP_DB_INSTANCE_NAME+uniqueNumber)
	fmt.Printf("Sql Ip address : %v", sqlIpAddress)

	if err := gcp.UpdateMysqlCredentials(constants.GCP_SQL_USER_NAME, constants.GCP_SQL_PASSWORD+uniqueNumber,
		sqlIpAddress); err != nil {
		gcpSpinner.Stop(false)
		fmt.Printf("Error updating file: %v", err)
	}

	return sqlService, serviceAccountEmailAddress
}

func createKubernetesClusterOnGcp(ctx context.Context, gcpSpinner *util.Spinner) {
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
	gcpSpinner.SetNewAction("Creating GCP cluster " + constants.GCP_CLUSTER_NAME + uniqueNumber)
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
	// Get the GCP cluster data
	projectName, accountName, region, zone = getGcpData()
	var gcpBucketName = constants.GCP_BUCKET_NAME + uniqueNumber
	jsonAuthFile := util.FindInDirectory(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.GCP),
		".json")
	if len(jsonAuthFile) > 0 {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", jsonAuthFile[0])
	} else {
		util.ExitWithErrorMessage("Failed to create gcp cluster", fmt.Errorf(
			"Could not find authentication json file in : %s. Please copy GCP service account credentials"+
				" json file into this directory.\n", filepath.Join(util.UserHomeDir(),
				constants.CELLERY_HOME, constants.GCP)))
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
		InitialClusterVersion: "1.12",
		EnableKubernetesAlpha: true,
		InitialNodeCount:      1,
		Location:              zone,
		NodeConfig:            gcpGKENode,
		ResourceLabels:        getCreatedByLabel(),
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
			time.Sleep(30 * time.Second)
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

func createSqlInstance(service *sqladmin.Service, projectId string, region string, zone string,
	instanceName string) ([]*sqladmin.DatabaseInstance, error) {
	settings := &sqladmin.Settings{
		DataDiskSizeGb: constants.GCP_SQL_DISK_SIZE_GB,
		Tier:           constants.GCP_SQL_TIER,
		UserLabels:     getCreatedByLabel(),
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
	attributes := &storage.BucketAttrs{
		Labels: getCreatedByLabel(),
	}
	if err := client.Bucket(bucketName).Create(ctx, projectID, attributes); err != nil {
		return err
	}
	return nil
}

func uploadSqlFile(client *storage.Client, bucket, object string) error {
	ctx := context.Background()
	if err := gcp.UpdateInitSql(constants.GCP_SQL_USER_NAME, constants.GCP_SQL_PASSWORD+uniqueNumber); err != nil {
		return err
	}
	f, err := os.Open(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.K8S_ARTIFACTS,
		constants.MYSQL, constants.DB_SCRIPTS, constants.INIT_SQL))
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
		Labels:     getCreatedByLabel(),
	}

	if _, err := nfsService.Projects.Locations.Instances.Create("projects/"+projectName+"/locations/"+zone,
		nfsInstance).InstanceId(constants.GCP_NFS_SERVER_INSTANCE + uniqueNumber).Do(); err != nil {
		return err
	}
	return nil
}

func getNfsServerIp(nfsService *file.Service) (string, error) {
	serverIp := ""
	for true {
		inst, err := nfsService.Projects.Locations.Instances.Get("projects/" + projectName + "/locations/" + zone +
			"/instances/" + constants.GCP_NFS_SERVER_INSTANCE + uniqueNumber).Do()
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

	_, err := service.Instances.Patch(projectId, instanceName, newDbInstance).Do()
	if err != nil {
		fmt.Printf("Error updating instance :%v", err)
	}
	return nil
}

func createController(errorMessage string) {
	// Give permission to the user
	if err := kubectl.CreateClusterRoleBinding("cluster-admin", accountName); err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}

	// Setup Celley namespace
	if err := runtime.CreateCelleryNameSpace(); err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}

	// Apply Istio CRDs
	if err := runtime.ApplyIstioCrds(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS)); err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}
	// sleep for few seconds - this is to make sure that the CRDs are properly applied
	time.Sleep(20 * time.Second)

	// Enabling Istio injection
	if err := kubectl.ApplyLable("namespace", "default", "istio-injection=enabled",
		false); err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}

	// Install istio
	if err := runtime.InstallIstio(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS)); err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}

	// Install knative serving
	if err := runtime.InstallKnativeServing(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS)); err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}

	// Apply controller CRDs
	if err := runtime.InstallController(filepath.Join(util.CelleryInstallationDir(), constants.K8S_ARTIFACTS)); err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}
}

func deployMinimalCelleryRuntime() error {
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"

	createController(errorDeployingCelleryRuntime)
	createAllDeploymentArtifacts()
	createIdpGcp(errorDeployingCelleryRuntime)
	createNGinx(errorDeployingCelleryRuntime)

	return nil
}

func deployCompleteCelleryRuntime() {
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"

	createController(errorDeployingCelleryRuntime)
	createAllDeploymentArtifacts()

	//Create gateway deployment and the service
	if err := gcp.AddApim(); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}

	// Create observability
	if err := gcp.AddObservability(); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}
	//Create NGinx
	createNGinx(errorDeployingCelleryRuntime)
}

func createIdpGcp(errorDeployingCelleryRuntime string) {
	// Create IDP deployment and the service
	if err := gcp.CreateIdp(); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}
}

func createNGinx(errorMessage string) {
	// Install nginx-ingress for control plane ingress
	if err := gcp.InstallNginx(); err != nil {
		util.ExitWithErrorMessage(errorMessage, err)
	}
}

func createAllDeploymentArtifacts() {
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"

	// Create apim NFS volumes and volume claims
	if err := gcp.CreatePersistentVolume(); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}
	// Create the gw config maps
	if err := gcp.CreateGlobalGatewayConfigMaps(); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}
	// Create Observability configmaps
	if err := gcp.CreateObservabilityConfigMaps(); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}
	// Create the IDP config maps
	if err := gcp.CreateIdpConfigMaps(); err != nil {
		util.ExitWithErrorMessage(errorDeployingCelleryRuntime, err)
	}
}

func getGcpData() (string, string, string, string) {
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
		util.ExitWithErrorMessage("Error occurred while getting gcp data", err)
	}
	err = cmd.Wait()
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while getting gcp data", err)
	}

	jsonOutput := &util.Gcp{}

	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson != nil {
		fmt.Println(errJson)
	}
	return jsonOutput.Core.Project, jsonOutput.Core.Account, jsonOutput.Compute.Region, jsonOutput.Compute.Zone
}

func getCreatedByLabel() map[string]string {
	// Add a label to identify the user who created it
	labels := make(map[string]string)
	createdBy := util.ConvertToAlphanumeric(accountName, "_")
	labels["created_by"] = createdBy
	return labels
}

/*
 * Copyright (c) 2018 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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

	"golang.org/x/oauth2/google"

	"cloud.google.com/go/storage"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"google.golang.org/api/sqladmin/v1beta4"

	"context"

	"google.golang.org/api/container/v1"
	"google.golang.org/api/file/v1"
)

func RunSetup() {
	selectTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Setup Cellery runtime",
		Items:     []string{constants.CELLERY_SETUP_MANAGE, constants.CELLERY_SETUP_CREATE, constants.CELLERY_SETUP_SWITCH, constants.CELLERY_SETUP_EXIT},
		Templates: selectTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option: %v", err)
	}

	switch value {
	case constants.CELLERY_SETUP_MANAGE:
		{
			manageEnvironment()
		}
	case constants.CELLERY_SETUP_CREATE:
		{
			createEnvironment()
		}
	case constants.CELLERY_SETUP_SWITCH:
		{
			selectEnvironment()
		}
	default:
		{
			os.Exit(1)
		}
	}
}

func selectEnvironment() error {
	bold := color.New(color.Bold).SprintFunc()
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Selected: bold("Selected cluster: ") + "{{ . }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select a Cellery Installed Kubernetes Cluster",
		Items:     getContexts(),
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select cluster: %v", err)
	}

	setContext(value)
	fmt.Printf(util.GreenBold("\n\U00002714") + " Successfully configured Cellery.\n")
	fmt.Println()
	fmt.Println(bold("What's next ?"))
	fmt.Println("======================")
	fmt.Println("To create your first project, execute the command: ")
	fmt.Println("  $ cellery init ")
	return nil
}

func getContexts() []string {
	contexts := []string{}
	cmd := exec.Command("kubectl", "config", "view", "-o", "json")
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

	execError := ""
	go func() {
		for stderrScanner.Scan() {
			execError += stderrScanner.Text()
		}
	}()
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error in executing cellery setup: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while configuring cellery: \x1b[0m %v \n", execError)
		os.Exit(1)
	}
	jsonOutput := &Config{}
	errJson := json.Unmarshal([]byte(output), jsonOutput)
	if errJson != nil {
		fmt.Println(errJson)
	}

	for i := 0; i < len(jsonOutput.Contexts); i++ {
		contexts = append(contexts, jsonOutput.Contexts[i].Name)
	}
	return contexts
}

func setContext(context string) error {
	cmd := exec.Command("kubectl", "config", "use-context", context)
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)

	execError := ""
	go func() {
		for stderrScanner.Scan() {
			execError += stderrScanner.Text()
		}
	}()
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error in executing cellery setup: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while configuring cellery: \x1b[0m %v \n", execError)
		os.Exit(1)
	}
	return nil
}

func createEnvironment() error {
	bold := color.New(color.Bold).SprintFunc()
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select an environment to be installed",
		Items:     getCreateEnvironmentList(),
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("Failed to install environment: %v", err)
	}

	switch value {
	case constants.CELLERY_CREATE_LOCAL:
		{
			vmLocation := filepath.Join(util.UserHomeDir(), ".cellery", "vm")
			repoCreateErr := util.CreateDir(vmLocation)
			if repoCreateErr != nil {
				fmt.Println("Error while creating VM location: " + repoCreateErr.Error())
				os.Exit(1)
			}
			spinner := util.StartNewSpinner("Downloading Cellery Runtime")
			defer func() {
				spinner.Stop(true)
			}()
			util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_VM, vmLocation)
			util.ExtractTarGzFile(vmLocation, filepath.Join(util.UserHomeDir(), ".cellery", "vm", constants.AWS_S3_ITEM_VM))
			util.DownloadFromS3Bucket(constants.AWS_S3_BUCKET, constants.AWS_S3_ITEM_CONFIG, vmLocation)
			util.ReplaceFile(filepath.Join(util.UserHomeDir(), ".kube", "config"), filepath.Join(util.UserHomeDir(), ".cellery", "vm", constants.AWS_S3_ITEM_CONFIG))
			installVM()
		}
	case constants.CELLERY_CREATE_GCP:
		{
			createGcp()
		}
	default:
		{
			RunSetup()
		}
	}

	fmt.Printf(util.GreenBold("\n\U00002714") + " Successfully installed Cellery runtime.\n")
	fmt.Println()
	fmt.Println(bold("What's next ?"))
	fmt.Println("======================")
	fmt.Println("To create your first project, execute the command: ")
	fmt.Println("  $ cellery init ")
	return nil
}

func installVM() error {
	spinner := util.StartNewSpinner("Installing Cellery Runtime")
	defer func() {
		spinner.Stop(true)
	}()
	util.ExecuteCommand(exec.Command("VBoxManage", "createvm", "--name", constants.VM_NAME, "--ostype", "Ubuntu_64", "--register"), "Error Installing VM")
	util.ExecuteCommand(exec.Command("VBoxManage", "modifyvm", constants.VM_NAME, "--ostype", "Ubuntu_64", "--cpus", "4", "--memory", "8000", "--natpf1", "guestkube,tcp,,6443,,6443", "--natpf1", "guestssh,tcp,,2222,,22", "--natpf1", "guesthttps,tcp,,443,,443", "--natpf1", "guesthttp,tcp,,80,,80"), "Error Installing VM")
	util.ExecuteCommand(exec.Command("VBoxManage", "storagectl", constants.VM_NAME, "--name", "hd1", "--add", "sata", "--portcount", "2"), "Error Installing VM")

	vmLocation := filepath.Join(util.UserHomeDir(), ".cellery", "vm", constants.VM_FILE_NAME)
	util.ExecuteCommand(exec.Command("VBoxManage", "storageattach", constants.VM_NAME, "--storagectl", "hd1", "--port", "1", "--type", "hdd", "--medium", vmLocation), "Error Installing VM")
	util.ExecuteCommand(exec.Command("VBoxManage", "startvm", constants.VM_NAME, "--type", "headless"), "Error Installing VM")

	return nil
}

func manageEnvironment() error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " " + getManageLabel(),
		Items:     getManageEnvOptions(),
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("Failed to install environment: %v", err)
	}

	switch value {
	case constants.CELLERY_MANAGE_STOP:
		{
			spinner := util.StartNewSpinner("Stopping Cellery Runtime")
			defer func() {
				spinner.Stop(true)
			}()
			util.ExecuteCommand(exec.Command("VBoxManage", "controlvm", constants.VM_NAME, "acpipowerbutton"), "Error stopping VM")
		}
	case constants.CELLERY_MANAGE_START:
		{
			util.ExecuteCommand(exec.Command("VBoxManage", "startvm", constants.VM_NAME, "--type", "headless"), "Error starting VM")
		}
	case constants.CELLERY_MANAGE_CLEANUP:
		{
			spinner := util.StartNewSpinner("Removing Cellery Runtime")
			defer func() {
				spinner.Stop(true)
			}()
			if isVmRuning() {
				util.ExecuteCommand(exec.Command("VBoxManage", "controlvm", constants.VM_NAME, "acpipowerbutton"), "Error stopping VM")
			}
			for isVmRuning() {
				time.Sleep(2 * time.Second)
			}
			util.ExecuteCommand(exec.Command("VBoxManage", "unregistervm", constants.VM_NAME, "--delete"), "Error deleting VM")
		}
	default:
		{
			RunSetup()
		}
	}
	return nil
}

func getManageLabel() string {
	var manageLabel string
	if isVmInstalled() {
		if isVmRuning() {
			manageLabel = constants.VM_NAME + " is running. Select `Stop` to stop the VM"
		} else {
			manageLabel = constants.VM_NAME + " is installed. Select `Start` to start the VM"
		}
	} else {
		manageLabel = "Cellery runtime is not installed"
	}
	return manageLabel
}

func isVmRuning() bool {
	if isVmInstalled() {
		cmd := exec.Command("vboxmanage", "showvminfo", constants.VM_NAME)
		stdoutReader, _ := cmd.StdoutPipe()
		stdoutScanner := bufio.NewScanner(stdoutReader)
		output := ""
		go func() {
			for stdoutScanner.Scan() {
				output = output + stdoutScanner.Text()
			}
		}()
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error occurred while checking VM status: %v \n", err)
			os.Exit(1)
		}
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("\x1b[31;1m Error occurred while checking VM status \x1b[0m %v \n", err)
			os.Exit(1)
		}
		if strings.Contains(output, "running (since") {
			return true
		}
	}
	return false
}

func getManageEnvOptions() []string {
	if isVmInstalled() {
		if isVmRuning() {
			return []string{constants.CELLERY_MANAGE_STOP, constants.CELLERY_MANAGE_CLEANUP, constants.CELLERY_SETUP_BACK}
		} else {
			return []string{constants.CELLERY_MANAGE_START, constants.CELLERY_MANAGE_CLEANUP, constants.CELLERY_SETUP_BACK}
		}
	}
	return []string{constants.CELLERY_SETUP_BACK}
}

func isVmInstalled() bool {
	cmd := exec.Command("vboxmanage", "list", "vms")
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	output := ""
	go func() {
		for stdoutScanner.Scan() {
			output = output + stdoutScanner.Text()
		}
	}()
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error occurred while checking if VMs installed: %v \n", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while checking if VMs installed \x1b[0m %v \n", err)
		os.Exit(1)
	}

	if strings.Contains(output, constants.VM_NAME) {
		return true
	}
	return false
}

func getCreateEnvironmentList() []string {
	if isVmInstalled() {
		return []string{constants.CELLERY_CREATE_GCP, constants.CELLERY_SETUP_BACK}
	}
	return []string{constants.CELLERY_CREATE_LOCAL, constants.CELLERY_CREATE_GCP, constants.CELLERY_SETUP_BACK}
}

func createGcp() error {
	var gcpBucketName = constants.GCP_BUCKET_NAME + "-" + strconv.Itoa(rand.Intn(1000))
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filepath.Join(util.UserHomeDir(), ".cellery", "gcp", "vick-team-1abab5311d43.json"))
	ctx := context.Background()

	// Create a GKE client
	gkeClient, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		fmt.Printf("Could not get authenticated client: %v", err)
	}

	gcpService, err := container.New(gkeClient)
	if err != nil {
		fmt.Printf("Could not initialize gke client: %v", err)
	}

	// Create GCP cluster
	errCreate := createGcpCluster(gcpService, constants.GCP_CLUSTER_NAME)
	if errCreate != nil {
		fmt.Printf("Could not create cluster: %v", errCreate)
	}

	// Update kube config file
	updateKubeConfig()

	hcSql, err := google.DefaultClient(ctx, sqladmin.CloudPlatformScope)

	if err != nil {
		fmt.Printf("Error creating client: %v", err)
	}

	sqlService, err := sqladmin.New(hcSql)
	if err != nil {
		fmt.Printf("Error creating sql service: %v", err)
	}

	//Create sql instance
	_, err = createSqlInstance(ctx, sqlService, constants.GCP_PROJECT_NAME, constants.GCP_REGION, constants.GCP_REGION, constants.GCP_DB_INSTANCE_NAME)
	if err != nil {
		log.Print(err)
	}

	_, serviceAccountEmailAddress := getSqlServieAccount(ctx, sqlService, constants.GCP_PROJECT_NAME, constants.GCP_DB_INSTANCE_NAME)

	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("Error creating storage client: %v", err)
	}

	// Create bucket
	if err := create(client, constants.GCP_PROJECT_NAME, gcpBucketName); err != nil {
		fmt.Printf("Error creating storage client: %v", err)
	}

	// Upload init file to S3 bucket
	if err := uploadSqlFile(client, gcpBucketName, "init.sql"); err != nil {
		fmt.Printf("Error Uploading Sql file: %v", err)
	}

	// Update bucket permission
	if err := updateBucketPermission(gcpBucketName, serviceAccountEmailAddress); err != nil {
		fmt.Printf("Error Updating bucket permission: %v", err)
	}

	// Import sql script
	var uri = "gs://" + gcpBucketName + "/init.sql"
	if err := importSqlScript(sqlService, constants.GCP_PROJECT_NAME, constants.GCP_DB_INSTANCE_NAME, uri); err != nil {
		fmt.Printf("Error Updating bucket permission: %v", err)
	}
	time.Sleep(30 * time.Second)

	// Update sql instance
	if err := updateInstance(sqlService, constants.GCP_PROJECT_NAME, constants.GCP_DB_INSTANCE_NAME); err != nil {
		fmt.Printf("Error Updating sql instance: %v", err)
	}

	// Create NFS server
	hcNfs, err := google.DefaultClient(ctx, file.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	nfsService, err := file.New(hcNfs)
	if err != nil {
		return err
	}

	if err := createNfsServer(nfsService); err != nil {
		fmt.Printf("Error creating NFS server: %v", err)
	}

	nfsIpAddress, errIp := getNfsServerIp(nfsService)
	fmt.Printf("Nfs Ip address : %v", nfsIpAddress)
	if errIp != nil {
		fmt.Printf("Error getting NFS server IP address: %v", errIp)
	}

	util.ReplaceInFile(filepath.Join(util.UserHomeDir(), ".cellery", "gcp", "artifacts", "k8s-artefacts", "global-apim", "artifacts-persistent-volume.yaml"), "NFS_SERVER_IP", nfsIpAddress, -1)
	// Deploy cellery runtime
	deployCelleryRuntime()

	return nil
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
		Location:              constants.GCP_ZONE,
		NodeConfig:            gcpGKENode,
	}
	createClusterRequest := &container.CreateClusterRequest{
		ProjectId: constants.GCP_PROJECT_NAME,
		Zone:      constants.GCP_ZONE,
		Cluster:   gcpCluster,
	}

	k8sCluster, err := gcpService.Projects.Zones.Clusters.Create(constants.GCP_PROJECT_NAME, constants.GCP_ZONE, createClusterRequest).Do()
	spinner := util.StartNewSpinner("Creating GCP cluster")
	defer func() {
		spinner.Stop(true)
	}()
	var clusterStatus string
	for i := 0; i < 50; i++ {
		clusterStatus, err = getClusterState(gcpService, constants.GCP_PROJECT_NAME, constants.GCP_ZONE, constants.GCP_CLUSTER_NAME)
		if err != nil {
			fmt.Printf("Error creating cluster : %v", err)
			break
		} else {
			if clusterStatus == "RUNNING" {
				fmt.Printf("Cluster: %v created", k8sCluster.Name)
				break
			}
			time.Sleep(10 * time.Second)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to list clusters: %v", err)
	}
	fmt.Printf("%s", k8sCluster.Name)

	return nil
}

func getClusterState(gcpService *container.Service, projectID, zone, clusterId string) (string, error) {
	status := ""
	rsp, err := gcpService.Projects.Zones.Clusters.Get(projectID, zone, clusterId).Do()
	if err != nil {
		status = rsp.Status
	}

	return status, err
}

func createSqlInstance(ctx context.Context, service *sqladmin.Service, projectId string, region string, zone string, instanceName string) ([]*sqladmin.DatabaseInstance, error) {
	settings := &sqladmin.Settings{
		DataDiskSizeGb: 12,
		Tier:           "db-f1-micro",
	}

	dbInstance := &sqladmin.DatabaseInstance{
		Name:            instanceName,
		Project:         projectId,
		Region:          region,
		GceZone:         zone,
		BackendType:     "SECOND_GEN",
		DatabaseVersion: "MYSQL_5_7",
		MaxDiskSize:     15,
		Settings:        settings,
	}

	if _, err := service.Instances.Insert(projectId, dbInstance).Do(); err != nil {
		fmt.Printf("Error in sql instance creation")
		log.Fatal(err)
	}
	return nil, nil
}

func create(client *storage.Client, projectID, bucketName string) error {
	ctx := context.Background()
	if err := client.Bucket(bucketName).Create(ctx, projectID, nil); err != nil {
		return err
	}
	return nil
}

func uploadSqlFile(client *storage.Client, bucket, object string) error {
	ctx := context.Background()
	err := util.ReplaceInFile(filepath.Join(util.UserHomeDir(), ".cellery", "gcp", "sql", "init.sql"), "DATABASE_USERNAME", constants.GCP_SQL_USER_NAME, -1)
	err = util.ReplaceInFile(filepath.Join(util.UserHomeDir(), ".cellery", "gcp", "sql", "init.sql"), "DATABASE_PASSWORD", constants.GCP_SQL_PASSWORD+strconv.Itoa(rand.Intn(1000)), -1)
	f, err := os.Open(filepath.Join(util.UserHomeDir(), ".cellery", "gcp", "sql", "init.sql"))
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
	cmd := exec.Command("bash", "-c", "gcloud container clusters get-credentials "+constants.GCP_CLUSTER_NAME+" --zone "+constants.GCP_ZONE+" --project "+constants.GCP_PROJECT_NAME)
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

	if _, err := nfsService.Projects.Locations.Instances.Create("projects/"+constants.GCP_PROJECT_NAME+"/locations/"+constants.GCP_ZONE, nfsInstance).InstanceId(constants.GCP_NFS_SERVER_INSTANCE).Do(); err != nil {
		return err
	}
	return nil
}

func getNfsServerIp(nfsService *file.Service) (string, error) {
	serverIp := ""
	for true {
		inst, err := nfsService.Projects.Locations.Instances.Get("projects/" + constants.GCP_PROJECT_NAME + "/locations/" + constants.GCP_ZONE + "/instances/" + constants.GCP_NFS_SERVER_INSTANCE).Do()
		if err != nil {
			return serverIp, err
		} else {
			if inst.State == "READY" {
				serverIp = inst.Networks[0].IpAddresses[0]
				fmt.Printf("Nfs Ip address : %v", serverIp)
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
			log.Print(err)
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
		log.Fatal(err)
	}

	ctx := context.Background()

	bucket := c.Bucket(bucketName)
	policy, err := bucket.IAM().Policy(ctx)
	if err != nil {
		return err
	}

	policy.Add(strings.Join([]string{"serviceAccount:", svcAccount}, ""), "roles/storage.objectViewer")
	if err := bucket.IAM().SetPolicy(ctx, policy); err != nil {
		fmt.Print(err)
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

	opp, err := service.Instances.Import(projectId, instanceName, iIR).Do()
	if err != nil {
		log.Print(err)
	} else {
		fmt.Print(opp)
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

func deployCelleryRuntime() error {
	var artifactPath = filepath.Join(util.UserHomeDir(), ".cellery", "gcp", "artifacts")
	errorDeployingCelleryRuntime := "Error deploying cellery runtime"

	// Give permission
	util.ExecuteCommand(exec.Command("kubectl", "create", "clusterrolebinding", "cluster-admin-binding", "--clusterrole", "cluster-admin", "--user", constants.GCP_ACCOUNT_NAME), errorDeployingCelleryRuntime)

	// Setup Celley namespace, create service account and the docker registry credentials
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/system/ns-init.yaml"), errorDeployingCelleryRuntime)

	// Create apim NFS volumes and volume claims
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/global-apim/artifacts-persistent-volume.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/global-apim/artifacts-persistent-volume-claim.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Create the gw config maps
	util.ExecuteCommand(exec.Command("kubectl", "create", "configmap", "gw-conf", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "create", "configmap", "gw-conf-datasources", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/datasources/", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Create KM config maps
	util.ExecuteCommand(exec.Command("kubectl", "create", "configmap", "conf-identity", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/identity", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "create", "configmap", "apim-template", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/resources/api_templates", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "create", "configmap", "apim-tomcat", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/tomcat", "-n", "cellery-system"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "create", "configmap", "apim-security", "--from-file", artifactPath+"/k8s-artefacts/global-apim/conf/security", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	//Create gateway deployment and the service
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/global-apim/global-apim.yaml", "-n", "cellery-system"), errorDeployingCelleryRuntime)

	// Istio
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/system/istio-crds.yaml"), errorDeployingCelleryRuntime)

	// Without security
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/system/istio-demo-cellery.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/system/istio-gateway.yaml"), errorDeployingCelleryRuntime)

	// Install Cellery crds
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/controller/01-cluster-role.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/controller/02-service-account.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/controller/03-cluster-role-binding.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/controller/04-crd-cell.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/controller/05-crd-gateway.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/controller/06-crd-token-service.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/controller/07-crd-service.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/controller/08-config.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/controller/09-controller.yaml"), errorDeployingCelleryRuntime)

	// Install nginx-ingress for control plane ingress
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/system/mandatory.yaml"), errorDeployingCelleryRuntime)
	util.ExecuteCommand(exec.Command("kubectl", "apply", "-f", artifactPath+"/k8s-artefacts/system/cloud-generic.yaml"), errorDeployingCelleryRuntime)

	return nil
}

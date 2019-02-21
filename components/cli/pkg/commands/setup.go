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

	"cloud.google.com/go/storage"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"google.golang.org/api/sqladmin/v1beta4"

	"context"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
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
	util.ExecuteCommand(exec.Command("VBoxManage", "modifyvm", constants.VM_NAME, "--ostype", "Ubuntu_64", "--cpus", "4", "--memory", "8000", "--natpf1", "guestkube,tcp,,6443,,6443", "--natpf1", "guestssh,tcp,,2222,,22"), "Error Installing VM")
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
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filepath.Join(util.UserHomeDir(), ".cellery", "gcp", "vick-team-1abab5311d43.json"))
	ctx := context.Background()

	hc, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		fmt.Printf("Could not get authenticated client: %v", err)
	}

	svc, err := container.New(hc)
	if err != nil {
		fmt.Printf("Could not initialize gke client: %v", err)
	}

	errCreate := createGcpCluster(svc)
	if errCreate != nil {
		fmt.Printf("Could not create cluster: %v", errCreate)
	}
	updateKubeConfig()

	hc_sql, err := google.DefaultClient(ctx, sqladmin.CloudPlatformScope)

	if err != nil {
		fmt.Printf("Error creating client: %v", err)
	}
	fmt.Print(hc_sql)

	sqlService, err := sqladmin.New(hc_sql)
	if err != nil {
		fmt.Printf("Error creating sql service: %v", err)
	}

	fmt.Print(sqlService)

	_, err = createSqlInstance(ctx, sqlService, constants.GCP_PROJECT_NAME, constants.GCP_REGION, constants.GCP_REGION, constants.GCP_DB_INSTANCE_NAME)
	if err != nil {
		log.Print(err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("Error creating storage client: %v", err)
	}

	if err := create(client, constants.GCP_PROJECT_NAME, constants.GCP_BUCKET_NAME); err != nil {
		fmt.Printf("Error creating storage client: %v", err)
	}

	if err := uploadSqlFile(client, constants.GCP_BUCKET_NAME, "init.sql"); err != nil {
		fmt.Printf("Error Uploading Sql file: %v", err)
	}

	return nil
}
func createGcpCluster(svc *container.Service) error {
	gcpGKENode := &container.NodeConfig{
		ImageType:   "cos_containerd",
		MachineType: "n1-standard-1",
		DiskSizeGb:  12,
	}
	gcpCluster := &container.Cluster{
		Name:                  constants.GCP_CLUSTER_NAME,
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

	k8sCluster, err := svc.Projects.Zones.Clusters.Create(constants.GCP_PROJECT_NAME, constants.GCP_ZONE, createClusterRequest).Do()
	spinner := util.StartNewSpinner("Creating GCP cluster")
	defer func() {
		spinner.Stop(true)
	}()
	var clusterStatus string
	for i := 0; i < 50; i++ {
		clusterStatus, err = getClusterState(svc, constants.GCP_PROJECT_NAME, constants.GCP_ZONE, constants.GCP_CLUSTER_NAME)
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

func getClusterState(svc *container.Service, projectID, zone, clusterId string) (string, error) {
	status := ""
	rsp, err := svc.Projects.Zones.Clusters.Get(projectID, zone, clusterId).Do()
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

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
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2/google"

	"cloud.google.com/go/storage"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"google.golang.org/api/sqladmin/v1beta4"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"

	"context"

	"google.golang.org/api/container/v1"
	"google.golang.org/api/file/v1"
)

type Config struct {
	Contexts []Context `json:"contexts"`
}

type Context struct {
	Name string `json:"name"`
}

var uniqueNumber string
var projectName string
var accountName string
var region string
var zone string

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	uniqueNumber = strconv.Itoa(rand.Intn(1000))
}

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
	cmd := exec.Command(constants.KUBECTL, "config", "view", "-o", "json")
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
	cmd := exec.Command(constants.KUBECTL, "config", "use-context", context)
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

func manageEnvironment() error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select a runtime",
		Items:     []string{constants.CELLERY_CREATE_LOCAL, constants.CELLERY_CREATE_GCP, constants.CELLERY_SETUP_BACK},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("Failed to install environment: %v", err)
	}

	switch value {
	case constants.CELLERY_CREATE_LOCAL:
		{
			manageLocal()
		}
	case constants.CELLERY_CREATE_GCP:
		{
			manageGcp()
		}
	default:
		{
			RunSetup()
		}
	}
	return nil
}

func manageLocal() error {
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
			util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "controlvm", constants.VM_NAME, "acpipowerbutton"), "Error stopping VM")
		}
	case constants.CELLERY_MANAGE_START:
		{
			util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "startvm", constants.VM_NAME, "--type", "headless"), "Error starting VM")
		}
	case constants.CELLERY_MANAGE_CLEANUP:
		{
			cleanupLocal()
		}
	default:
		{
			manageEnvironment()
		}
	}
	return nil
}

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
		return fmt.Errorf("Failed to install environment: %v", err)
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
		cmd := exec.Command(constants.VBOX_MANAGE, "showvminfo", constants.VM_NAME)
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
	cmd := exec.Command(constants.VBOX_MANAGE, "list", "vms")
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

func cleanupLocal() error {
	spinner := util.StartNewSpinner("Removing Cellery Runtime")
	defer func() {
		spinner.Stop(true)
	}()
	if isVmRuning() {
		util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "controlvm", constants.VM_NAME, "acpipowerbutton"), "Error stopping VM")
	}
	for isVmRuning() {
		time.Sleep(2 * time.Second)
	}
	util.ExecuteCommand(exec.Command(constants.VBOX_MANAGE, "unregistervm", constants.VM_NAME, "--delete"), "Error deleting VM")
	os.RemoveAll(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, constants.VM, constants.AWS_S3_ITEM_VM))
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

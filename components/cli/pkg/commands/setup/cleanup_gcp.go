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
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/gcp"
	"cellery.io/cellery/components/cli/pkg/util"
)

func manageGcp(cli cli.Cli) error {
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select `cleanup` to remove an existing GCP cluster",
		Items:     []string{cleanup, setupBack},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select an option: %v", err)
	}

	switch value {
	case cleanup:
		{
			return cleanupGcp(cli)
		}
	default:
		{
			return manageEnvironment(cli)
		}
	}
}

func cleanupGcp(cli cli.Cli) error {
	gcpPlatform, err := gcp.NewGcp()
	if err != nil {
		return fmt.Errorf("failed to initialize celleryGcp")
	}
	clusters, err := gcpPlatform.GetClusterList()
	if err != nil {
		return fmt.Errorf("failed to get cluster list")
	}
	userCreatedGcpClusters, err := getGcpClustersCreatedByUser(cli, clusters)
	if err != nil {
		return fmt.Errorf("failed to get clusters created by user, %v", err)
	}

	selectTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select a GCP cluster to delete",
		Items:     append(userCreatedGcpClusters, setupBack),
		Templates: selectTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		util.ExitWithErrorMessage("Failed to select an option: ", err)
	}
	if value == setupBack {
		return manageGcp(cli)
	}
	return RunCleanupGcp(value)
}

func ValidateGcpCluster(cluster string) (bool, error) {
	valid := false
	gcpPlatform, err := gcp.NewGcp()
	if err != nil {
		return false, fmt.Errorf("failed to initialize celleryGcp")
	}
	clusters, err := gcpPlatform.GetClusterList()
	if err != nil {
		return false, fmt.Errorf("failed to get cluster list")
	}
	clusterNameSlice := strings.Split(cluster, gcpClusterName)
	if len(clusterNameSlice) > 1 {
		uniqueNumber := strings.Split(cluster, gcpClusterName)[1]
		if util.ContainsInStringArray(clusters, gcpClusterName+uniqueNumber) {
			valid = true
		}
	}
	return valid, nil
}

func RunCleanupGcp(value string) error {
	cleanupSpinner := util.StartNewSpinner("Removing GCP cluster")
	platform, err := gcp.NewGcp(gcp.SetClusterName(value))
	if err != nil {
		return fmt.Errorf("failed to initialize celleryGcp platform, %v", err)
	}
	err = platform.TearDown()
	if err != nil {
		cleanupSpinner.Stop(false)
	}
	return nil
}

func getGcpClustersCreatedByUser(cli cli.Cli, gcpClusters []string) ([]string, error) {
	var userCreatedGcpClusters []string
	// Get the list of clusters created by the user
	userCreatedClusters, err := getContexts(cli)
	if err != nil {
		return nil, fmt.Errorf("failed to get contexts, %v", err)
	}
	for _, gcpCluster := range gcpClusters {
		for _, userCreatedCluster := range userCreatedClusters {
			// Check if the given celleryGcp cluster is created by the user
			if strings.Contains(userCreatedCluster, gcpCluster) {
				// If the celleryGcp cluster is created by user add it to the list of celleryGcp clusters created by the user
				userCreatedGcpClusters = append(userCreatedGcpClusters, gcpCluster)
				break
			}
		}
	}
	return userCreatedGcpClusters, nil
}

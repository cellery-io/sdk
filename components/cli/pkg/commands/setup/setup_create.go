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
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/gcp"
	"cellery.io/cellery/components/cli/pkg/minikube"
	"cellery.io/cellery/components/cli/pkg/util"
)

func createEnvironment(cli cli.Cli) error {
	artifactsPath := filepath.Join(cli.FileSystem().UserHome(), constants.CelleryHome, constants.K8sArtifacts)
	os.RemoveAll(artifactsPath)
	util.CopyDir(filepath.Join(cli.FileSystem().CelleryInstallationDir(), constants.K8sArtifacts), artifactsPath)
	cli.Runtime().SetArtifactsPath(artifactsPath)
	bold := color.New(color.Bold).SprintFunc()
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}
	var items []string
	minikubeStatus, err := minikube.ClusterStatus(CelleryLocalSetup)
	if err != nil {
		return fmt.Errorf("failed to check if minikube is running, %v", err)
	}
	if minikubeStatus == minikube.Removed {
		items = []string{celleryLocal, celleryGcp, existingCluster, setupBack}
	} else {
		items = []string{celleryGcp, existingCluster, setupBack}
	}
	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select an environment to be installed",
		Items:     items,
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select environment to create option: %v", err)
	}

	switch value {
	case celleryGcp:
		{
			isCompleteSetup, isBackSelected, err := isCompleteSetupSelected()
			if err != nil {
				return fmt.Errorf("failed to select gcp basic or complete, %v", err)
			}
			if isBackSelected {
				return RunSetup(cli)
			}
			confirmed, _, err := util.GetYesOrNoFromUser("This will create a Cellery runtime on a gcp cluster. Do you want to continue", false)
			if err != nil {
				return fmt.Errorf("failed to get confirmation to create, %v", err)
			}
			if !confirmed {
				os.Exit(0)
			}
			platform, err := gcp.NewGcp()
			if err != nil {
				return fmt.Errorf("failed to initialize gcp platform, %v", err)
			}
			_, mysql, nfs, err := RunSetupCreateCelleryPlatform(cli, platform)
			if err != nil {
				return fmt.Errorf("failed to create gcp platform, %v", err)
			}
			if err := RunSetupCreateCelleryRuntime(cli, isCompleteSetup, true, true, true, nfs, mysql, ""); err != nil {
				cleanupErr := RunSetupCleanupPlatform(cli, platform, true)
				if cleanupErr != nil {
					return fmt.Errorf("failed to create cellery runtime on gcp cluster, %v. Failed to remove "+
						"partially created gcp platform, %v", err, cleanupErr)
				} else {
					return fmt.Errorf("failed to create cellery runtime on gcp cluster, %v", err)
				}
			}
		}
	case celleryLocal:
		{
			isCompleteSetup, isBackSelected, err := isCompleteSetupSelected()
			if isBackSelected {
				return RunSetup(cli)
			}
			confirmed, _, err := util.GetYesOrNoFromUser("This will create a Cellery runtime on a minikube cluster. Do you want to continue", false)
			if err != nil {
				return fmt.Errorf("failed to get confirmation to create, %v", err)
			}
			if !confirmed {
				os.Exit(0)
			}
			platform, err := minikube.NewMinikube(
				minikube.SetProfile(CelleryLocalSetup),
				minikube.SetCpus(MinikubeCpus),
				minikube.SetMemory(MinikubeMemory),
				minikube.SetkubeVersion(MinikubeKubernetesVersion))
			if err != nil {
				return fmt.Errorf("failed to initialize minikube platform, %v", err)
			}
			nodeportIp, mysql, nfs, err := RunSetupCreateCelleryPlatform(cli, platform)
			if err != nil {
				return fmt.Errorf("failed to create minikube platform, %v", err)
			}
			if err := RunSetupCreateCelleryRuntime(cli, isCompleteSetup, false, false, false, nfs, mysql, nodeportIp); err != nil {
				cleanupErr := RunSetupCleanupPlatform(cli, platform, true)
				if cleanupErr != nil {
					return fmt.Errorf("failed to create cellery runtime on minikube cluster, %v. Failed to remove "+
						"partially created minikube platform, %v", err, cleanupErr)
				} else {
					return fmt.Errorf("failed to create cellery runtime on minikube cluster, %v", err)
				}
			}
		}
	case existingCluster:
		{
			return createOnExistingCluster(cli)
		}
	default:
		{
			return RunSetup(cli)
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

func isCompleteSetupSelected() (bool, bool, error) {
	var isCompleteSelected = false
	var isBackSelected = false
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select the type of runtime",
		Items:     []string{setupBasic, setupComplete, setupBack},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return false, false, fmt.Errorf("failed to select an option: %v", err)
	}
	if value == setupBack {
		isBackSelected = true
	}
	if value == setupComplete {
		isCompleteSelected = true
	}
	return isCompleteSelected, isBackSelected, nil
}

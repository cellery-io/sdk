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

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
	"github.com/cellery-io/sdk/components/cli/pkg/vbox"
)

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
		return fmt.Errorf("failed to select environment to create option: %v", err)
	}

	switch value {
	case constants.CELLERY_SETUP_LOCAL:
		{
			createLocal()
		}
	case constants.CELLERY_SETUP_GCP:
		{
			createGcp()
		}
	case constants.CELLERY_SETUP_EXISTING_CLUSTER:
		{
			createOnExistingCluster()
		}
	default:
		{
			RunSetup()
			return nil
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

func getCreateEnvironmentList() []string {
	if util.IsCommandAvailable("VBoxManage") {
		if !vbox.IsVmInstalled() {
			return []string{constants.CELLERY_SETUP_LOCAL, constants.CELLERY_SETUP_GCP, constants.CELLERY_SETUP_EXISTING_CLUSTER, constants.CELLERY_SETUP_BACK}
		}
	}
	return []string{constants.CELLERY_SETUP_GCP, constants.CELLERY_SETUP_EXISTING_CLUSTER, constants.CELLERY_SETUP_BACK}
}

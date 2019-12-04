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

	"github.com/manifoldco/promptui"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/constants"
	gcpPlatform "cellery.io/cellery/components/cli/pkg/gcp"
	"cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
)

//var accountName string

func RunSetupCreateGcp(cli cli.Cli, isCompleteSetup bool) error {
	platform, err := gcpPlatform.NewGcp()
	if err != nil {
		return fmt.Errorf("failed to initialize celleryGcp platform, %v", err)
	}
	return RunSetupCreate(cli, platform, isCompleteSetup, true, true,
		true, runtime.Nfs{}, runtime.MysqlDb{}, "")
}

func createGcp(cli cli.Cli) error {
	var isCompleteSelected = false
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     util.Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     util.YellowBold("?") + " Select the type of runtime",
		Items:     []string{constants.BASIC, constants.COMPLETE, setupBack},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select an option: %v", err)
	}
	if value == setupBack {
		createEnvironment(cli)
		return nil
	}
	if value == constants.COMPLETE {
		isCompleteSelected = true
	}
	RunSetupCreateGcp(cli, isCompleteSelected)
	return nil
}

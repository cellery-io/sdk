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

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"

	cliPkg "cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
)

func RunSetup(cli cliPkg.Cli) error {
	err := cli.ExecuteUserSelection("Setup Cellery runtime", []cliPkg.Selection{
		{
			Number: 1,
			Label:  create,
			Function: func() error {
				if err := createEnvironment(cli); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Number: 2,
			Label:  manage,
			Function: func() error {
				if err := manageEnvironment(cli); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Number: 3,
			Label:  modify,
			Function: func() error {
				var err error
				apimEnabled, err = runtime.IsApimEnabled()
				if err != nil {
					return fmt.Errorf("failed check if apim is enabled, %v", err)
				}
				enableApim = !apimEnabled
				observabilityEnabled, err = runtime.IsObservabilityEnabled()
				if err != nil {
					return fmt.Errorf("failed check if observability is enabled, %v", err)
				}
				enableObservability = !observabilityEnabled
				knativeEnabled, err = runtime.IsKnativeEnabled()
				if err != nil {
					return fmt.Errorf("failed check if knative is enabled, %v", err)
				}
				enableKnative = !knativeEnabled
				hpaEnabled, err = cli.Runtime().IsHpaEnabled()
				if err != nil {
					return fmt.Errorf("failed check if hpa is enabled, %v", err)
				}
				enableHpa = !hpaEnabled
				if err := modifyRuntime(cli); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Number: 4,
			Label:  setupSwitch,
			Function: func() error {
				if err := selectEnvironment(cli); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Number: 5,
			Label:  exit,
			Function: func() error {
				os.Exit(1)
				return nil
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get user input, %v", err)
	}
	return nil
}

func selectEnvironment(cli cliPkg.Cli) error {
	contexts, err := getContexts(cli)
	if err != nil {
		return fmt.Errorf("failed to get contexts, %v", err)
	}
	contexts = append(contexts, setupBack)
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
		Items:     contexts,
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select cluster: %v", err)
	}

	if value == setupBack {
		return RunSetup(cli)
	}

	RunSetupSwitch(cli, value)
	fmt.Printf(util.GreenBold("\n\U00002714") + " Successfully configured Cellery.\n")
	fmt.Println()
	fmt.Println(bold("What's next ?"))
	fmt.Println("======================")
	fmt.Println("To create your first project, execute the command: ")
	fmt.Println("  $ cellery init ")
	return nil
}

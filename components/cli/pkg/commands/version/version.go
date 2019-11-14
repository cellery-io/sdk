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

package version

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/fatih/color"

	"github.com/cellery-io/sdk/components/cli/cli"
	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/version"
)

func RunVersion(cli cli.Cli) error {
	// Creating functions for printing version details
	white := color.New(color.FgWhite)
	boldWhite := white.Add(color.Bold)
	// Printing Cellery version information
	_, _ = boldWhite.Fprintln(cli.Out(), "Cellery:")
	fmt.Fprintf(cli.Out(), " CLI Version:\t\t%s\n", version.BuildVersion())
	fmt.Fprintf(cli.Out(), " OS/Arch:\t\t%s/%s\n", runtime.GOOS, runtime.GOARCH)
	// Printing Ballerina version information
	_, _ = boldWhite.Fprintln(cli.Out(), "\nBallerina:")
	balVersion, err := cli.BalExecutor().Version()
	if err != nil {
		// Having b7a locally is optional since the cellery build and run can be invoked via a docker container.
		fmt.Fprintln(cli.Out(), fmt.Sprintf("Ballerina %s not installed locally", constants.BallerinaVersion))
	} else {
		fmt.Fprintln(cli.Out(), " Version:\t\t"+strings.TrimSpace(balVersion))
	}
	// Printing Kubernetes version information
	serverVersion, clientVersion, err := cli.KubeCli().Version()
	if err == nil {
		_, _ = boldWhite.Fprintln(cli.Out(), "\nKubernetes")
		fmt.Fprintln(cli.Out(), " Server Version:\t"+serverVersion)
		fmt.Fprintln(cli.Out(), " Client Version:\t"+clientVersion)
		fmt.Fprintln(cli.Out(), " CRD:\t\t\ttrue") // TODO
	}
	// Printing Docker version information
	_, _ = boldWhite.Println("\nDocker:")
	dockerServerVersion, err := cli.DockerCli().ServerVersion()
	if err != nil {
		fmt.Fprintf(cli.Out(), "\x1b[31;1m Error while getting Docker Server version, \x1b[0m%v\n",
			strings.TrimSpace(dockerServerVersion))
	} else {
		fmt.Fprintf(cli.Out(), " Server Version:\t"+strings.TrimSpace(dockerServerVersion))
	}
	dockerClientVersion, err := cli.DockerCli().ClientVersion()
	if err != nil {
		fmt.Fprintf(cli.Out(), "\x1b[31;1m Error while getting Docker Client version, \x1b[0m%v\n",
			strings.TrimSpace(dockerClientVersion))
	} else {
		fmt.Fprintln(cli.Out(), " Client Version:\t"+strings.TrimSpace(dockerClientVersion))
	}
	return nil
}

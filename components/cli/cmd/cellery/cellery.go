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

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func newCliCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "cellery [OPTIONS] COMMAND [ARG...]",
		Short:         "Manage immutable cell based applications",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s, build %s", "0.1.0", "c69f31c"),
	}

	cmd.AddCommand(
		newConfigureCommand(),
		newCompletionCommand(cmd),
		newBuildCommand(),
		newImageCommand(),
		newVersionCommand(),
		newInitCommand(),
		newRunCommand(),
		newStopCommand(),
		newPsCommand(),
		newDescribeCommand(),
		newApisCommand(),
		newComponentsCommand(),
		newStatusCommand(),
		newLogsCommand(),
		newPushCommand(),
		newPullCommand(),
		newSetupCommand(),
	)
	return cmd
}

func main() {
	logFileDirectory := filepath.Join(util.UserHomeDir(), ".cellery", "logs")
	logFilePath := filepath.Join(logFileDirectory, "cli.log")

	// Creating the log directory if it does not exist
	err := util.CreateDir(logFileDirectory)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}

	// Setting the log output to the log file
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if logFile != nil {
		defer func() {
			err := logFile.Close()
			if err != nil {
				log.Fatalf("Failed to close log file: %v", err)
			}
		}()
	}
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
		log.Printf("Writing log to stdout")
	} else {
		log.SetOutput(logFile)
	}

	cmd := newCliCommand()
	if err := cmd.Execute(); err != nil {
		log.Fatal(fmt.Sprintf("%s: %s", "cellery", err))
		os.Exit(1)
	}
}

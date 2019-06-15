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
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func newCliCommand() *cobra.Command {
	var verboseMode = false
	cmd := &cobra.Command{
		Use:   "cellery <command>",
		Short: "Manage immutable cell based applications",
	}

	cmd.AddCommand(
		newCompletionCommand(cmd),
		newBuildCommand(),
		newVersionCommand(),
		newInitCommand(),
		newRunCommand(),
		newTerminateCommand(),
		newListCommand(),
		newDescribeCommand(),
		newStatusCommand(),
		newLogsCommand(),
		newLoginCommand(),
		newLogoutCommand(),
		newPushCommand(),
		newPullCommand(),
		newSetupCommand(),
		newExtractResourcesCommand(),
		newInspectCommand(),
		newViewCommand(),
		newDeleteImageCommand(),
		newExportPolicyCommand(),
		newApplyPolicyCommand(),
		newUpdateCellComponentsCommand(),
	)
	cmd.PersistentFlags().BoolVarP(&verboseMode, "verbose", "v", false, "Run on verbose mode")
	viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose"))
	return cmd
}

func main() {
	logFileDirectory := filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "logs")
	logFilePath := filepath.Join(logFileDirectory, "cli.log")

	// Creating the log directory if it does not exist
	err := util.CreateDir(logFileDirectory)
	if err != nil {
		log.Printf("Failed to create log file: %v", err)
	}

	// Setting the log output to the log file
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if logFile != nil {
		defer func() {
			err := logFile.Close()
			if err != nil {
				log.Printf("Failed to close log file: %v", err)
			}
		}()
	}
	if err != nil {
		log.Printf("Writing log to stdout because error occured while opening log file: %v", err)
	} else {
		log.SetOutput(logFile)
	}

	// copy cellery installation artifacts to user repo
	moduleMgr := &util.BLangManager{}
	if err := moduleMgr.Init(); err != nil {
		util.ExitWithErrorMessage("Unable to copy cellery installation artifacts to user repo", err)
	}

	cmd := newCliCommand()
	if err := cmd.Execute(); err != nil {
		util.ExitWithErrorMessage("Error executing cellery main function", err)
	}
}

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

package util

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

type LangManager interface {
	Init() error
	GetExecutablePath() (string, error)
}

type BLangManager struct{}

func (langMgr *BLangManager) Init() error {
	// if the module is not present in the cellery installation directory, skip trying to copy
	installPath := filepath.Join(CelleryInstallationDir(), "repo")
	var paths []string
	var err error
	celleryModPath := filepath.Join(installPath, "celleryio", "cellery", "*", "cellery.zip")
	if paths, err = filepath.Glob(celleryModPath); err != nil {
		return err
	}
	if len(paths) == 0 {
		// cellery module does not exist, can't copy
		log.Printf("Cellery module not found at %s, hence not copying to user repository \n", celleryModPath)
		return nil
	}

	userRepo := filepath.Join(UserHomeDir(), ".ballerina")
	// if not exists, create the location
	if _, err := os.Stat(userRepo); os.IsNotExist(err) {
		if err = os.Mkdir(userRepo, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	// copy from cellery installation location to user repository
	cmd := exec.Command("cp", "-r", installPath, userRepo)
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)

	execError := ""
	go func() {
		for stderrScanner.Scan() {
			execError += stderrScanner.Text()
		}
	}()
	if err := cmd.Start(); err != nil {
		log.Printf("Error: %v \n", execError)
		return err
	}
	if err := cmd.Wait(); err != nil {
		log.Printf("Error: %v \n", execError)
		return err
	}
	return nil
}

func (langMgr *BLangManager) GetExecutablePath() (string, error) {
	balVersionCmdOutput := ""
	cmd := exec.Command("ballerina", "version")
	stdoutReader, _ := cmd.StdoutPipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			balVersionCmdOutput += stdoutScanner.Text()
		}
	}()
	err := cmd.Start()
	if err != nil {
		return getBallerinaExecutablePath()
	}
	err = cmd.Wait()
	if err != nil {
		return getBallerinaExecutablePath()
	}
	if strings.Contains(balVersionCmdOutput, "Ballerina") {
		if len(strings.Split(balVersionCmdOutput, " ")) > 0 {
			if strings.Split(balVersionCmdOutput, " ")[1] == constants.BALLERINA_VERSION {
				// If existing ballerina version is as the expected version, execute ballerina run without executable path
				return "ballerina", nil
			}
		}
	}
	return getBallerinaExecutablePath()
}

func getBallerinaExecutablePath() (string, error) {
	exePath := BallerinaInstallationDir() + constants.BALLERINA_EXECUTABLE_PATH
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		exePath = strings.TrimSuffix(CelleryInstallationDir(), "/") + constants.CELLERY_EXECUTABLE_PATH
		if _, err := os.Stat(exePath); os.IsNotExist(err) {
			return "", nil
		} else if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	log.Printf("Executable path: %sballerina", exePath)
	return exePath + "ballerina", nil
}

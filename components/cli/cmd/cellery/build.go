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

package cellery

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tj/go-spin"
	"github.com/wso2/cellery/components/cli/pkg/constants"
	"github.com/wso2/cellery/components/cli/pkg/util"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var isSpinning = true
var isFirstPrint = true
var tag string
var fileName string

func newBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build CELL_FILE_NAME",
		Short: "Build an immutable cell image with required dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Printf("'cellery build' requires exactly 1 argument.\n" +
					"See 'cellery build --help' for more infomation.\n")
				return nil
			}
			fileName = args[0]
			err := runBuild(tag, fileName)
			if err != nil {
				cmd.Help()
				return err
			}
			return nil
		},
		Example: "  cellery build my-project.bal\n  cellery build my-project.bal -t myproject:1.0.0",
	}
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "Name and optionally a tag in the 'name:tag' format")
	return cmd
}

/**
Spinner
*/
func buildSpinner(tag string) {
	s := spin.New()
	for {
		if isSpinning {
			fmt.Printf("\r\033[36m%s\033[m Building %s %s", s.Next(), "image", util.Bold(tag))
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func runBuild(tag string, fileName string) error {

	fileExist, err := util.FileExists(fileName)
	if !fileExist {
		fmt.Printf("Please check the filename. File '%s' does not exist.\n", fileName)
		os.Exit(1)
	}

	var extension = filepath.Ext(fileName)
	var fileNameSuffix = fileName[0 : len(fileName)-len(extension)]

	viper.SetConfigName("Cellery") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")        // optionally look for config in the working directory
	confErr := viper.ReadInConfig() // Find and read the config file

	if confErr != nil { // Handle errors reading the config file
		fmt.Printf("Error while readng toml file: %s\n", confErr)
		os.Exit(1)
	}

	registryHost := constants.CENTRAL_REGISTRY_HOST
	organization := viper.GetString("project.organization")
	imageName := fileNameSuffix
	imageVersion := viper.GetString("project.version")

	if tag == "" {
		tag = organization + "/" + fileNameSuffix + ":" + imageVersion
	} else {
		strArr := strings.Split(tag, "/")
		if len(strArr) == 3 {
			registryHost = strArr[0]
			organization = strArr[1]
			imageTag := strings.Split(strArr[2], ":")
			if len(imageTag) != 2 {
				util.ExitWithImageFormatError()
			}
			imageName = imageTag[0]
			imageVersion = imageTag[1]
		} else if len(strArr) == 2 {
			organization = strArr[0]
			imageTag := strings.Split(strArr[1], ":")
			if len(imageTag) != 2 {
				util.ExitWithImageFormatError()
			}
			imageName = imageTag[0]
			imageVersion = imageTag[1]
		} else {
			util.ExitWithImageFormatError()
		}
	}

	repoLocation := filepath.Join(util.UserHomeDir(), ".cellery", "repos", registryHost, organization, imageName,
		imageVersion)
	go buildSpinner(tag)

	//first clean target directory if exists
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("Error in getting current directory location: " + err.Error())
		os.Exit(1)
	}
	_ = os.RemoveAll(filepath.Join(dir, "target"))

	cmd := exec.Command("ballerina", "run", fileName+":build")
	execError := ""
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			execError += stderrScanner.Text()
		}
	}()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Error in executing cell build: %v \n", err)
		errStr := string(stderr.Bytes())
		fmt.Printf("%s\n", errStr)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println()
		fmt.Printf("\x1b[31;1m\nBuild Failed.\x1b[0m %v \n", execError)
		fmt.Println("\x1b[31;1m======================\x1b[0m")
		errStr := string(stderr.Bytes())
		fmt.Printf("\x1b[31;1m%s\x1b[0m", errStr)
		os.Exit(1)
	}

	outStr := string(stdout.Bytes())
	fmt.Printf("\n\033[36m%s\033[m\n", outStr)

	folderCopyError := util.CopyDir(filepath.Join(dir, "target"), filepath.Join(dir, "artifacts"))
	if folderCopyError != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	folders := []string{"artifacts"}
	files := []string{fileName, constants.CONFIG_FILE}
	output := imageName + ".zip"
	err = util.RecursiveZip(files, folders, output)
	if err != nil {
		fmt.Printf("\x1b[31;1mCell build finished with error: \x1b[0m %v \n", err)
		os.Exit(1)
	}

	_ = os.RemoveAll(filepath.Join(dir, "artifacts"))

	repoCreateErr := util.CreateDir(repoLocation)
	if repoCreateErr != nil {
		fmt.Println("Error while creating image location: " + repoCreateErr.Error())
		os.Exit(1)
	}

	zipSrc := filepath.Join(dir, output)
	zipDst := filepath.Join(repoLocation, output)
	zipCopyError := util.CopyFile(zipSrc, zipDst)
	if zipCopyError != nil {
		fmt.Println("Error while saving image: " + zipCopyError.Error())
		os.Exit(1)
	}

	_ = os.Remove(zipSrc)

	fmt.Printf(util.GreenBold("\U00002714")+" Successfully built cell image: %s\n", util.Bold(tag))
	util.PrintWhatsNextMessage("cellery run " + tag)
	return nil
}

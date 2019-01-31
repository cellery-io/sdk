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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tj/go-spin"
	"github.com/wso2/cellery/components/cli/pkg/constants"
	"github.com/wso2/cellery/components/cli/pkg/util"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var cellImage string

type Response struct {
	Message string
	Image   ResponseImage
}
type ResponseImage struct {
	Organization  string
	Name          string
	ImageVersion  string
	ImageRevision string
}

func newPushCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push [CELL IMAGE]",
		Short: "push cell image to the remote repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
				return nil
			}
			cellImage = args[0]
			err := runPush(cellImage)
			if err != nil {
				cmd.Help()
				return err
			}
			return nil
		},
		Example: "  cellery push wso2/hello-world:1.0.0\n  cellery push registry.foo.com/foo/hello-world:1.0.0",
	}
	return cmd
}

/**
Spinner
*/
func pushSpinner(tag string) {
	s := spin.New()
	for {
		if isSpinning {
			fmt.Printf("\r\033[36m%s\033[m Pushing %s %s", s.Next(), "image", util.Bold(tag))
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func runPush(cellImage string) error {

	if cellImage == "" {
		return fmt.Errorf("no cell image specified")
	}

	registryHost := constants.CENTRAL_REGISTRY_HOST
	organization := ""
	imageName := ""
	imageVersion := ""

	strArr := strings.Split(cellImage, "/")
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

	go pushSpinner(cellImage)

	var url = "http://" + registryHost + constants.REGISTRY_BASE_PATH + "/images/" + organization +
		"/" + imageName + "/" + imageVersion

	path := filepath.Join(util.UserHomeDir(), ".cellery", "repos", registryHost, organization, imageName, imageVersion,
		imageName+constants.CELL_IMAGE_EXT)

	request, err := util.FileUploadRequest(url, nil, "file", path, true)
	if err != nil {
		fmt.Printf("\x1b[31;1m\nError occurred while pushing the cell image: \x1b[0m %v \n", err)
		os.Exit(1)
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	var responseBody string
	if err != nil {
		fmt.Printf("\x1b[31;1m Error occurred while pushing the cell image: \x1b[0m %v \n", err)
		os.Exit(1)
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		responseBody = body.String()
		resp.Body.Close()
	}

	if resp.StatusCode == 200 {
		var response Response
		err := json.Unmarshal([]byte(responseBody), &response)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println()
		fmt.Printf(util.GreenBold("\n\U00002714")+" Successfully pushed cell image: %s\n", util.Bold(cellImage))
		fmt.Println("Digest : " + util.Bold(response.Image.ImageRevision))
		util.PrintWhatsNextMessage("cellery pull " + cellImage)
	} else {
		fmt.Printf("\x1b[31;1mError occurred while pushing the cell image: \x1b[0m %v \n", err)
		os.Exit(1)
	}
	return nil
}

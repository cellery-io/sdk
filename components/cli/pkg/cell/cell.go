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

package cell

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func GetInstance(instance string) (*util.Cell, error) {
	cmd := exec.Command("kubectl", "get", "cell", instance, "-o", "json")
	outfile, err := os.Create("./" + instance + ".txt")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = outfile.Close()
		_ = os.Remove(outfile.Name())
	}()
	cmd.Stdout = outfile

	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)

	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	outputByteArray, err := ioutil.ReadFile(outfile.Name())
	os.Remove(outfile.Name())
	cell := util.Cell{}

	errJson := json.Unmarshal(outputByteArray, &cell)
	if errJson != nil {
		fmt.Println(errJson)
	}

	return &cell, nil
}

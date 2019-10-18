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

package kubernetes

import (
	"encoding/json"
	"os/exec"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/osexec"
)

// KubeCli represents kubernetes client.
type KubeCli interface {
	GetCells() ([]Cell, error)
}

type CelleryKubeCli struct {
}

// NewCelleryCli returns a CelleryCli instance.
func NewCelleryKubeCli() *CelleryKubeCli {
	kubeCli := &CelleryKubeCli{}
	return kubeCli
}

// GetCells returns mock cell instances array.
func (kubecli *CelleryKubeCli) GetCells() ([]Cell, error) {
	cmd := exec.Command(
		constants.KUBECTL,
		"get",
		"cells",
		"-o",
		"json",
	)
	jsonOutput := Cells{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return jsonOutput.Items, err
	}
	err = json.Unmarshal(out, &jsonOutput)
	return jsonOutput.Items, err
}

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

package runtime

import (
	"path/filepath"

	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/util"
)

func CreateCelleryNameSpace() error {
	for _, v := range buildNameSpaceYamlPaths() {
		err := kubernetes.ApplyFile(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildNameSpaceYamlPaths() []string {
	base := buildArtifactsPath(System, filepath.Join(util.CelleryInstallationDir(), constants.K8sArtifacts))
	return []string{
		filepath.Join(base, "ns-init.yaml"),
	}
}

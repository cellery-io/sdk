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

package runtime

import (
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/util"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
)

func createPersistentVolume(artifactsPath string, hasNfs bool) error {
	for _, v := range buildPersistentVolumeYamlPaths(artifactsPath, hasNfs) {
		err := kubectl.ApplyFileWithNamespace(v, "cellery-system")
		if err != nil {
			return err
		}
	}
	return nil
}

func updateNfsServerDetails(ipAddress, fileShare, artifactsPath string) error {
	for _, v := range buildPersistentVolumeYamlPaths(artifactsPath, true) {
		if err := util.ReplaceInFile(v, "NFS_SERVER_IP", ipAddress, -1); err != nil {
			util.ExitWithErrorMessage("Error updating file", err)
		}
		if err := util.ReplaceInFile(v, "NFS_SHARE_LOCATION", fileShare, -1); err != nil {
			util.ExitWithErrorMessage("Error updating file", err)
		}
	}
	return nil
}

func buildPersistentVolumeYamlPaths(artifactsPath string, hasNfs bool) []string {
	base := buildArtifactsPath(ApiManager, artifactsPath)
	if hasNfs {
		return []string{
			filepath.Join(base, "artifacts-persistent-volume.yaml"),
			filepath.Join(base, "artifacts-persistent-volume-claim.yaml"),
		}
	}
	return []string{
		filepath.Join(base, "persistent-volume-local-dev.yaml"),
		filepath.Join(base, "persistent-volume-claim-local.yaml"),
	}
}

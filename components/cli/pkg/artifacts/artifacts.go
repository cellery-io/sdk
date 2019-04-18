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

package artifacts

import (
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func buildArtifactsPath(component SystemComponent) string {
	artifactsPath := filepath.Join(util.UserHomeDir(), ".cellery", "gcp", "artifacts", "k8s-artefacts")
	switch component {
	case ApiManager:
		return filepath.Join(artifactsPath, "global-apim")
	case IdentityProvider:
		return filepath.Join(artifactsPath, "global-idp")
	case Observability:
		return filepath.Join(artifactsPath, "observability")
	case Controller:
		return filepath.Join(artifactsPath, "controller")
	case System:
		return filepath.Join(artifactsPath, "system")
	case Mysql:
		return filepath.Join(artifactsPath, "mysql")
	default:
		return filepath.Join(artifactsPath)
	}
}

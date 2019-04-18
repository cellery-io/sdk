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
	"fmt"
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func AddComponent(component SystemComponent) error {
	switch component {
	case ApiManager:
		return addApim()
	case IdentityProvider:
		return addIdp()
	case Observability:
		return addObservability()

	default:
		return fmt.Errorf("unknown system componenet %q", component)
	}
}

func DeleteComponent(component SystemComponent) error {
	switch component {
	case ApiManager:
		return deleteApim()
	case IdentityProvider:
		return deleteIdp()
	case Observability:
		return deleteObservability()
	default:
		return fmt.Errorf("unknown system componenet %q", component)
	}
}

func UpdateRuntime(apiManagement, observability bool) error {
	var err error
	err = DeleteComponent(Observability)
	if err != nil {
		return err
	}
	if apiManagement {
		err = DeleteComponent(IdentityProvider)
		if err != nil {
			return err
		}
		err = AddComponent(ApiManager)
		if err != nil {
			return err
		}
	} else {
		err = DeleteComponent(ApiManager)
		if err != nil {
			return err
		}
		err = AddComponent(IdentityProvider)
		if err != nil {
			return err
		}
	}

	if observability {
		err = AddComponent(Observability)
		if err != nil {
			return err
		}
	} else {
		err = DeleteComponent(Observability)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildArtifactsPath(component SystemComponent) string {
	artifactsPath := filepath.Join(util.CelleryInstallationDir(), "k8s-artefacts")
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
	default:
		return filepath.Join(artifactsPath)
	}
}

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

	"cellery.io/cellery/components/cli/cli"
)

func RunSetNamespace(cli cli.Cli, namespace string) error {
	if err := cli.ExecuteTask(fmt.Sprintf("Changing the namespace to %s", namespace),
		fmt.Sprintf("Failed to changed the namespace to %s", namespace),
		"", func() error {
			err := cli.KubeCli().CreateNamespace(namespace)
			if err != nil {
				return fmt.Errorf("failed to create namespace, %v", err)
			}
			err = cli.KubeCli().SetNamespace(namespace)
			if err != nil {
				return fmt.Errorf("failed to set namespace, %v", err)
			}
			err = cli.KubeCli().ApplyLabel("namespace", namespace, "istio-injection=enabled",
				true)
			if err != nil {
				return fmt.Errorf("failed enable istio injection, %v", err)
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to change the namespace, %v", err)
	}
	return nil
}

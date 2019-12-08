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

package setup

import (
	"fmt"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
)

func RunSetupCleanup(cli cli.Cli, platform cli.Platform, removeKnative, removeIstio, removeIngress, removeHpa, confirmed bool) error {
	var err error
	var confirmCleanup = confirmed
	if !confirmed {
		confirmCleanup, _, err = util.GetYesOrNoFromUser("Do you want to delete the cellery runtime (This will "+
			"delete all your cells and data)", false)
		if err != nil {
			return fmt.Errorf("failed to select option, %v", err)
		}
	}
	if confirmCleanup {
		if platform != nil {
			if err := cli.ExecuteTask("Removing kubernetes cluster", "Failed to remove kubernetes cluster",
				"", func() error {
					return platform.RemoveCluster()
				}); err != nil {
				return err
			}
			if err := cli.ExecuteTask("Removing sql instance", "Failed to remove sql instance",
				"", func() error {
					return platform.RemoveSqlInstance()
				}); err != nil {
				return err
			}
			if err := cli.ExecuteTask("Removing file system", "Failed to remove file system",
				"", func() error {
					return platform.RemoveFileSystem()
				}); err != nil {
				return err
			}
			if err := cli.ExecuteTask("Removing storage", "Failed to remove storage",
				"", func() error {
					return platform.RemoveStorage()
				}); err != nil {
				return err
			}
			return nil
		}
		if err := cli.ExecuteTask("Removing cellery system", "Failed to remove cellery system",
			"", func() error {
				return cli.KubeCli().DeleteNameSpace("cellery-system")
			}); err != nil {
			return err
		}
		if removeKnative {
			if err := cli.ExecuteTask("Removing knative serving", "Failed to remove knative serving",
				"", func() error {
					return cli.Runtime().DeleteComponent(runtime.ScaleToZero)
				}); err != nil {
				return err
			}
		}
		if removeIstio {
			if err := cli.ExecuteTask("Removing istio system", "Failed to remove istio system",
				"", func() error {
					return cli.KubeCli().DeleteNameSpace("istio-system")
				}); err != nil {
				return err
			}
		}
		if removeIngress {
			if err := cli.ExecuteTask("Removing ingress-nginx", "Failed to remove ingress-nginx",
				"", func() error {
					return cli.KubeCli().DeleteNameSpace("ingress-nginx")
				}); err != nil {
				return err
			}
		}
		if removeHpa {
			if err := cli.ExecuteTask("Removing hpa", "Failed to remove hpa",
				"", func() error {
					return cli.Runtime().DeleteComponent(runtime.HPA)
				}); err != nil {
				return err
			}
		}
		if err := cli.ExecuteTask("Removing all cells", "Failed to remove all cells",
			"", func() error {
				return cli.KubeCli().DeleteAllCells()
			}); err != nil {
			return err
		}

		if err := cli.ExecuteTask("Removing persistent volume", "Failed to remove persistent volume",
			"", func() error {
				if err := cli.KubeCli().DeletePersistedVolume("wso2apim-local-pv"); err != nil {
					return fmt.Errorf("failed to remove persisted volumes-apim, %v", err)
				}
				if err := cli.KubeCli().DeletePersistedVolume("wso2apim-with-analytics-mysql-pv"); err != nil {
					return fmt.Errorf("failed to remove persisted volumes-mysql, %v", err)
				}
				return nil
			}); err != nil {
			return err
		}
	}
	return nil
}

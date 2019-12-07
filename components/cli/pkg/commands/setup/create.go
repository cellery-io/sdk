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
	"os"
	"path/filepath"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/runtime"
	"cellery.io/cellery/components/cli/pkg/util"
)

func RunSetupCreate(cli cli.Cli, platform cli.Platform, complete bool, isPersistentVolume, hasNfsStorage, isLoadBalancerIngressMode bool,
	nfs runtime.Nfs, db runtime.MysqlDb, nodePortIpAddress string) error {
	var err error
	artifactsPath := filepath.Join(cli.FileSystem().UserHome(), constants.CelleryHome, constants.K8sArtifacts)
	os.RemoveAll(artifactsPath)
	util.CopyDir(filepath.Join(cli.FileSystem().CelleryInstallationDir(), constants.K8sArtifacts), artifactsPath)
	cli.Runtime().SetArtifactsPath(artifactsPath)
	if platform != nil {
		// Create platform resources.
		if err := cli.ExecuteTask("Creating k8s cluster", "Failed to create k8s cluster",
			"", func() error {
				return platform.CreateK8sCluster()
			}); err != nil {
			return err
		}
		if err := cli.ExecuteTask("Creating sql instance", "Failed to create sql instance",
			"", func() error {
				db, err = platform.ConfigureSqlInstance()
				return err
			}); err != nil {
			return err
		}
		if err := cli.ExecuteTask("Creating storage", "Failed to create storage",
			"", func() error {
				return platform.CreateStorage()
			}); err != nil {
			return err
		}
		if err := cli.ExecuteTask("Creating file system", "Failed to create file system",
			"", func() error {
				nfs, err = platform.CreateNfs()
				return err
			}); err != nil {
			return err
		}
		if err := cli.ExecuteTask("Updating kube config", "Failed to update kube config",
			"", func() error {
				return platform.UpdateKubeConfig()
			}); err != nil {
			return err
		}
	}
	if err := cli.Runtime().Validate(); err != nil {
		return fmt.Errorf("runtime validation failed. %v", err)
	}
	// Install Cellery runtime.
	if isPersistentVolume && !hasNfsStorage {
		if err := cli.Runtime().CreatePersistentVolumeDirs(); err != nil {
			return fmt.Errorf("failed to create persistent volume directories, %v", err)
		}
	}
	dbHostName := constants.MysqlHostNameForExistingCluster
	dbUserName := constants.CellerySqlUserName
	dbPassword := constants.CellerySqlPassword
	if hasNfsStorage {
		dbHostName = db.DbHostName
		dbUserName = db.DbUserName
		dbPassword = db.DbPassword
		if err := cli.Runtime().UpdateNfsServerDetails(nfs.NfsServerIp, nfs.FileShare); err != nil {
			return fmt.Errorf("failed to update nfs server details")
		}
	}
	if err := cli.Runtime().UpdateMysqlCredentials(dbUserName, dbPassword, dbHostName); err != nil {
		return fmt.Errorf("error updating mysql credentials: %v", err)
	}
	if err := cli.Runtime().UpdateInitSql(dbUserName, dbPassword); err != nil {
		return fmt.Errorf("error updating mysql init script: %v", err)
	}

	if isPersistentVolume && !cli.Runtime().IsGcpRuntime() {
		nodeName, err := cli.KubeCli().GetMasterNodeName()
		if err != nil {
			return fmt.Errorf("error getting master node name: %v", err)
		}
		if err := cli.KubeCli().ApplyLabel("nodes", nodeName, "disk=local", true); err != nil {
			return fmt.Errorf("error applying master node lable: %v", err)
		}
	}
	// Setup Cellery namespace
	if err := cli.ExecuteTask("Creating Cellery namespace", "Failed to create Cellery namespace",
		"", func() error {
			return cli.Runtime().CreateCelleryNameSpace()
		}); err != nil {
		return fmt.Errorf("error creating cellery namespace: %v", err)
	}
	// Apply Istio CRDs
	if err := cli.ExecuteTask("Creating Istio CRDS", "Failed to create Istio CRDS",
		"", func() error {
			return cli.Runtime().ApplyIstioCrds()
		}); err != nil {
		return fmt.Errorf("error creating istio crds: %v", err)
	}
	// Apply nginx resources
	if err := cli.ExecuteTask("Installing ingress-nginx", "Failed to install ingress-nginx",
		"", func() error {
			return cli.Runtime().InstallIngressNginx(isLoadBalancerIngressMode)
		}); err != nil {
		return fmt.Errorf("error installing ingress-nginx: %v", err)
	}
	// sleep for few seconds - this is to make sure that the CRDs are properly applied
	cli.Sleep(20)

	// Enabling Istio injection
	if err := cli.ExecuteTask("Enabling istio injection", "Failed to enabling istio injection",
		"", func() error {
			return cli.KubeCli().ApplyLabel("namespace", "default", "istio-injection=enabled",
				true)
		}); err != nil {
		return fmt.Errorf("error enabling istio injection: %v", err)
	}
	if err := cli.ExecuteTask("Installing istio", "Failed to install istio",
		"", func() error {
			return cli.Runtime().InstallIstio()
		}); err != nil {
		return fmt.Errorf("error installing istio: %v", err)
	}
	if cli.Runtime().IsGcpRuntime() {
		// Install knative serving
		if err := cli.ExecuteTask("Installing knative", "Failed to install knative",
			"", func() error {
				return cli.Runtime().InstallKnativeServing()
			}); err != nil {
			return fmt.Errorf("error installing knative: %v", err)
		}
	}
	// Apply only knative serving CRD's
	if err := cli.ExecuteTask("Applying knative CRDs", "Failed to apply knative CRDs",
		"", func() error {
			return cli.Runtime().ApplyKnativeCrds()
		}); err != nil {
		return fmt.Errorf("error installing knative serving: %v", err)
	}
	// Apply controller CRDs
	if err := cli.ExecuteTask("Installing controller", "Failed to install controller",
		"", func() error {
			return cli.Runtime().InstallController()
		}); err != nil {
		return fmt.Errorf("error creating cellery controller: %v", err)
	}
	if !cli.Runtime().IsGcpRuntime() {
		if err := cli.ExecuteTask("Configuring mysql", "Failed to configure mysql",
			"", func() error {
				return cli.Runtime().InstallMysql(isPersistentVolume)
			}); err != nil {
			return fmt.Errorf("error configuring mysql: %v", err)
		}
	}
	if err := cli.ExecuteTask("Creating config maps", "Failed to create config maps",
		"", func() error {
			return cli.Runtime().CreateConfigMaps()
		}); err != nil {
		return fmt.Errorf("error creating configmaps: %v", err)
	}
	if err := cli.ExecuteTask("Creating persistent volume", "Failed to create persistent volume",
		"", func() error {
			return cli.Runtime().CreatePersistentVolume(hasNfsStorage)
		}); err != nil {
		return fmt.Errorf("error creating persistent volume: %v", err)
	}
	if complete {
		if err := cli.ExecuteTask("Creating apim deployment", "Failed to create apim deployment",
			"", func() error {
				return cli.Runtime().AddApim(isPersistentVolume)
			}); err != nil {
			return fmt.Errorf("error creating apim deployment: %v", err)
		}
		if err := cli.ExecuteTask("Creating observability deployment", "Failed to create observability deployment",
			"", func() error {
				return cli.Runtime().AddObservability()
			}); err != nil {
			return fmt.Errorf("error creating observability deployment: %v", err)
		}
	} else {
		if err := cli.ExecuteTask("Creating idp deployment", "Failed to create idp deployment",
			"", func() error {
				return cli.Runtime().AddIdp()
			}); err != nil {
			return fmt.Errorf("error creating idp deployment: %v", err)
		}
	}
	if !isLoadBalancerIngressMode {
		if nodePortIpAddress != "" {
			if err := cli.ExecuteTask("Updating nodeport ip address", "Failed to update nodeport ip address",
				"", func() error {
					return cli.Runtime().UpdateNodePortIpAddress(nodePortIpAddress)
				}); err != nil {
				return fmt.Errorf("failed to update nodeport ip address, %v", err)
			}
		}
	}
	return cli.Runtime().WaitFor(false, false)
}

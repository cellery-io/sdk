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

	"github.com/cellery-io/sdk/components/cli/kubernetes"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func AddMysql(artifactsPath string, isPersistentVolume bool) error {
	base := buildArtifactsPath(Mysql, artifactsPath)
	for _, confMap := range buildMysqlConfigMaps(artifactsPath) {
		if err := kubernetes.CreateConfigMapWithNamespace(confMap.Name, confMap.Path, "cellery-system"); err != nil {
			return err
		}
	}
	if isPersistentVolume {
		for _, persistentVolumeYaml := range buildPersistentVolumePaths(artifactsPath) {
			if err := kubernetes.ApplyFileWithNamespace(persistentVolumeYaml, "cellery-system"); err != nil {
				return err
			}
		}
	}
	if err := kubernetes.ApplyFileWithNamespace(buildMysqlDeploymentPath(artifactsPath, isPersistentVolume), "cellery-system"); err != nil {
		return err
	}
	if err := kubernetes.WaitForDeployment("available", 900,
		"wso2apim-with-analytics-mysql-deployment", "cellery-system"); err != nil {
		return err
	}
	if err := kubernetes.ApplyFileWithNamespace(filepath.Join(base, "mysql-service.yaml"), "cellery-system"); err != nil {
		return err
	}
	return nil
}

func updateMysqlCredentials(dbUserName, dbPassword, dbHost, artifactsPath string) error {
	for _, file := range buildMysqlConfigFilesPath(artifactsPath) {
		if err := util.ReplaceInFile(file, "DATABASE_USERNAME", dbUserName, -1); err != nil {
			return err
		}
		if err := util.ReplaceInFile(file, "DATABASE_PASSWORD", dbPassword, -1); err != nil {
			return err
		}
		if err := util.ReplaceInFile(file, "MYSQL_DATABASE_HOST", dbHost, -1); err != nil {
			return err
		}
	}
	return nil
}

func updateInitSql(dbUserName, dbPassword, artifactsPath string) error {
	if err := util.ReplaceInFile(buildInitSqlPath(artifactsPath), "DATABASE_USERNAME", dbUserName, -1); err != nil {
		return err
	}
	if err := util.ReplaceInFile(buildInitSqlPath(artifactsPath), "DATABASE_PASSWORD", dbPassword, -1); err != nil {
		return err
	}
	return nil
}

func buildMysqlDeploymentPath(artifactsPath string, isPersistentVolume bool) string {
	base := buildArtifactsPath(Mysql, artifactsPath)
	if isPersistentVolume {
		return filepath.Join(base, "mysql-deployment.yaml")
	}
	return filepath.Join(base, "mysql-deployment-volatile.yaml")
}

func buildMysqlConfigMaps(artifactsPath string) []ConfigMap {
	base := buildArtifactsPath(Mysql, artifactsPath)
	return []ConfigMap{
		{"mysql-dbscripts", filepath.Join(base, "dbscripts")},
	}
}

func buildMysqlConfigFilesPath(artifactsPath string) []string {
	var configFiles []string
	configFiles = append(configFiles, filepath.Join(buildArtifactsPath(ApiManager, artifactsPath), "conf", "datasources",
		"master-datasources.xml"))
	configFiles = append(configFiles, filepath.Join(buildArtifactsPath(Observability, artifactsPath), "sp", "conf",
		"deployment.yaml"))
	configFiles = append(configFiles, filepath.Join(buildArtifactsPath(IdentityProvider, artifactsPath), "conf",
		"datasources", "master-datasources.xml"))

	return configFiles
}

func buildPersistentVolumePaths(artifactsPath string) []string {
	base := buildArtifactsPath(Mysql, artifactsPath)
	return []string{
		filepath.Join(base, "mysql-persistent-volumes-local-dev.yaml"),
		filepath.Join(base, "mysql-persistent-volume-claim.yaml"),
	}
}

func buildInitSqlPath(artifactsPath string) string {
	base := buildArtifactsPath(Mysql, artifactsPath)
	return filepath.Join(base, "dbscripts", "init.sql")
}

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

package gcp

import (
	"path/filepath"

	"cellery.io/cellery/components/cli/pkg/runtime"

	"cellery.io/cellery/components/cli/pkg/util"
)

func UpdateMysqlCredentials(dbUserName, dbPassword, dbHost string) error {
	for _, file := range buildMysqlConfigFilesPath() {
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

func UpdateInitSql(dbUserName, dbPassword string) error {
	if err := util.ReplaceInFile(buildInitSqlPath(), "DATABASE_USERNAME", dbUserName, -1); err != nil {
		return err
	}
	if err := util.ReplaceInFile(buildInitSqlPath(), "DATABASE_PASSWORD", dbPassword, -1); err != nil {
		return err
	}
	return nil
}

func buildMysqlConfigFilesPath() []string {
	var configFiles []string
	configFiles = append(configFiles, filepath.Join(buildArtifactsPath(runtime.ApiManager), "conf", "datasources",
		"master-datasources.xml"))
	configFiles = append(configFiles, filepath.Join(buildArtifactsPath(runtime.Observability), "sp", "conf", "deployment.yaml"))
	configFiles = append(configFiles, filepath.Join(buildArtifactsPath(runtime.IdentityProvider), "conf", "datasources",
		"master-datasources.xml"))

	return configFiles
}

func buildInitSqlPath() string {
	base := buildArtifactsPath(runtime.Mysql)
	return filepath.Join(base, "dbscripts", "init.sql")
}

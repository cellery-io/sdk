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

const clusterNamePrefix = "cellery-cluster"
const dbInstanceNamePrefix = "cellery-sql"
const storagePrefix = "cellery-gcp-bucket"
const sqlUserName = "cellery-sql-user"
const sqlPassword = "cellery-sql-user"
const fileStorePrefix = "nfs-server"
const fileStoreFileShare = "data"
const fileStoreCapacity = 1024
const clusterImageType = "cos_containerd"
const clusterMachineType = "n1-standard-4"
const clusterDiskSizeGb = 100
const sqlTier = "db-n1-standard-1"
const sqlDiskSizeGb = 20
const initSql = "init.sql"
const dbScripts = "dbscripts"

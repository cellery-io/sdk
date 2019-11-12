/*
 * Copyright (c) 2018 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package constants

const DOMAIN_NAME_PATTERN = "[a-z0-9]+((?:-|.)[a-z0-9]+)*(:[0-9]+)?"
const CELLERY_ID_PATTERN = "[a-z0-9]+(-[a-z0-9]+)*"
const CELLERY_ALIAS_PATTERN = "[a-zA-Z0-9_-]+"
const IMAGE_VERSION_PATTERN = "[a-z0-9]+((?:-|.)[a-z0-9]+)*"
const CELL_IMAGE_PATTERN = CELLERY_ID_PATTERN + "\\/" + CELLERY_ID_PATTERN + ":" + IMAGE_VERSION_PATTERN
const DEPENDENCY_LINK_PATTERN = "(" + CELLERY_ID_PATTERN + "\\.)?" + CELLERY_ALIAS_PATTERN + ":" + CELLERY_ID_PATTERN

const CLI_ARG_ENV_VAR_KEY_PATTERN = "(?P<key>[^:]+)"
const CLI_ARG_ENV_VAR_VALUE_PATTERN = "(?P<value>.+)"
const CLI_ARG_ENV_VAR_PATTERN = "(((?P<instance>" + CELLERY_ID_PATTERN + "):" +
	CLI_ARG_ENV_VAR_KEY_PATTERN + "=" + CLI_ARG_ENV_VAR_VALUE_PATTERN + ")|(" +
	CLI_ARG_ENV_VAR_KEY_PATTERN + "=" + CLI_ARG_ENV_VAR_VALUE_PATTERN + "))"

const GROUP_NAME = "mesh.cellery.io"

const CELL_IMAGE_EXT = ".zip"
const JSON_EXT = ".json"

const CentralRegistryHost = "registry.hub.cellery.io"
const CENTRAL_REGISTRY_HOST_REGX = "^.*registry.hub.cellery.(io|net)$"

const CELLERY_SETUP_MANAGE = "Manage"
const CELLERY_SETUP_CREATE = "Create"
const CELLERY_SETUP_MODIFY = "Modify"
const CELLERY_SETUP_SWITCH = "Switch"
const CELLERY_SETUP_BACK = "BACK"
const CELLERY_SETUP_EXIT = "EXIT"

const CELLERY_SETUP_LOCAL = "Local"
const CELLERY_SETUP_EXISTING_CLUSTER = "Existing cluster"
const CELLERY_SETUP_GCP = "GCP"

const CELLERY_MANAGE_STOP = "stop"
const CELLERY_MANAGE_START = "start"
const CELLERY_MANAGE_CLEANUP = "cleanup"

const AWS_S3_BUCKET = "cellery-runtime-installation"
const AWS_REGION = "ap-south-1"

const VM_NAME = "cellery-runtime-local"

const GCP_CLUSTER_NAME = "cellery-cluster"
const GCP_DB_INSTANCE_NAME = "cellery-sql"
const GCP_BUCKET_NAME = "cellery-gcp-bucket"
const GCP_SQL_USER_NAME = "cellery-sql-user"
const GCP_SQL_PASSWORD = "cellery-sql-user"
const MYSQL_HOST_NAME_FOR_EXISTING_CLUSTER = "wso2apim-with-analytics-rdbms-service"
const GCP_NFS_SERVER_INSTANCE = "nfs-server"

const GCP_NFS_CONFIG_NAME = "data"
const GCP_NFS_CONFIG_CAPACITY = 1024
const GCP_CLUSTER_IMAGE_TYPE = "cos_containerd"
const GCP_CLUSTER_MACHINE_TYPE = "n1-standard-4"
const GCP_CLUSTER_DISK_SIZE_GB = 100
const GCP_SQL_TIER = "db-n1-standard-1"
const GCP_SQL_DISK_SIZE_GB = 20

const ZIP_BALLERINA_SOURCE = "src"
const ZIP_ARTIFACTS = "artifacts"
const ZIP_TESTS = "tests"
const ZIP_META_SUFFIX = "_meta"

const CELLERY_HOME_DOCS_VIEW_DIR = "docs-view"

const CELLERY_HOME = ".cellery"
const GCP = "gcp"
const K8S_ARTIFACTS = "k8s-artefacts"
const VM = "vm"
const MYSQL = "mysql"
const SP = "sp"
const DB_SCRIPTS = "dbscripts"
const INIT_SQL = "init.sql"
const TELEPRESENCE = "telepresence"

const KUBECTL = "kubectl"
const CREATE = "create"
const DELETE = "delete"
const BASIC = "Basic"
const COMPLETE = "Complete"

const CELLERY_INSTALLATION_PATH_MAC = "/Library/Cellery"
const CELLERY_INSTALLATION_PATH_UBUNTU = "/usr/share/cellery"
const CELLERY_EXECUTABLE_PATH = "/runtime/executable/"

const BALLERINA_VERSION = "1.0.3"
const BALLERINA_INSTALLATION_PATH_MAC = "/Library/Ballerina/ballerina-" + BALLERINA_VERSION
const BALLERINA_INSTALLATION_PATH_UBUNTU = "/usr/lib/ballerina/ballerina-" + BALLERINA_VERSION
const BALLERINA_EXECUTABLE_PATH = "/bin/"
const BALLERINA_CONF = "ballerina.conf"
const BALLERINA_TOML = "Ballerina.toml"
const BALLERINA_LOCAL_REPO = ".ballerina/"
const TEMP_TEST_MODULE = "tmp"

const DOCKER_CLI_BALLERINA_EXECUTABLE_PATH = "/usr/lib/ballerina/ballerina-" + BALLERINA_VERSION + "/bin/ballerina"

const WSO2_APIM_HOST = "https://wso2-apim-gateway"

const CELLERY_IMAGE_DIR_ENV_VAR = "CELLERY_IMAGE_DIR"
const TEST_DEGUB_FLAG = "DEBUG_MODE"

const ROOT_DIR = "/"
const VAR = "var"
const TMP = "tmp"
const CELLERY = "cellery"
const APIM_REPOSITORY_DEPLOYMENT_SERVER = "apim_repository_deployment_server"

const PERSISTENT_VOLUME = "Persistent volume"
const NON_PERSISTENT_VOLUME = "Non persistent volume"

const CELLERY_SQL_USER_NAME = "cellery-sql-user"
const CELLERY_SQL_PASSWORD = "cellery-sql-user"

const INGRESS_MODE_NODE_PORT = "Node port [kubeadm, minikube]"
const INGRESS_MODE_LOAD_BALANCER = "Load balancer [gcp, docker for desktop]"

const CELLERY_DOCKER_CLI_USER_ID = "1000"

const TELEPRESENCE_EXEC_PATH = "telepresence-0.101/bin/"

const IP_ADDRESS_PATTERN = "(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])"

const HTTP_API_INGRESS = "Http"
const WEB_INGRESS = "Web"
const GRPC_INGRESS = "GRPC"
const TCP_INGRESS = "TCP"
const N_A = "N/A"
const GATEWAY_HOST = "gateway_host"
const PORT = "port"
const HOST = "host"

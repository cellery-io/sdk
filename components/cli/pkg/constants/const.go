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

const DomainNamePattern = "[a-z0-9]+((?:-|.)[a-z0-9]+)*(:[0-9]+)?"
const CelleryIdPattern = "[a-z0-9]+(-[a-z0-9]+)*"
const CelleryAliasPattern = "[a-zA-Z0-9_-]+"
const ImageVersionPattern = "[a-z0-9]+((?:-|.)[a-z0-9]+)*"
const CellImagePattern = CelleryIdPattern + "\\/" + CelleryIdPattern + ":" + ImageVersionPattern
const DependencyLinkPattern = "(" + CelleryIdPattern + "\\.)?" + CelleryAliasPattern + ":" + CelleryIdPattern

const CliArgEnvVarKeyPattern = "(?P<key>[^:]+)"
const CliArgEnvVarValuePattern = "(?P<value>.+)"
const CliArgEnvVarPattern = "(((?P<instance>" + CelleryIdPattern + "):" +
	CliArgEnvVarKeyPattern + "=" + CliArgEnvVarValuePattern + ")|(" +
	CliArgEnvVarKeyPattern + "=" + CliArgEnvVarValuePattern + "))"

const GroupName = "mesh.cellery.io"

const CellImageExt = ".zip"
const JsonExt = ".json"

const CentralRegistryHost = "registry.hub.cellery.io"
const CentralRegistryHostRegx = "^.*registry.hub.cellery.(io|net)$"

const CellerySetupManage = "Manage"
const CellerySetupCreate = "Create"
const CellerySetupModify = "Modify"
const CellerySetupSwitch = "Switch"
const CellerySetupBack = "BACK"
const CellerySetupExit = "EXIT"

const CellerySetupLocal = "Local"
const CellerySetupExistingCluster = "Existing cluster"
const CellerySetupGcp = "GCP"

const CelleryManageStop = "stop"
const CelleryManageStart = "start"
const CelleryManageCleanup = "cleanup"

const AwsS3Bucket = "cellery-runtime-installation"
const AwsRegion = "ap-south-1"

const VmName = "cellery-runtime-local"

const GcpClusterName = "cellery-cluster"
const GcpDbInstanceName = "cellery-sql"
const GcpBucketName = "cellery-gcp-bucket"
const GcpSqlUserName = "cellery-sql-user"
const GcpSqlPassword = "cellery-sql-user"
const MysqlHostNameForExistingCluster = "wso2apim-with-analytics-rdbms-service"
const GcpNfsServerInstance = "nfs-server"

const GcpNfsConfigName = "data"
const GcpNfsConfigCapacity = 1024
const GcpClusterImageType = "cos_containerd"
const GcpClusterMachineType = "n1-standard-4"
const GcpClusterDiskSizeGb = 100
const GcpSqlTier = "db-n1-standard-1"
const GcpSqlDiskSizeGb = 20

const ZipBallerinaSource = "src"
const ZipArtifacts = "artifacts"
const ZipTests = "tests"
const ZipMetaSuffix = "_meta"

const CelleryHome = ".cellery"
const GCP = "gcp"
const K8sArtifacts = "k8s-artefacts"
const VM = "vm"
const MySql = "mysql"
const SP = "sp"
const DbScripts = "dbscripts"
const InitSql = "init.sql"
const TELEPRESENCE = "telepresence"

const KubeCtl = "kubectl"
const CREATE = "create"
const DELETE = "delete"
const BASIC = "Basic"
const COMPLETE = "Complete"

const CelleryInstallationPathMac = "/Library/Cellery"
const CelleryInstallationPathUbuntu = "/usr/share/cellery"
const CelleryExecutablePath = "/runtime/executable/"

const BallerinaVersion = "1.0.3"
const BallerinaInstallationPathMac = "/Library/Ballerina/ballerina-" + BallerinaVersion
const BallerinaInstallationPathUbuntu = "/usr/lib/ballerina/ballerina-" + BallerinaVersion
const BallerinaExecutablePath = "/bin/"
const BallerinaConf = "ballerina.conf"
const BallerinaToml = "Ballerina.toml"
const TARGET_DIR_NAME = "target"
const TempTestModule = "tmp"

const DockerCliBallerinaExecutablePath = "/usr/lib/ballerina/ballerina-" + BallerinaVersion + "/bin/ballerina"

const Wso2ApimHost = "https://wso2-apim-gateway"

const CelleryImageDirEnvVar = "CELLERY_IMAGE_DIR"
const TestDegubFlag = "DEBUG_MODE"

const RootDir = "/"
const VAR = "var"
const TMP = "tmp"
const CELLERY = "cellery"
const ApimRepositoryDeploymentServer = "apim_repository_deployment_server"

const PersistentVolume = "Persistent volume"
const NonPersistentVolume = "Non persistent volume"

const CellerySqlUserName = "cellery-sql-user"
const CellerySqlPassword = "cellery-sql-user"

const IngressModeNodePort = "Node port [kubeadm, minikube]"
const IngressModeLoadBalancer = "Load balancer [gcp, docker for desktop]"

const CelleryDockerCliUserId = "1000"

const TelepresenceExecPath = "telepresence-0.101/bin/"

const IpAddressPattern = "(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])"

const HttpApiIngress = "Http"
const WebIngress = "Web"
const GrpcIngress = "GRPC"
const TcpIngress = "TCP"
const NA = "N/A"
const GatewayHost = "gateway_host"
const PORT = "port"
const HOST = "host"

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
const BalExt = ".bal"

const CentralRegistryHost = "registry.hub.cellery.io"
const CentralRegistryHostRegx = "^.*registry.hub.cellery.(io|net)$"
const CellerySetupBack = "BACK"

const GcpClusterName = "cellery-cluster"
const MysqlHostNameForExistingCluster = "wso2apim-with-analytics-rdbms-service"

const ZipArtifacts = "artifacts"
const ZipMetaSuffix = "_meta"

const CelleryHome = ".cellery"
const GCP = "gcp"
const K8sArtifacts = "k8s-artefacts"
const VM = "vm"
const MySql = "mysql"
const SP = "sp"
const TELEPRESENCE = "telepresence"

const KubeCtl = "kubectl"

const CelleryInstallationPathMac = "/Library/Cellery"
const CelleryInstallationPathUbuntu = "/usr/share/cellery"
const CelleryExecutablePath = "/runtime/executable/"

const BallerinaVersion = "1.0.3"
const BallerinaInstallationPathMac = "/Library/Ballerina/ballerina-" + BallerinaVersion
const BallerinaInstallationPathUbuntu = "/usr/lib/ballerina/ballerina-" + BallerinaVersion
const BallerinaExecutablePath = "/bin/"
const BallerinaConf = "ballerina.conf"
const BallerinaToml = "Ballerina.toml"
const TargetDirName = "target"
const BalTestExecFIle = "test.sh"
const BalInitTestExecFIle = "init-project.sh"

const Wso2ApimHost = "https://wso2-apim-gateway"

const CelleryImageDirEnvVar = "CELLERY_IMAGE_DIR"

const RootDir = "/"
const VAR = "var"
const TMP = "tmp"
const CELLERY = "cellery"
const Ref = "ref"
const ApimRepositoryDeploymentServer = "apim_repository_deployment_server"

const CellerySqlUserName = "cellery-sql-user"
const CellerySqlPassword = "cellery-sql-user"

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

const CELL  = "cell"
const COMPOSITE  = "composite"

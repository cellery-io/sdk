// Copyright (c) 2018 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
import ballerina/config;
import ballerina/io;
import ballerina/log;
import ballerina/stringutils;
import ballerinax/java;


# Cell/Composite image identifier details.
#
# + org - Organization name
# + name - Cell/Composite Image Name
# + ver - Cell Image version
# + instanceName -  Cell/Composite instanceName
# + isRoot - Is this the root cell/composite
public type ImageName record {|
    string org;
    string name;
    string ver;
    string instanceName?;
    boolean isRoot?;
|};

# Labels with default optional fields.
#
# + team - Team name
# + maintainer - Maintainer name
# + owner - Owner name
public type Label record {
    string team?;
    string maintainer?;
    string owner?;
};

# Ingress Expose types
public type Expose "global" | "local";

public type ImageType "Cell" | "Composite";

public type ComponentType "Job" | "Deployment" | "StatefulSet";

# Dockerfile as a component source.
#
# + dockerDir - Path to Dockerfile directory
# + tag - Tag to for generated Docker Image
public type DockerSource record {|
    string dockerDir;
    string tag;
|};

# Docker image as a component source.
#
# + image - Source docker image tag
public type ImageSource record {|
    string image;
|};

# Git repository URL as a component source.
#
# + gitRepo - Git repository URL
# + tag - Tag to for generated Docker Image 
public type GitSource record {|
    string gitRepo;
    string tag;
|};

# File path to the test source file directory.
#
# + filepath - Path to the test source file directory
public type FileSource record {|
    string filepath;
|};

# API resource.
#
# + path - Resource path
# + method - Resource HTTP method
public type ResourceDefinition record {|
    string path;
    string method;
|};

# API Definition as an array of resources.
#
# + resources - Resource definitions
public type ApiDefinition record {|
    ResourceDefinition[] resources;
|};

# ZeroScalingPolicy details.
#
# + maxReplicas - Maximum number of replicas
# + concurrencyTarget - Concurrency Target to scale
public type ZeroScalingPolicy record {|
    int maxReplicas?;
    int concurrencyTarget?;
|};

# Autoscaling policy details.
#
# + overridable - Is autosaling policy overridable
# + minReplicas - Minimum number of replicas
# + maxReplicas - Maximum number of replicas
# + metrics - Scaling metrics
public type AutoScalingPolicy record {|
    boolean overridable = true;
    int minReplicas;
    int maxReplicas;
    Metrics metrics;
|};

# Autoscaling metrics.
#
# + cpu - CPU utilazation
# + memory - Memory utilization
public type Metrics record {|
    Value | Percentage cpu?;
    Value | Percentage memory?;
|};

# Thrshold as a Target Average Value
#
# + threshold - Threshold TargetAverageValue
public type Value record {|
    string threshold;
|};

# Thrshold as a Target Average Utilization
#
# + threshold - Threshold as TargetAverageUtilization
public type Percentage record {|
    int threshold;
|};

# Dependency information of a component.
#
# + components - Depending components
# + cells - Depending cells
# + composites - Depending composites
public type Dependencies record {|
    Component?[] components?;
    map<ImageName | string> cells?;
    map<ImageName | string> composites?;
|};

# Probes configurations.
#
# + initialDelaySeconds - Initial delay seconds before first probe
# + periodSeconds - Period in between two probes
# + timeoutSeconds - Timeout for the probes
# + failureThreshold - No of allowed failures 
# + successThreshold - Required number of successful probes
# + kind - Probe Kind
public type Probe record {|
    int initialDelaySeconds = 0;
    int periodSeconds = 10;
    int timeoutSeconds = 1;
    int failureThreshold = 3;
    int successThreshold = 1;
    TcpSocket | Exec | HttpGet kind;
|};

# TCP Socket probe.
#
# + port - port TCP port
# + host - host TCP host
public type TcpSocket record {|
    int port;
    string host?;
|};

# HttpGet probe.
#
# + path - Reosource path
# + port - Server port
# + httpHeaders - HttpHeader parameters
public type HttpGet record {|
    string path;
    int port;
    map<string> httpHeaders?;
|};

# Command(Exec) based probe.
#
# + commands - commands to be executed.
public type Exec record {|
    string[] commands;
|};

# Component probe configurations.
#
# + readiness - Readiness probe
# + liveness - Liveness probe
public type Probes record {|
    Probe readiness?;
    Probe liveness?;
|};

# Resource quotas.
#
# + requests - requests allowed
# + limits - limits allowed.
public type Resources record {|
    Quota requests?;
    Quota limits?;
|};

# Quota configurataions.
#
# + memory - Memory limts
# + cpu - CPU limits
public type Quota record {|
    string memory?;
    string cpu?;
|};

# Component configurations.
#
# + name - Component name
# + src - Component source
# + replicas - No of replicas
# + ingresses - Component ingresses
# + labels - Labels to be applied to component
# + envVars - Environment variables
# + dependencies - Component dependencies
# + scalingPolicy - scalingPolicy Parameter Description 
# + probes - Component probes
# + resources - Component resource quotas
# + componentType - Component type
# + volumes - Volume mounts for component
public type Component record {|
    string name;
    ImageSource | DockerSource src;
    int replicas = 1;
    map<TCPIngress | HttpApiIngress | GRPCIngress | WebIngress | HttpPortIngress | HttpsPortIngress> ingresses?;
    Label labels?;
    map<Env> envVars?;
    Dependencies dependencies?;
    AutoScalingPolicy | ZeroScalingPolicy scalingPolicy?;
    Probes probes?;
    Resources resources?;
    ComponentType componentType = "Deployment";
    map<VolumeMount> volumes?;
|};

# TCP ingress configurations.
#
# + backendPort - Backend service port
# + gatewayPort - Gateway port to expose
# + ingressTypeTCP - Ingress Type
public type TCPIngress record {|
    int backendPort;
    int gatewayPort?;
    string ingressTypeTCP = "TCP";
|};

# GRPC ingress configurations.
#
# + protoFile - Proto file path
# + ingressType - Ingress Type
public type GRPCIngress record {|
    *TCPIngress;
    string protoFile?;
    string ingressType = "GRPC";
|};

# HTTP API ingress configurations.
#
# + port - API service port
# + context - API context 
# + definition - API definition
# + apiVersion - API version
# + ingressType - Ingress Type
# + expose - Expose API via local or global gateway
# + authenticate - should API be authenticated?
public type HttpApiIngress record {|
    int port;
    string context?;
    ApiDefinition definition?;
    string apiVersion = "0.1";    // Default api version for 0.1
    string ingressType = "Http";
    Expose expose?;
    boolean authenticate = true;
|};

# Web Ingress configurations.
#
# + port - Service port
# + gatewayConfig - Gateway configurations
# + ingressType - Ingress type
public type WebIngress record {|
    int port;
    GatewayConfig gatewayConfig;
    string ingressType = "Web";
|};

# HTTP port Ingress configuration.
#
# + port - HTTP service port
# + ingressType - Ingress type
public type HttpPortIngress record {|
    int port;
    string ingressType = "Http";
|};

# HTTPS port Ingress configuration.
public type HttpsPortIngress record {|
    *HttpPortIngress;
|};

# Gateway configurations.
#
# + vhost - Virtual host
# + context - Gateway context
# + tls - TLS configurations
# + oidc - Open ID Connect configurations
public type GatewayConfig record {|
    string vhost;
    string context = "/";
    TLS tls?;
    OIDC oidc?;
|};

# TLS configurations.
#
# + key - TLS key
# + cert - TLS cert
public type TLS record {|
    string key;
    string cert;
|};

# OpenID connect configurations.
#
# + nonSecurePaths - Non secure paths
# + securePaths - Secure paths
# + providerUrl - OIDC provider URL
# + clientId - OIDC client ID
# + clientSecret - OIDC client secret
# + redirectUrl - Redirect URL
# + baseUrl - Base URL
# + subjectClaim - ODIC subject claim
public type OIDC record {|
    string[] nonSecurePaths = [];
    string[] securePaths = [];
    string providerUrl;
    string clientId;
    string | DCR clientSecret;
    string redirectUrl;
    string baseUrl;
    string subjectClaim?;
|};

# Dynamic client registrations configurations.
#
# + dcrUrl - DCR url
# + dcrUser - DCR username
# + dcrPassword - DCR password
public type DCR record {|
    string dcrUrl?;
    string dcrUser;
    string dcrPassword;
|};

# Environment variable value.
#
# + value - environment variable value
public type ParamValue record {
    string | int | boolean | float value?;
};

# Environment variable.
public type Env record {|
    *ParamValue;
|};

# Gloabal API publisher configurations.
#
# + context - Global API context 
# + apiVersion - Global API version
public type GlobalApiPublisher record {|
    string context?;
    string apiVersion?;
|};

# Cell Image descriptor.
#
# + kind - Image kind
# + globalPublisher - Global publisher configs
# + components - Cell image components 
public type CellImage record {|
    ImageType kind = "Cell";
    GlobalApiPublisher globalPublisher?;
    map<Component> components;
|};

# Composite Image descriptor.
#
# + kind - Image kind
# + components - Composite image components 
public type Composite record {|
    ImageType kind = "Composite";
    map<Component> components;
|};

# Open record to hold cell Reference fields.
public type Reference record {

};

# Test configurations.
#
# + name - Test name
# + src - Test source
# + envVars - Environment variables required for tests
public type Test record {|
    string name;
    ImageSource | FileSource src;
    map<Env> envVars?;
|};

# Test suite configuration.
#
# + tests - Test descriptors
public type TestSuite record {|
    Test?[] tests = [];
|};

# Image instance state holder.
#
# + iName - Image name descriptor
# + isRunning - Is instance running
# + alias - Instance alias
public type InstanceState record {|
    ImageName iName;
    boolean isRunning;
    string alias = "";
|};

# Kubernetes shared persistence configurations.
#
# + name - PVC name
public type K8sSharedPersistence record {|
    string name;
|};

public type Mode "Filesystem" | "Block";

public type AccessMode "ReadWriteOnce" | "ReadOnlyMany" | "ReadWriteMany";

# Selector expression.
#
# + key - Expression key
# + operator - Expression operator
# + values - Expression values
public type Expression record {|
    string key;
    string operator;
    string[] values;
|};

# Node selector lookup configurations.
#
# + labels - Labels parameter  
# + expressions - Expression descriptors 
public type Lookup record {|
    map<string> labels?;
    Expression?[] expressions?;
|};

# Kubernetes non shared persistence volume claim configs.
#
# + name - Volume name 
# + mode - Volume mode 
# + storageClass - Storage class 
# + accessMode - Access Mode
# + lookup - Lookup 
# + request - Request
public type K8sNonSharedPersistence record {|
    string name;
    Mode mode?;
    string storageClass?;
    AccessMode?[] accessMode?;
    Lookup lookup?;
    string request;
|};

# Shared config map descriptor.
#
# + name - Config map name
public type SharedConfiguration record {|
    string name;
|};

# Non shared config map descriptor.
#
# + name - Config map name 
# + data - Config map data
public type NonSharedConfiguration record {|
    string name;
    map<string> data;
|};

# Shared secret.
#
# + name - Secret name
public type SharedSecret record {|
    string name;
|};

# Shared secret.
#
# + name - Secret name 
# + data - Secret data
public type NonSharedSecret record {|
    string name;
    map<string> data;
|};

# Volume mount configurations.
#
# + path - Mount path
# + readOnly - Is volume mount read only?
# + volume - volume configurations
public type VolumeMount record {|
    string path;
    boolean readOnly = false;
    K8sNonSharedPersistence | K8sSharedPersistence | SharedConfiguration | NonSharedConfiguration | SharedSecret
    | NonSharedSecret volume;
|};

public type TestConfig record {|
    ImageName iName;
    map<ImageName> dependencyLinks = {};
    boolean startDependencies = false;
    boolean shareDependencies = false;
|};

# Build the cell artifacts and persist metadata
#
# + image - The cell/composite image definition
# + iName - The cell image org, name & version
# + return - error
public function createImage(CellImage | Composite image, ImageName iName) returns @tainted (error?) {
    //Persist the Ballerina cell image record as a json
    json jsonValue = check json.constructFrom(image.clone());
    string filePath = "./target/cellery/" + iName.name + "_meta.json";
    var wResult = write(jsonValue, filePath);
    if (wResult is error) {
        log:printError("Error occurred while persisiting cell: " + iName.name, err = wResult);
        return wResult;
    }
    // Validate Cell
    validateCell(image);

    //Generate yaml file and other artifacts via extern function
    return createCellImage(image, iName);
}


# Validate Image.
#
# + image - Image descriptor
function validateCell(CellImage | Composite image) {
    image.components.forEach(function (Component component) {
        if (!(component["ingresses"] is ())) {
            map<TCPIngress | HttpApiIngress | GRPCIngress | WebIngress | HttpPortIngress | HttpsPortIngress> ingresses =
            <map<TCPIngress | HttpApiIngress | GRPCIngress | WebIngress | HttpPortIngress | HttpsPortIngress>>
            component?.ingresses;
            if (image.kind == "Composite") {
                ingresses.forEach(function (TCPIngress | HttpApiIngress | GRPCIngress | WebIngress | HttpPortIngress | HttpsPortIngress ingress) {
                    if (ingress is HttpApiIngress || ingress is WebIngress) {
                        string errMsg = "Invalid ingress type in component " + component.name + ". Composites doesn't support HttpApiIngress and WebIngress.";
                        error e = error(errMsg);
                        log:printError("Unspported ingress found ", err = e);
                        panic e;
                    }
                });

            }
        }
        if (!(component["scalingPolicy"] is ()) && (component?.scalingPolicy is AutoScalingPolicy)) {
            AutoScalingPolicy policy = <AutoScalingPolicy>component?.scalingPolicy;
            if ((!(policy?.metrics["cpu"] is ()) && (policy?.metrics?.cpu is Percentage)) &&
            ((component["resources"] is ()) || component?.resources["limits"] is ())) {
                io:println("Warning: cpu percentage is defined without resource limits in component: [" + component.name + "]." +
                " Scaling may not work due to the missing resource limits.");
            }
            if ((!(policy?.metrics["memory"] is ()) && (policy?.metrics?.memory is Percentage))
            && ((component["resources"] is ()) || component?.resources["limits"] is ())) {
                io:println("Warning: memory percentage is defined without resource limits in component [" + component.name + "]." +
                " Scaling may not work due to the missing resource limits.");
            }
        }
    });
}

# Generate Volume Name.
#
# + name - The volume mount name
# + return - Name prefixed with instance name place Holder.
public function generateVolumeName(string name) returns (string) {
    return "{{instance_name}}-" + name;
}

# Construct the image from image descriptor.
#
# + iName - Image name descriptor.
# + return - Return composite image or error
public function constructImage(ImageName iName) returns @tainted (CellImage | Composite) {
    string filePath = config:getAsString("CELLERY_IMAGE_DIR") + "/artifacts/cellery/" + iName.name + "_meta.json";
    var rResult = read(filePath);
    if (rResult is error) {
        log:printError("Error occurred while reading cell image from json: " + iName.name, err = rResult);
        panic rResult;
    }
    json result = <json>rResult;
    string | error kind = string.constructFrom(<json>result.kind);
    if (kind is error) {
        log:printError("Error occurred while getting the image kind from json: " + iName.name, err = kind);
        panic kind;
    }
    CellImage | Composite | error image;
    if (kind == "Cell") {
        image = CellImage.constructFrom(<json>rResult);
    } else {
        image = Composite.constructFrom(<json>rResult);
    }
    if (image is error) {
        log:printError("Error occurred while constructing the image from json: " + iName.name, err = image);
        panic image;
    }
    return <CellImage|Composite>image;
}

# Returns a Reference record with url information.
#
# + iName - The cell instance name
# + return - Reference record
public function resolveReference(ImageName iName) returns (Reference) {
    Reference | error? ref = readReference(iName);
    if (ref is error) {
        log:printError("Error occured while reading reference file ", err = ref);
        panic ref;
    }
    if (ref is ()) {
        error err = error("Empty reference retrieved for " + <string>iName?.instanceName + "\n");
        panic err;
    }
    Reference myRef = <Reference>ref;
    return replaceInRef(myRef);
}
# Returns a Reference record with url information.
#
# + component - Component
# + dependencyAlias - Dependency alias
# + return - Reference record
public function getReference(Component component, string dependencyAlias) returns (Reference) {
    ImageName | string? alias;
    if (!(component?.dependencies["cells"] is ())) {
        alias = component?.dependencies?.cells[dependencyAlias];
    } else {
        alias = component?.dependencies?.composites[dependencyAlias];
    }
    ImageName aliasImage;
    if (alias is string) {
        aliasImage = parseCellDependency(alias);
    } else if (alias is ImageName) {
        aliasImage = alias;
    } else {
        error e = error("Invalid reference error " + dependencyAlias);
        log:printError("Invalid reference found ", err = e);
        panic e;
    }
    aliasImage.instanceName = dependencyAlias;
    Reference | error? ref = readReference(aliasImage);
    if (ref is error) {
        log:printError("Error occured while reading reference file ", err = ref);
        panic ref;
    }
    if (ref is ()) {
        error err = error("Empty reference for dependency `" + dependencyAlias + "`.\n Did you pull/build cell image denoted by alias `" + dependencyAlias + "`? ");
        panic err;
    }
    return <Reference>ref;
}

# Constructs a configuration for cell testing from config API
#
# + return - TestConfig record
public function getTestConfig() returns @tainted TestConfig {
    map<any> configMap  = config:getAsMap("test.config");
    string iNameStr = <string>configMap["IMAGE_NAME"];
    string dependencyLinksStr = <string>configMap["DEPENDENCY_LINKS"];
    boolean startDependencies = <boolean>configMap["START_DEPENDENCIES"];
    boolean shareDependencies = <boolean>configMap["SHARE_DEPENDENCIES"];
    
    // Construct iName from string
    io:StringReader reader = new (iNameStr);
    json | error result = reader.readJson();
    if (result is error) {
        log:printError("Error occured while deriving image name ", err = result);
        panic result;
    }
    ImageName | error iName = ImageName.constructFrom(<json>result);
    if (iName is error) {
        log:printError("Error occured while constructing image name from json", err = iName);
        panic iName;
    }

    // Construct dependencies from string
    reader = new (dependencyLinksStr);
    json | error dependencyJson = reader.readJson();
    if (dependencyJson is error) {
        log:printError("Error occured while deriving dependency links ", err = dependencyJson);
        panic dependencyJson;
    }
    map<ImageName> | error iNameMap = map<ImageName>.constructFrom(<json>dependencyJson);
    if (iNameMap is error) {
        log:printError("Error occured while deriving dependency links ", err = iNameMap);
        panic iNameMap;
    }

    TestConfig testConfig = {
        iName: <ImageName>iName, dependencyLinks: <map<ImageName>>iNameMap,
        startDependencies: startDependencies,shareDependencies: shareDependencies
        };
    return testConfig;
}

# Returns exposed endpoints of a given instance.
#
# + alias - (optional) dependency alias of instance
# + return - reference record
public function getInstanceEndpoints(string alias = "") returns Reference {
    ImageName iName =  {org: "", name: "", ver: ""};
    InstanceState[]|error resultList = getInstanceListFromConfig();
    if (resultList is error) {
        log:printError("Error occurred while getting instance list ", err = resultList);
        panic resultList;
    }
    InstanceState[] instanceList = <InstanceState[]>resultList;
    foreach var inst in instanceList {
            if (inst.alias == alias) {
                iName = <@untainted>inst.iName;
                break;
            }
    }   
    CellImage | Composite | error imageResult = constructImage(iName);
    if (imageResult is error) {
        panic imageResult;
    }
    if(imageResult is CellImage|Composite){
        map<Component> components = imageResult.components;
        foreach var [k, comp] in components.entries() {
            if (comp["dependencies"] is ()) {
                break;
            }
            Reference ref = getReference(comp, alias);
            return replaceInRef(ref, alias = alias, name = <string>iName?.instanceName);
        }
    }
    Reference | error? ref = resolveReference(<ImageName>iName);
    if (ref is error) {
        log:printError("Error occured while resolving reference", err = ref);
        panic ref;
    }
    Reference tempRef = <Reference>ref;
    return tempRef;
}

# Start docker based tests.
#
# + imageName - Image name descriptor
# + envVars - Enviroment variables
# + return - Return error if occured
public function runDockerTest(string imageName, map<Env> envVars) returns (error?) {
    TestConfig testConfig = getTestConfig();
    io:println(testConfig);
    ImageName iName = testConfig.iName;
    string? instanceNameResult = testConfig.iName?.instanceName;
    string instanceName;
    if (instanceNameResult is string && instanceNameResult != "") {
        instanceName = instanceNameResult;
    } else {
        instanceName = "tmp";
    }
    Test dockerTest = {
        name: instanceName + "-test",
        src: {
            image: imageName
        },
        envVars: envVars
    };
    TestSuite testSuite = {
        tests: [dockerTest]
    };

    return runTestSuite(iName, testSuite);
}

# Pared dependecy alias and get image name descriptor
#
# + alias - Dependecy alias
# + return - Return Image name descriptor
function parseCellDependency(string alias) returns ImageName {
    string org = alias.substring(0, <int>alias.indexOf("/"));
    string name = alias.substring(<int>alias.indexOf("/") + 1, <int>alias.indexOf(":"));
    string ver = alias.substring(<int>alias.indexOf(":") + 1, alias.length());
    ImageName imageName = {
        name: name,
        org: org,
        ver: ver
    };
    return imageName;
}

# Returns the hostname of the target component with placeholder for instances name
#
# + component - Target component
# + return - hostname
public function getHost(Component component) returns (string) {
    string host = "{{instance_name}}--" + getValidName(component.name) + "-service";
    return host;
}

# Returns the port number of the target component
#
# + component - Target component
# + return - port number
public function getPort(Component component) returns (int) {
    int port = 0;
    if (component["ingresses"] is ()) {
        error err = error("getPort is invoked on a component: [" + component.name + "] with empty ingress");
        panic err;
    }
    map<TCPIngress | HttpApiIngress | GRPCIngress | WebIngress | HttpPortIngress | HttpsPortIngress> ingresses =
    <map<TCPIngress | HttpApiIngress | GRPCIngress | WebIngress | HttpPortIngress | HttpsPortIngress>>component?.ingresses;
    if (ingresses.length() > 0) {
        var ingress = ingresses[ingresses.keys()[0]];
        if (ingress is TCPIngress) {
            TCPIngress ing = <TCPIngress>ingress;
            port = ing.backendPort;
        } else if (ingress is HttpApiIngress) {
            if (!(component["scalingPolicy"] is ()) && component?.scalingPolicy is ZeroScalingPolicy) {
                port = 80;
            } else {
                HttpApiIngress ing = <HttpApiIngress>ingress;
                port = ing.port;
            }
        } else if (ingress is GRPCIngress) {
            if (!(component["scalingPolicy"] is ()) && component?.scalingPolicy is ZeroScalingPolicy) {
                port = 81;
            } else {
                GRPCIngress ing = <GRPCIngress>ingress;
                port = ing.backendPort;
            }
        } else if (ingress is WebIngress) {
            if (!(component["scalingPolicy"] is ()) && component?.scalingPolicy is ZeroScalingPolicy) {
                port = 80;
            } else {
                WebIngress ing = <WebIngress>ingress;
                port = ing.port;
            }
        }
    }
    return port;
}

# Get a valid kubernetes name.
#
# + name - Name 
# + return - Return valid k8s name
function getValidName(string name) returns string {
    string validName = name.toLowerAscii();
    validName = stringutils:replaceAll(validName, "_", "-");
    return stringutils:replaceAll(validName, "\\.", "-");
}

# Close readable character channel.
#
# + rc - ReadableCharacterChannel
function closeRc(io:ReadableCharacterChannel rc) {
    var result = rc.close();
    if (result is error) {
        log:printError("Error occurred while closing character stream",
        err = result);
    }
}

# Close WritableCharacterChannel.
#
# + wc - WritableCharacterChannel
function closeWc(io:WritableCharacterChannel wc) {
    var result = wc.close();
    if (result is error) {
        log:printError("Error occurred while closing character stream",
        err = result);
    }
}

# Write json to a file.
#
# + content - JSON content
# + path - target file path
# + return - Return error if occured
function write(json content, string path) returns @tainted error? {
    io:WritableByteChannel wbc = check io:openWritableFile(path);
    io:WritableCharacterChannel wch = new (wbc, "UTF8");
    var result = wch.writeJson(content);
    closeWc(wch);
    return result;
}

# Read JSON content from file.
#
# + path - File path.
# + return - Return JSON content or error.
function read(string path) returns @tainted json | error {
    io:ReadableByteChannel rbc = check io:openReadableFile(path);
    io:ReadableCharacterChannel rch = new (rbc, "UTF8");
    var result = rch.readJson();
    closeRc(rch);
    return result;
}

# Replace alias place holder in Reference records.
#
# + ref - Reference value.
# + alias - Alias name
# + name -  Instance name of alias
# + return - Return Value Description
function replaceInRef(Reference ref, string alias = "", string name = "") returns Reference {
    foreach var [key, value] in ref.entries() {
        string temp = <string>value;
        temp = stringutils:replaceAll(temp, "\\{", "");
        temp = stringutils:replaceAll(temp, "\\}", "");
        if (alias != "") {
            temp = stringutils:replace(temp, alias, name);
        }
        ref[key] = temp;
    }
    return ref;
}

# Retrieves the instance list related to cell being tested.
#
# + return - array of InstanceState record
function getInstanceListFromConfig() returns  @tainted (InstanceState[] | error) {
    string instanceListStr = config:getAsString("INSTANCE_LIST");
    // Construct instance list from string
    io:StringReader reader = new (instanceListStr);
    json | error resultJson = reader.readJson();
    if (resultJson is error) {
        log:printError("Error occured while deriving image name ", err = resultJson);
        return resultJson;
    }
    InstanceState[] | error instanceList = InstanceState[].constructFrom(<json>resultJson);
    if (instanceList is error) {
        log:printError("Error occured while constructing image name from json", err = instanceList);
        return instanceList;
    }
    return instanceList;
}

# Build the cell yaml
#
# + image - The cell image definition
# + imageName - The cell image org, name & version
# + return - error
public function createCellImage(CellImage | Composite image, ImageName imageName) returns error? = @java:Method {
    class: "io.cellery.impl.CreateCellImage"
} external;

# Update the cell aritifacts with runtime changes
#
# + image - The cell image definition
# + iName - The cell instance name
# + instances - The cell instance dependencies
# + startDependencies - Whether to start dependencies
# + shareDependencies - Whether to share dependencies
# + return - error optional
public function createInstance(CellImage | Composite image, ImageName iName, map<ImageName> instances,
boolean startDependencies, boolean shareDependencies) returns (InstanceState[] | error?) {
    return trap createInstanceExternal(image, iName, instances, startDependencies, shareDependencies);
}

# Update the cell aritifacts with runtime changes
#
# + image - The cell image definition
# + iName - The cell instance name
# + instances - The cell instance dependencies
# + startDependencies - Whether to start dependencies
# + shareDependencies - Whether to share dependencies
# + return - error optional
public function createInstanceExternal(CellImage | Composite image, ImageName iName, map<ImageName> instances,
boolean startDependencies, boolean shareDependencies) returns (InstanceState[] | error?) = @java:Method {
    class: "io.cellery.impl.CreateInstance"
} external;

# Parse the swagger file and returns API Defintions
#
# + swaggerFilePath - The swaggerFilePath
# + return - Array of ApiDefinitions
public function readSwaggerFile(string swaggerFilePath) returns (ApiDefinition | error) {
    return trap readSwaggerFileExternal(swaggerFilePath);
}

# Parse the swagger file and returns API Defintions
#
# + swaggerFilePath - The swaggerFilePath
# + return - Array of ApiDefinitions
public function readSwaggerFileExternal(string swaggerFilePath) returns (ApiDefinition) = @java:Method {
    class: "io.cellery.impl.ReadSwaggerFile"
} external;


# Returns a Reference record with url information
#
# + iName - Dependency Image Name
# + return - Reference record
public function readReference(ImageName iName) returns (Reference | error?) {
    return trap readReferenceExternal(iName);
}

# Returns a Reference record with url information
#
# + iName - Dependency Image Name
# + return - Reference record
public function readReferenceExternal(ImageName iName) returns (Reference) = @java:Method {
    class: "io.cellery.impl.ReadReference"
} external;

# Run instances required for executing tests
#
# + testConfig - configurations related to cell integration tests
public function runInstances(TestConfig testConfig) {
    InstanceState[] instanceList = [];
    CellImage | Composite | error result = constructImage(<@untainted> testConfig.iName);
    if (result is error) {
        panic result;
    }
    CellImage | Composite instance = <CellImage|Composite>result;
    InstanceState[] | error? resultList = createInstance(<@untainted>instance, testConfig.iName, 
        testConfig.dependencyLinks, testConfig.startDependencies, testConfig.shareDependencies);
    if (resultList is error) {
        if (resultList.detail().toString().endsWith("is already available in the runtime")) {
            log:printInfo(resultList.detail().toString());
            InstanceState iNameState = {
                iName : testConfig.iName, 
                isRunning: true
            };
            instanceList[instanceList.length()] = <@untainted> iNameState;
        } else {
            panic resultList;
        }
    } else {
        instanceList = <InstanceState[]>resultList;
    }
    foreach var inst in <InstanceState[]>instanceList {
        if (inst.alias == "") {
            json|error iNameJson = json.constructFrom(inst.iName);
            if (iNameJson is error) {
                panic iNameJson;
            } else {
                config:setConfig("test.config.IMAGE_NAME", iNameJson.toJsonString());
                io:println(getTestConfig());
            }
            break;
        }
    }
    json|error instanceListJson = json.constructFrom(<InstanceState[]>instanceList);
    if (instanceListJson is error) {
        panic instanceListJson;
    } else {
        config:setConfig("INSTANCE_LIST", instanceListJson.toJsonString());
    }
}

# Start test suite.
#
# + iName - Image name descriptor
# + testSuite - Test suite configs
# + return - Return error if occured
public function runTestSuite(ImageName iName, TestSuite testSuite) returns (error?) = @java:Method {
    class: "io.cellery.impl.RunTestSuite"
} external;

# Terminate instances started for testing.
#
# + return - error optional
public function stopInstances() returns @tainted (error?) {
    InstanceState[]|error instanceList = getInstanceListFromConfig();
    if (instanceList is error) {
        log:printError("Error occurred while terminating instances ", err = instanceList);
        return instanceList;
    }
    return stopInstancesExternal(<InstanceState[]>instanceList);
}

# Terminate instances started for testing.
#
# + instances -  The cell instance dependencies
# + return - error optional
public function stopInstancesExternal(InstanceState[] instances) returns (error?) = @java:Method {
    class: "io.cellery.impl.StopInstances"
} external;

# Create a Persistence Claim.
#
# + pvc -  The K8sNonSharedPersistence record
# + return - error optional
public function createPersistenceClaim(K8sNonSharedPersistence pvc) returns (error?) = @java:Method {
    class: "io.cellery.impl.CreatePersistenceClaim"
} external;

# Create a Secret.
#
# + secret -  The NonSharedSecret record
# + return - error optional
public function createSecret(NonSharedSecret secret) returns (error?) = @java:Method {
    class: "io.cellery.impl.CreateSecret"
} external;

# Create a ConfigMap.
#
# + configuration -  The NonSharedConfiguration record
# + return - error optional
public function createConfiguration(NonSharedConfiguration configuration) returns (error?) = @java:Method {
    class: "io.cellery.impl.CreateConfiguration"
} external;

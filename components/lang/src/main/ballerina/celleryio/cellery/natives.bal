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
import ballerina/log;
import ballerina/io;
import ballerina/config;

public type ImageName record {|
    string org;
    string name;
    string ver;
    string instanceName?;
    boolean isRoot?;
|};

public type Label record {
    string team?;
    string maintainer?;
    string owner?;
};

# Ingress Expose types
public type Expose "global" | "local";

public type ImageType "Cell" | "Composite";

public type ComponentType "Job" | "Deployment" | "StatefulSet";

public type DockerSource record {|
    string dockerDir;
    string tag;
|};

public type ImageSource record {|
    string image;
|};

public type GitSource record {|
    string gitRepo;
    string tag;
|};

public type FileSource record {|
    string filepath;
|};

public type ResourceDefinition record {|
    string path;
    string method;
|};

public type ApiDefinition record {|
    ResourceDefinition[] resources;
|};

public type ZeroScalingPolicy record {|
    int maxReplicas?;
    int concurrencyTarget?;
|};

public type AutoScalingPolicy record {|
    boolean overridable = true;
    int minReplicas;
    int maxReplicas;
    Metrics metrics;
|};

public type Metrics record {|
    Value | Percentage cpu?;
    Value | Percentage memory?;
|};

public type Value record {|
    string threshold;
|};

public type Percentage record {|
    int threshold;
|};

public type Dependencies record {|
    Component?[] components?;
    map<ImageName | string> cells?;
    map<ImageName | string> composites?;
|};

public type Probe record {|
    int initialDelaySeconds = 0;
    int periodSeconds = 10;
    int timeoutSeconds = 1;
    int failureThreshold = 3;
    int successThreshold = 1;
    TcpSocket | Exec | HttpGet kind;
|};

public type TcpSocket record {|
    int port;
    string host?;
|};

public type HttpGet record {|
    string path;
    int port;
    map<string> httpHeaders?;
|};

public type Exec record {|
    string[] commands;
|};

public type Probes record {|
    Probe readiness?;
    Probe liveness?;
|};

public type Resources record {|
    Quota requests?;
    Quota limits?;
|};

public type Quota record {|
    string memory?;
    string cpu?;
|};

public type Component record {|
    string name;
    ImageSource | DockerSource source;
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

public type TCPIngress record {|
    int backendPort;
    int gatewayPort?;
|};

public type GRPCIngress record {|
    *TCPIngress;
    string protoFile?;
|};

public type HttpApiIngress record {|
    int port;
    string context?;
    ApiDefinition definition?;
    string apiVersion?;    // version is a keyword in Ballerina.
    Expose expose?;
    boolean authenticate = true;
|};

public type WebIngress record {|
    int port;
    GatewayConfig gatewayConfig;
|};

public type HttpPortIngress record {|
    int port;
|};

public type HttpsPortIngress record {|
    *HttpPortIngress;
|};

public type GatewayConfig record {|
    string vhost;
    string context = "/";
    TLS tls?;
    OIDC oidc?;
|};

public type URI record {|
    string vhost;
    string context = "/";
|};

public type TLS record {|
    string key;
    string cert;
|};

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

public type DCR record {|
    string dcrUrl?;
    string dcrUser;
    string dcrPassword;
|};

public type ParamValue record {
    string | int | boolean | float value?;
};

public type Env record {|
    *ParamValue;
|};

public type GlobalApiPublisher record {|
    string context?;
    string apiVersion?;
|};

public type CellImage record {|
    ImageType kind = "Cell";
    GlobalApiPublisher globalPublisher?;
    map<Component> components;
|};

public type Composite record {|
    ImageType kind = "Composite";
    map<Component> components;
|};

# Open record to hold cell Reference fields.
public type Reference record {

};

public type Test record {|
    string name;
    ImageSource | FileSource source;
    map<Env> envVars?;
|};

public type TestSuite record {|
    Test?[] tests = [];
|};

public type InstanceState record {|
    ImageName iName;
    boolean isRunning;
    string alias = "";
|};

public type K8sSharedPersistence record {|
    string name;
|};

public type Mode "Filesystem" | "Block";

public type StorageClass "null" | "" | "slow" | "fast";

public type AccessMode "ReadWriteOnce" | "ReadOnlyMany" | "ReadWriteMany";

public type Expression record {|
    string key;
    string operator;
    string[] values;
|};

public type Lookup record {|
    map<string> labels?;
    Expression?[] expressions?;
|};

public type K8sNonSharedPersistence record {|
    string name;
    Mode mode?;
    StorageClass storageClass?;
    AccessMode?[] accessMode?;
    Lookup lookup?;
    string request;
|};

public type SharedConfiguration record {|
    string name;
|};

public type NonSharedConfiguration record {|
    string name;
    map<string> data;
|};

public type SharedSecret record {|
    string name;
|};
public type NonSharedSecret record {|
    string name;
    map<string> data;
|};

public type VolumeMount record {|
    string path;
    boolean readOnly = false;
    K8sNonSharedPersistence | K8sSharedPersistence | SharedConfiguration | NonSharedConfiguration | SharedSecret
    | NonSharedSecret volume;
|};

# Build the cell artifacts and persist metadata
#
# + image - The cell/composite image definition
# + iName - The cell image org, name & version
# + return - error
public function createImage(CellImage | Composite image, ImageName iName) returns ( error?) {
    //Persist the Ballerina cell image record as a json
    json jsonValue = check json.stamp(image.clone());
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


function validateCell(CellImage image) {
    foreach var(key,component) in image.components {
        if (!(component["ingresses"] is ()) && component.ingresses.length() > 1) {
            error err = error("component: [" + component.name + "] has more than one ingress");
            panic err;
        } else if (image.kind == "Composite") {
            //TODO: Fix this when multiple ingress support is added.
            var ingress = component.ingresses[component.ingresses.keys()[0]];
            if (ingress is HttpApiIngress || ingress is WebIngress) {
                string errMsg = "Invalid ingress type in component " + component.name + ". Composites doesn't support HttpApiIngress and WebIngress.";
                error e = error(errMsg);
                log:printError("Invalid ingress found ", err = e);
                panic e;
            }
        }
        if (!(component["scalingPolicy"] is ()) && (component.scalingPolicy is AutoScalingPolicy)) {
            AutoScalingPolicy policy = <AutoScalingPolicy> component.scalingPolicy;
            if ((!(policy.metrics["cpu"] is ()) && (policy.metrics.cpu is Percentage)) &&
            ((component["resources"] is ()) || component.resources["limits"] is ())) {
                io:println("Warning: cpu percentage is defined without resource limits in component: [" + component.name + "]." +
                " Scaling may not work due to the missing resource limits.");
            }
            if ((!(policy.metrics["memory"] is ()) && (policy.metrics.memory is Percentage))
            && ((component["resources"] is ()) || component.resources["limits"] is ())) {
                io:println("Warning: memory percentage is defined without resource limits in component [" + component.name + "]." +
                " Scaling may not work due to the missing resource limits.");
            }
        }
    }
}

# Generate Volume Name.
#
# + name - The volume mount name
# + return - Name prefixed with instance name place Holder.
public function generateVolumeName(string name) returns (string) {
    return "{{instance_name}}-" + name;
}


# Build the cell yaml
#
# + image - The cell image definition
# + iName - The cell image org, name & version
# + return - error
public function createCellImage(CellImage image, ImageName iName) returns ( error?) = external;

# Update the cell aritifacts with runtime changes
#
# + image - The cell image definition
# + iName - The cell instance name
# + instances - The cell instance dependencies
# + startDependencies - Whether to start dependencies
# + return - error optional
public function createInstance(CellImage | Composite image, ImageName iName, map<ImageName> instances,
boolean startDependencies, boolean shareDependencies) returns (InstanceState[] | error?) = external;

# Update the cell aritifacts with runtime changes
#
# + iName - The cell instance name
# + return - error or CellImage record
public function constructCellImage(ImageName iName) returns (CellImage | error) {
    string filePath = config:getAsString("CELLERY_IMAGE_DIR") + "/artifacts/cellery/" + iName.name + "_meta.json";
    var rResult = read(filePath);
    if (rResult is error) {
        log:printError("Error occurred while constructing reading cell image from json: " + iName.name, err = rResult);
        return rResult;
    }
    CellImage | error image = CellImage.stamp(rResult);
    return image;
}

public function constructImage(ImageName iName) returns (Composite | error) {
    string filePath = config:getAsString("CELLERY_IMAGE_DIR") + "/artifacts/cellery/" + iName.name + "_meta.json";
    var rResult = read(filePath);
    if (rResult is error) {
        log:printError("Error occurred while constructing reading cell image from json: " + iName.name, err = rResult);
        return rResult;
    }
    Composite | error image = Composite.stamp(rResult);
    return image;
}

# Parse the swagger file and returns API Defintions
#
# + swaggerFilePath - The swaggerFilePath
# + return - Array of ApiDefinitions
public function readSwaggerFile(string swaggerFilePath) returns (ApiDefinition | error) = external;

# Returns a Reference record with url information
#
# + iName - Dependency Image Name
# + return - Reference record
public function readReference(ImageName iName) returns (Reference | error | ()) = external;

# Returns a Reference record with url information
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
        error err = error("Empty reference retrieved for " + iName.instanceName + "\n");
        panic err;
    }
    Reference myRef = <Reference>ref;
    return replaceInRef(myRef);
}
# Returns a Reference record with url information
#
# + component - Component
# + dependencyAlias - Dependency alias
# + return - Reference record
public function getReference(Component component, string dependencyAlias) returns (Reference) {
    ImageName | string? alias;
    if (!(component.dependencies["cells"] is ())) {
        alias = component.dependencies.cells[dependencyAlias];
    } else {
        alias = component.dependencies.composites[dependencyAlias];
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
        error err = error("Empty reference for dependency `" + dependencyAlias + "`.\n
        Did you pull/build cell image denoted by alias `" + dependencyAlias + "`? ");
        panic err;
    }
    return <Reference>ref;
}

# Run instances required for executing tests
#
# + iName - Cell instance name to start before executing tests
# + instances - The cell instance dependencies
# + return - error optional
public function runInstances(ImageName iName, map<ImageName> instances) returns ImageName[] = external;

public function runTestSuite(InstanceState[] instances, TestSuite testSuite) returns ( error?) = external;

# Terminate instances started for testing.
#
# + instances -  The cell instance dependencies
# + return - error optional
public function stopInstances(InstanceState[] instances) returns ( error?) = external;

# Returns the Image Name of the cell
#
# + return - ImageName
public function getCellImage() returns (ImageName | error){
    string iNameStr = config:getAsString("IMAGE_NAME",
    defaultValue = "{org:\"\", name:\"\", ver:\"\", instanceName:\"\"}");
    io:StringReader reader = new(iNameStr);
    json|error iNameJson = reader.readJson();
    if (iNameJson is error) {
        return iNameJson;
    }
    ImageName|error iName = ImageName.convert(iNameJson);
    return iName;
}

public function createPersistenceClaim(K8sNonSharedPersistence pvc) returns ( error?) = external;

public function createSecret(NonSharedSecret secret) returns ( error?) = external;

public function createConfiguration(NonSharedConfiguration configuration) returns ( error?) = external;

# Get cell dependencies map
#
# + return - map of dependencies ImageName
public function getDependencies() returns (map<ImageName> | error) {
    string dependencyStr = config:getAsString("DEPENDENCY_LINKS", defaultValue = "{}");
    io:StringReader reader = new(dependencyStr);
    json|error dependencyJson = reader.readJson();
    if (dependencyJson is error) {
        return dependencyJson;
    }
    map<ImageName>|error instances = map<ImageName>.convert(dependencyJson);
    return instances;
}
# Returns cell gateway URL of the started cell
#
# + iNameList - list of InstanceState
# + alias - (optional) dependency alias of instance
# + kind - Composite/Cell and defaults
# + return - URL of the cell gateway
public function getGatewayHost(InstanceState[] iNameList, string alias = "", string kind = "Cell") returns (Reference|error) {
    ImageName iName = {org: "", name:"", ver:""};
    foreach var inst in iNameList {
        if (inst.alias == "") {
            iName = inst.iName;
            break;
        }
    }
    foreach var instState in iNameList {
        if (instState.alias != "" && instState.alias == alias) {
            if (kind == "Cell") {
                CellImage cellImage = <CellImage>constructCellImage(iName);
                foreach var(k, comp) in cellImage.components {
                    if (comp["dependencies"] is ()) {
                        break;
                    }
                    Reference ref = getReference(comp, instState.alias);

                    Reference myref = ref;
                    return replaceInRef(myref, alias = instState.alias, name = instState.iName.instanceName);
                }
            } else {
                Composite composite = <Composite>constructCellImage(iName);
                foreach var(k, comp) in composite.components {
                    if (comp["dependencies"] is ()) {
                        break;
                    }
                    Reference ref = getReference(comp, instState.alias);

                    Reference myref = ref;
                    return replaceInRef(myref, alias = instState.alias, name = instState.iName.instanceName);
                }
            }
        }
    }
    Reference | error? ref = resolveReference(<ImageName>iName);
    Reference tempRef = <Reference> ref;
    return tempRef;
}

function parseCellDependency(string alias) returns ImageName {
    string org = alias.substring(0, alias.indexOf("/"));
    string name = alias.substring(alias.indexOf("/") + 1, alias.indexOf(":"));
    string ver = alias.substring(alias.indexOf(":") + 1, alias.length());
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
    if (!(component["scalingPolicy"] is ()) && component.scalingPolicy is ZeroScalingPolicy) {
        host += "-rev";
    }
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
    if (component.ingresses.length() > 0) {
        var ingress = component.ingresses[component.ingresses.keys()[0]];
        if (ingress is TCPIngress) {
            TCPIngress ing = <TCPIngress>ingress;
            port = ing.backendPort;
        } else if (ingress is HttpApiIngress) {
            if (!(component["scalingPolicy"] is ()) && component.scalingPolicy is ZeroScalingPolicy) {
                port = 80;
            } else {
                HttpApiIngress ing = <HttpApiIngress>ingress;
                port = ing.port;
            }
        } else if (ingress is GRPCIngress) {
            if (!(component["scalingPolicy"] is ()) && component.scalingPolicy is ZeroScalingPolicy) {
                port = 81;
            } else {
                GRPCIngress ing = <GRPCIngress>ingress;
                port = ing.backendPort;
            }
        } else if (ingress is WebIngress) {
            if (!(component["scalingPolicy"] is ()) && component.scalingPolicy is ZeroScalingPolicy) {
                port = 80;
            } else {
                WebIngress ing = <WebIngress>ingress;
                port = ing.port;
            }
        }
    }
    return port;
}

function getValidName(string name) returns string {
    return name.toLower().replace("_", "-").replace(".", "-");
}

function closeRc(io:ReadableCharacterChannel rc) {
    var result = rc.close();
    if (result is error) {
        log:printError("Error occurred while closing character stream",
            err = result);
    }
}

function closeWc(io:WritableCharacterChannel wc) {
    var result = wc.close();
    if (result is error) {
        log:printError("Error occurred while closing character stream",
            err = result);
    }
}

function write(json content, string path) returns error? {
    io:WritableByteChannel wbc = io:openWritableFile(path);
    io:WritableCharacterChannel wch = new(wbc, "UTF8");
    var result = wch.writeJson(content);
    if (result is error) {
        closeWc(wch);
        return result;
    } else {
        closeWc(wch);
        return result;
    }
}

function read(string path) returns json | error {
    io:ReadableByteChannel rbc = io:openReadableFile(path);
    io:ReadableCharacterChannel rch = new(rbc, "UTF8");
    var result = rch.readJson();
    if (result is error) {
        closeRc(rch);
        return result;
    } else {
        closeRc(rch);
        return result;
    }
}

function replaceInRef (Reference ref, string alias = "", string name = "") returns Reference {
    foreach var(key,value) in ref {
        string temp = <string> value;
        temp = temp.replaceAll("\\{", "");
        temp = temp.replaceAll("\\}", "");
        if (alias != "") {
            temp = temp.replace(alias, name);
        }
        ref[key] = temp;
    }
    return ref;
}

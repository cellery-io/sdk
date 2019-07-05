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
|};

public type Label record {
    string team?;
    string maintainer?;
    string owner?;
};

# Ingress Expose types
public type Expose "global" | "local";

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
    map<TCPIngress | HttpApiIngress | GRPCIngress | WebIngress> ingresses?;
    Label labels?;
    map<Env> envVars?;
    Dependencies dependencies?;
    AutoScalingPolicy | ZeroScalingPolicy scalingPolicy?;
    Probes probes?;
    Resources resources?;
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
    Expose expose?;
    boolean authenticate = true;
|};

public type WebIngress record {|
    int port;
    GatewayConfig gatewayConfig;
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

public type Secret record {|
    *ParamValue;
    string mountPath;
    boolean readOnly;
|};

public type CellImage record {|
    map<Component> components;
|};

# Open record to hold cell Reference fields.
public type Reference record {

};

public type Test record {|
    string name;
    ImageSource source;
    map<Env> envVars?;
|};

public type TestSuite record {|
	Test[] tests = [];
|};

# Build the cell artifacts and persist metadata
#
# + cellImage - The cell image definition
# + iName - The cell image org, name & version
# + return - error
public function createImage(CellImage cellImage, ImageName iName) returns ( error?) {
    //Persist the Ballerina cell image record as a json
    json jsonValue = check json.stamp(cellImage.clone());
    string filePath = "./target/cellery/" + iName.name + "_meta.json";
    var wResult = write(jsonValue, filePath);
    if (wResult is error) {
        log:printError("Error occurred while persisiting cell: " + iName.name, err = wResult);
        return wResult;
    }
    validateScaling(cellImage);
    //Generate yaml file and other artifacts via extern function
    return createCellImage(cellImage, iName);
}


function validateScaling(CellImage cellImage) {
    foreach var(key,component) in cellImage.components {
        if (!(component["scalingPolicy"] is ()) && (component.scalingPolicy is AutoScalingPolicy)) {
            AutoScalingPolicy policy = <AutoScalingPolicy> component.scalingPolicy;
            if ((!(policy.metrics["cpu"] is ()) && (policy.metrics.cpu is Percentage)) &&
            ((component["resources"] is ())|| component.resources["limits"] is ())) {
                io:println("Warning: cpu percentage is defined without resource limits in component: [" + component.name+ "]."+
                " Scaling may not work due to the missing resource limits.");
            }
            if ((!(policy.metrics["memory"] is ()) && (policy.metrics.memory is Percentage))
            && ((component["resources"] is ()) || component.resources["limits"] is ())) {
                io:println("Warning: memory percentage is defined without resource limits in component [" + component.name+ "]."+
                " Scaling may not work due to the missing resource limits.");
            }
        }
    }
}


# Build the cell yaml
#
# + cellImage - The cell image definition
# + iName - The cell image org, name & version
# + return - error
public function createCellImage(CellImage cellImage, ImageName iName) returns ( error?) = external;

# Update the cell aritifacts with runtime changes
#
# + cellImage - The cell image definition
# + iName - The cell instance name
# + instances - The cell instance dependencies
# + return - error optional
public function createInstance(CellImage cellImage, ImageName iName, map<ImageName> instances) returns ( error?) = external;

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
    CellImage | error cellImage = CellImage.stamp(rResult);
    return cellImage;
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
    foreach var(key,value) in myRef {
        string temp = <string> value;
        temp = temp.replaceAll("\\{", "");
        temp = temp.replaceAll("\\}", "");
        myRef[key]=temp;
    }
    return myRef;
}
# Returns a Reference record with url information
#
# + component - Component
# + dependencyAlias - Dependency alias
# + return - Reference record
public function getReference(Component component, string dependencyAlias) returns (Reference) {
    var alias = component.dependencies.cells[dependencyAlias];
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

# Run tests for cell instance
#
# + iName - The cell instance name to test
# + testSuite - The testsuite to run
# + return - error optional
public function runTestSuite(ImageName iName, TestSuite testSuite) returns (error?) = external;

# Description
#
# + iName - Cell instance name to stop after executing tests
# + instances -  The cell instance dependencies 
# + return - error optional
public function stopInstances(ImageName iName, ImageName[] instances) returns (error?) = external;

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

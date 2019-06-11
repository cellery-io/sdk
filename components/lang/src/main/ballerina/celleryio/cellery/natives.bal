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

public type ImageName record {
    string org;
    string name;
    string ver;
    string instanceName?;
    !...;
};

public type Label record {
    string team?;
    string maintainer?;
    string owner?;
};

# Ingress Expose types
public type Expose "global"|"local";

public type DockerSource record {
    string dockerDir;
    string tag;
    !...;
};

public type ImageSource record {
    string image;
    !...;
};

public type GitSource record {
    string gitRepo;
    string tag;
    !...;
};

public type ResourceDefinition record {
    string path;
    string method;
    !...;
};

public type ApiDefinition record {
    ResourceDefinition[] resources;
    !...;
};

public type AutoScaling record {
    AutoScalingPolicy policy;
    boolean overridable = true;
    !...;
};

public type AutoScalingPolicy record {
    int minReplicas;
    int maxReplicas;
    CpuUtilizationPercentage cpuPercentage;
    !...;
};

public type CpuUtilizationPercentage record {
    int percentage;
    !...;
};

public type Dependencies record {
    map<ImageName|string> components?;
    map<ImageName|string> cells?;
    !...;
};

public type Component record {
    string name;
    ImageSource|DockerSource source;
    int replicas = 1;
    map<TCPIngress|HttpApiIngress|GRPCIngress|WebIngress> ingresses?;
    Label labels?;
    map<Env> envVars?;
    Dependencies dependencies?;
    AutoScaling autoscaling?;
    !...;
};

public type TCPIngress record {
    int backendPort;
    int gatewayPort;
    !...;
};

public type GRPCIngress record {
    *TCPIngress;
    string protoFile?;
    !...;
};

public type HttpApiIngress record {
    int port;
    string context?;
    ApiDefinition definition?;
    Expose expose?;
    boolean authenticate = true;
    !...;
};

public type WebIngress record {
    int port;
    GatewayConfig gatewayConfig;
    !...;
};

public type GatewayConfig record {
    string vhost;
    string context = "/";
    TLS tls?;
    OIDC oidc?;
    !...;
};

public type URI record {
    string vhost;
    string context = "/";
    !...;
};

public type TLS record {
    string key;
    string cert;
    !...;
};

public type OIDC record {
    string[] nonSecurePaths = [];
    string[] securePaths = [];
    string providerUrl;
    string clientId;
    string|DCR clientSecret;
    string redirectUrl;
    string baseUrl;
    string subjectClaim?;
    !...;
};

public type DCR record {
    string dcrUrl?;
    string dcrUser;
    string dcrPassword;
    !...;
};

public type ParamValue record {
    string|int|boolean|float value?;
};

public type Env record {
    *ParamValue;
    !...;
};

public type Secret record {
    *ParamValue;
    string mountPath;
    boolean readOnly;
    !...;
};

public type CellImage record {
    map<Component> components;
    !...;
};

# Open record to hold cell Reference fields.
public type Reference record {

};

# Build the cell aritifacts and persist metadata
#
# + cellImage - The cell image definition
# + iName - The cell image org, name & version
# + return - error
public function createImage(CellImage cellImage, ImageName iName) returns (error?) {
    //Persist the Ballerina cell image record as a json
    json jsonValue = check json.stamp(cellImage.clone());
    string filePath = "./target/cellery/" + iName.name + "_meta.json";
    var wResult = write(jsonValue, filePath);
    if (wResult is error) {
        log:printError("Error occurred while persisiting cell: " + iName.name, err = wResult);
        return wResult;
    }
    //Generate yaml file and other artifacts via extern function
    return createCellImage(cellImage, iName);
}

# Build the cell yaml
#
# + cellImage - The cell image definition
# + iName - The cell image org, name & version
# + return - error
public extern function createCellImage(CellImage cellImage, ImageName iName) returns (error?);

# Update the cell aritifacts with runtime changes
#
# + cellImage - The cell image definition
# + iName - The cell instance name
# + instances - The cell instance dependencies
# + return - error optinal
public extern function createInstance(CellImage cellImage, ImageName iName, map<ImageName> instances) returns (error?);


# Update the cell aritifacts with runtime changes
#
# + iName - The cell instance name
# + return - error or CellImage record
public function constructCellImage(ImageName iName) returns (CellImage|error) {
    string filePath = config:getAsString("CELLERY_IMAGE_DIR") + "/artifacts/cellery/" + iName.name + "_meta.json";
    var rResult = read(filePath);
    if (rResult is error) {
        log:printError("Error occurred while constructing reading cell image from json: " + iName.name, err = rResult);
        return rResult;
    }
    CellImage|error cellImage = CellImage.stamp(rResult);
    return cellImage;
}

# Parse the swagger file and returns API Defintions
#
# + swaggerFilePath - The swaggerFilePath
# + return - Array of ApiDefinitions
public extern function readSwaggerFile(string swaggerFilePath) returns (ApiDefinition|error);

# Returns a Reference record with url information
#
# + iName - Dependency Image Name
# + return - Reference record
public extern function readReference(ImageName iName) returns (Reference|error);

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
    } else if (alias is ImageName){
        aliasImage = alias;
    } else {
        error e = error("Invalid reference error " + dependencyAlias);
        log:printError("Invalid reference found ", err = e);
        panic e;
    }
    aliasImage.instanceName = dependencyAlias;
    Reference|error ref = readReference(aliasImage);
    if (ref is error) {
        log:printError("Error occured while reading reference file", err = ref);
        panic ref;
    }
    return <Reference>ref;
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
    return "{{instance_name}}--" + getValidName(component.name) + "-service";
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

function read(string path) returns json|error {
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

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

public type StructuredName record{
    string orgName;
    string imageName;
    string imageVersion;
    string instanceName?;
    !...;
};

# Pre-defined labels for cellery components.
public type Label record {
    string team?;
    string maintainer?;
    string owner?;
};

# Ingress Expose types
public type Expose "global"|"local";

public type DockerSource record {
    string Dockerfile;
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

public type ApiDefinition record {
    string path;
    string method;
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

public type Component record {
    string name;
    ImageSource source;
    int replicas = 1;
    map<TCPIngress|HttpApiIngress|GRPCIngress|WebIngress> ingresses?;
    Label labels?;
    map<Env> envVars?;
    map<StructuredName> dependencies?;
    AutoScaling autoscaling?;
    !...;
};

public type TCPIngress record {
    int backendPort;
    int gatewayPort;
    Expose expose;
    !...;
};

public type GRPCIngress record {
    *TCPIngress;
    string protoFile?;
    !...;
};

public type HttpApiIngress record {
    int port;
    string context;
    ApiDefinition[] definitions;
    Expose expose;
    boolean authenticate = true;
    !...;
};

public type WebIngress record {
    int port;
    URI uri;
    TLS tls?;
    OIDC oidc?;
    !...;
};

public type URI record{
    string vhost;
    string context = "/";
    !...;
};

public type TLS record{
    string key;
    string cert;
    !...;
};

# OpenId Connect properties
public type OIDC record {
    string[] context;
    string provider;
    string clientId;
    string clientSecret;
    string redirectUrl;
    string baseUrl;
    string subjectClaim;
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
    !...;
};

public type CellImage record {
    Component[] components;
};

# Build the cell aritifacts
#
# + cellImage - The cell image definition
# + sname - The cell image org, name & version
# + return - error
public extern function createImage(CellImage cellImage, StructuredName sName) returns (error?);

# Update the cell aritifacts with runtime changes
#
# + cellImage - The cell image definition
# + sName - The cell instance name
# + return - true/false
public extern function createInstance(CellImage cellImage, StructuredName sName) returns (error?);

# Parse the swagger file and returns API Defintions
#
# + swaggerFilePath - The swaggerFilePath
# + return - Array of ApiDefinitions
public extern function readSwaggerFile(string swaggerFilePath) returns (ApiDefinition[]|error);

public function getHost(string cellImageName, Component component) returns (string) {
    return cellImageName + "--" + getValidName(component.name) + "-service";
}

function getValidName(string name) returns string {
    return name.toLower().replace("_", "-").replace(".", "-");
}

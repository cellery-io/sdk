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

public type ImageName record{
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

public type ResourceDefinition record{
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

public type Component record {
    string name;
    ImageSource source;
    int replicas = 1;
    map<TCPIngress|HttpApiIngress|GRPCIngress|WebIngress> ingresses?;
    Label labels?;
    map<Env> envVars?;
    map<ImageName> dependencies?;
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
    ApiDefinition? definition;
    Expose expose;
    boolean authenticate = true;
    !...;
};

public type WebIngress record {
    int port;
    GatewayConfig gatewayConfig;
    !...;
};

public type GatewayConfig record{
    string vhost;
    string context = "/";
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

public type OIDC record {
    string[] nonSecurePaths = [];
    string[] securePaths = [];
    string discoveryUrl;
    string clientId;
    string clientSecret?;
    DCR dcr?;
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
public type Reference record{

};

# Build the cell aritifacts
#
# + cellImage - The cell image definition
# + iName - The cell image org, name & version
# + return - error
public extern function createImage(CellImage cellImage, ImageName iName) returns (error?);

# Update the cell aritifacts with runtime changes
#
# + cellImage - The cell image definition
# + iName - The cell instance name
# + return - true/false
public extern function createInstance(CellImage cellImage, ImageName iName) returns (error?);

# Parse the swagger file and returns API Defintions
#
# + swaggerFilePath - The swaggerFilePath
# + return - Array of ApiDefinitions
public extern function readSwaggerFile(string swaggerFilePath) returns (ApiDefinition|error);

# Returns a Reference record with url information
#
# + iName - Dependency Image Name
# + return - Reference record
public extern function getReferenceRecord(ImageName iName) returns (Reference|error);

public function getHost(string cellImageName, Component component) returns (string) {
    return cellImageName + "--" + getValidName(component.name) + "-service";
}

function getValidName(string name) returns string {
    return name.toLower().replace("_", "-").replace(".", "-");
}

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

public const string TEAM = "TEAM";
public const string OWNER = "OWNER";
public const string MAINTAINER = "MAINTAINER";

# Pre-defined labels for cellery components.
public type Label "TEAM"|"OWNER"|"MAINTAINER";

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

public type API record {
    string name?;
    string targetComponent;
    boolean global;
    HttpApiIngress ingress;
    !...;
};

public type TCP record{
    string name?;
    string targetComponent;
    TCPIngress ingress;
};

public type Resiliency record {
    RetryConfig retryConfig?;
    FailoverConfig failoverConfig?;
    !...;
};

public type RetryConfig record {
    int interval;
    int count;
    float backOffFactor;
    int maxWaitInterval;
    !...;
};

public type FailoverConfig record {
    int timeOut;
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

public type CpuUtilizationPercentage object {
    public int percentage;
    public function __init(int percentage) {
        self.percentage = percentage;
    }
};

public type Component record {
    string name;
    ImageSource source;
    int replicas = 1;
    map<TCPIngress|HttpApiIngress> ingresses?;
    map<string> labels?;
    map<ParamValue> parameters?;
    AutoScaling autoscaling?;
    !...;
};

public type TCPIngress object {
    public int port;
    public int targetPort;
    public function __init(int port, int targetPort) {
        self.port = port;
        self.targetPort = targetPort;
    }
};

public type HttpApiIngress object {
    public int port;
    public string context;
    public ApiDefinition[] definitions;

    public function __init(int port, string context, ApiDefinition[] definitions) {
        self.port = port;
        self.context = context;
        self.definitions = definitions;
    }
};

public type ParamValue abstract object {
    public string|int|boolean|float? value;
};

public type Env object {
    *ParamValue;
    public function __init(string|int|boolean|float? default = ()) {
        self.value = default;
    }
};

public type Secret object {
    *ParamValue;
    public string path;

    public function __init() {
        self.path = "";
        self.value = "";
    }
};

public type CellImage object {
    public map<Component> components = {};
    public map<API> apis = {};
    public map<TCP> tcp = {};

    public function addComponent(Component component) {
        self.components[component.name] = component;
    }

    # Expose the all the ingresses in a component via Cell Gateway
    #
    # + component - The component record
    public function exposeLocal(Component component) {
        foreach var (name, ingressTemp) in component.ingresses {
            if (ingressTemp is HttpApiIngress) {
                self.apis[name] = {
                    targetComponent: component.name,
                    ingress: ingressTemp,
                    global: false
                };
            } else if (ingressTemp is TCPIngress){
                self.tcp[name] = {
                    targetComponent: component.name,
                    ingress: ingressTemp
                };
            }
        }
    }

    # Expose a given ingress via Cell Gateway
    #
    # + component - The component record
    # + ingressName - Name of the ingress to be exposed
    public function exposeIngressLocal(Component component, string ingressName) {
        TCPIngress|HttpApiIngress? ingress = component.ingresses[ingressName];
        if (ingress is HttpApiIngress) {
            self.apis[ingressName] = {
                targetComponent: component.name,
                ingress: ingress,
                global: false
            };
        } else if (ingress is TCPIngress){
            self.tcp[ingressName] = {
                targetComponent: component.name,
                ingress: ingress
            };
        }
        else {
            error err = error("Ingress " + ingressName + " not found in the component " + component.name + ".");
            panic err;
        }
    }

    # Expose the all the ingresses in a component via Global Gateway
    #
    # + component - The component record
    public function exposeGlobal(Component component) {
        foreach var (name, ingressTemp) in component.ingresses {
            if (ingressTemp is HttpApiIngress) {
                self.apis[name] = {
                    targetComponent: component.name,
                    ingress: ingressTemp,
                    global: true
                };
            }
        }
    }

    # Expose a given ingress via Global Cell Gateway
    #
    # + component - The component record
    # + ingressName - Name of the ingress to be exposed
    public function exposeIngressGlobal(Component component, string ingressName) {
        TCPIngress|HttpApiIngress? ingress = component.ingresses[ingressName];
        if (ingress is HttpApiIngress) {
            self.apis[ingressName] = {
                targetComponent: component.name,
                ingress: ingress,
                global: true
            };
        } else if (ingress is TCPIngress){
            self.tcp[ingressName] = {
                targetComponent: component.name,
                ingress: ingress
            };
        }
        else {
            error err = error("Ingress " + ingressName + " not found in the component " + component.name + ".");
            panic err;
        }
    }

};

# Build the cell aritifacts
#
# + cellImage - The cell image definition
# + orgName - The cell  image org
# + imageName - The cell image name
# + imageVersion - The cell image version
# + return - true/false
public extern function createImage(CellImage cellImage, string orgName,
                                   string imageName, string imageVersion) returns (boolean|error);

# Update the cell aritifacts with runtime changes
#
# + cellImage - The cell image definition
# + imageName - The cell image name
# + imageVersion - The cell image version
# + instanceName - The cell instance name
# + return - true/false
public extern function createInstance(CellImage cellImage, string imageName,
                                      string imageVersion, string instanceName) returns (boolean|error);

# Parse the swagger file and returns API Defintions
#
# + swaggerFilePath - The swaggerFilePath
# + return - Array of ApiDefinitions
public extern function readSwaggerFile(string swaggerFilePath) returns (ApiDefinition[]);

public function getHost(string cellImageName, Component component) returns (string) {
    return cellImageName + "--" + getValidName(component.name) + "-service";
}

function getValidName(string name) returns string {
    return name.toLower().replace("_", "-").replace(".", "-");
}

public function setParameter(Env|Secret? param, string|int|boolean|float value) {
    if (param is (Env)) {
        param.value = value;
    } else if (param is Secret) {
        param.value = value;
    } else {
        error err = error("Parameter not declared in the component.");
        panic err;
    }
}

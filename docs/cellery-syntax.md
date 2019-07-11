## Cellery Language Syntax
This section focuses on the Cellery syntax and explain how to compose cells. 

- To get more information on the concepts and specification, please check [Cellery specification](https://github.com/wso2-cellery/spec/tree/master).

The Cellery Language is a subset of Ballerina and Ballerina extensions. Therefore the language syntax of Cellery 
resembles with normal Ballerina language syntax.

This README explains, 
* [Cell](#Cell)
* [Component](#Component)
* [Ingresses](#Ingresses)
    * [HTTP Ingress](#1.-HTTP-Ingress)
    * [Web Ingress](#2-web-ingress)
        * [Define TLS](#21-define-tls-for-web-ingress)
        * [Authentication](#22-authenticate-web-ingress)
    * [TCP Ingress](#3-tcp-ingress)
    * [GRPC Ingress](#4-grpc-ingress)
* [Environmental Variables](#envvars)
* [Resources](#resources)
* [Scaling](#scaling)
    * [Auto-Scaling](#1-auto-scaling)
    * [Zero-Scaling](#2-zero-scaling)
* [Probes](#probes)
    * [TCP Socket](#1-tcp-socket)
    * [Command](#2-command)
    * [HTTP-GET](#3-http-get)    
* [Intra-cell Communication](#intra-cell-communication)
* [Inter-cell Communication](#inter-cell-communication)
    

#### Cell
A Cell is a collection of components, APIs, Ingresses and Policies. A Cell record is initialized as follows;
Components can be added to the components map in the CellImage record.
```ballerina
cellery:CellImage helloCell = {
   components: {
       helloComp:helloWorldComp
   }
};

```

#### Component
A component represents an implementation of the business logic (in a docker image or a legacy system) and a collections 
of network accessible entry points (Ingresses) and parameters. A sample component with inline definitions would be as follows:
```ballerina
cellery:Component helloComponent = {
    name: "hello-api",
    source: {
        image: "docker.io/wso2cellery/samples-hello-world-api" // source docker image
    },
    ingresses: {
        helloApi: <cellery:HttpApiIngress>{ 
            port: 9090,
            context: "hello",
            definition: {
                resources: [
                    {
                        path: "/",
                        method: "GET"
                    }
                ]
            },
            expose: "global"
        }
    },
    envVars: {
        MESSAGE: { value: "hello" }
    }
};
```
#### Ingresses
An Ingress represents an entry point into a cell component. An ingress can be an HTTP, TCP, GRPC or Web endpoint.

##### 1. HTTP Ingress
`HttpApiIngress` supports defining HTTP API as an entry point for a cell. API definitions can be provided inline or as a swagger file. 
A sample `HttpApiIngress` instance  with inline API definition would be as follows:
```ballerina
cellery:HttpApiIngress helloAPI = {
    port: 9090,
    context: "hello",
    definition: {
        resources: [
            {
                path: "/",
                method: "GET"
            }
        ]
    },
    expose: "global"
};

```
A sample `HttpApiIngress` record  with swagger 2.0 definition can be defined as follows. The definitions are resolved at the build time. 
Therefore the build method is implemented to parse swagger file, and assign to ingress. 
```ballerina
cellery:HttpApiIngress employeeIngress = {
    port: 8080,
    context: "employee",
    expose: "local"
};

public function build(cellery:ImageName iName) returns error? {
    cellery:HttpApiIngress helloAPI = {
        port: 9090,
        context: "hello",
        definition: <cellery:ApiDefinition>cellery:readSwaggerFile("./resources/employee.swagger.json"),
        expose: "global"
    };
    ...
}
```

###### Expose
An `HttpApiIngress` can be exposed as an API by setting `expose` field. This field accepts two values.

    -  `local`: Expose an HTTP API via local cell gateway.
    -  `global`: Expose an HTTP API via global gateway.


##### 2. Web Ingress
Web cell ingress allows web traffic to the cell. A sample Web ingress would be as following: 
Web ingress are always exposed globally.

```ballerina
cellery:WebIngress webIngress = { 
    port: 8080,
    gatewayConfig: {
        vhost: "abc.com",
        context: "/demo" //default to “/”
    }
};
```
###### 2.1 Define TLS for Web Ingress
TLS can be defined to a web ingress as below:
```ballerina
// Web Component
cellery:Component webComponent = {
    name: "web-ui",
    source: {
        image: "wso2cellery/samples-hello-world-webapp"
    },
    ingresses: { 
        webUI: <cellery:WebIngress>{ // Ingress is defined in line in the ingresses map.
            port: 80,
            gatewayConfig: {
                vhost: "hello.com",
                tls: {
                    key: "",
                    cert: ""
                }
            }
        }
    }
};
// Create the cell image with cell web component.
cellery:CellImage webCell = {
      components: {
          webComp: webComponent
      }
  };

```
Values for tls key and tls cert can be assigned at the run method as below.
```ballerina
public function run(cellery:ImageName iName, map<cellery:ImageName> instance) returns error? {
    //Read TLS key file path from ENV and get the value
    string tlsKey = readFile(config:getAsString("tls.key"));
    string tlsCert = readFile(config:getAsString("tls.cert"));

    //Assign values to cell->component->ingress
    cellery:CellImage webCell = check cellery:constructCellImage(untaint iName);
    cellery:WebIngress webUI = <cellery:WebIngress>webCell.components.webComp.ingresses.webUI;
    webUI.gatewayConfig.tls.key = tlsKey;
    webUI.gatewayConfig.tls.cert = tlsCert;
    // Create the cell instance
    return cellery:createInstance(webCell, iName);
}

// Read the file given in filePath and return the content as a string.
function readFile(string filePath) returns (string) {
    io:ReadableByteChannel bchannel = io:openReadableFile(filePath);
    io:ReadableCharacterChannel cChannel = new io:ReadableCharacterChannel(bchannel, "UTF-8");

    var readOutput = cChannel.read(2000);
    if (readOutput is string) {
        return readOutput;
    } else {
        error err = error("Unable to read file " + filePath);
        panic err;
    }
}
```
###### 2.2 Authenticate Web Ingress
Web ingress support Open ID connect. OIDC config can be defined as below.
```ballerina
cellery:Component portalComponent = {
    name: "portal",
    source: {
        image: "wso2cellery/samples-pet-store-portal"
    },
    ingresses: {
        portal: <cellery:WebIngress>{ // Web ingress will be always exposed globally.
            port: 80,
            gatewayConfig: {
                vhost: "pet-store.com",
                context: "/",
                oidc: {
                    nonSecurePaths: ["/", "/app/*"],
                    providerUrl: "",
                    clientId: "",
                    clientSecret: "",
                    redirectUrl: "http://pet-store.com/_auth/callback",
                    baseUrl: "http://pet-store.com/",
                    subjectClaim: "given_name"
                }
            }
        }
    }
};
```

If dynamic client registration is used dcr configs can be provided as below in the `clientSecret` field.
```ballerina
    ingresses: {
        portal: <cellery:WebIngress>{
            port: 80,
            gatewayConfig: {
                vhost: "pet-store.com",
                context: "/portal",
                oidc: {
                    nonSecurePaths: ["/portal"], // Default [], optional field
                    providerUrl: "https://idp.cellery-system/oauth2/token",
                    clientId: "petstoreapplicationcelleryizza",
                    clientSecret: {
                        dcrUser: "admin",
                        dcrPassword: "admin"
                    },
                    redirectUrl: "http://pet-store.com/_auth/callback",
                    baseUrl: "http://pet-store.com/items/",
                    subjectClaim: "given_name"
                }
            }
        }
    },
```
Similar to above sample the `clientSecret` and `clientId` values be set at the run method to pass value at the run time without burning to the image.

##### 3. TCP Ingress
TCP ingress supports defining TCP endpoints. A sample TCP ingress would be as following:
```ballerina
cellery:TCPIngress tcpIngress = {
    backendPort: 3306,
    gatewayPort: 31406
};
```

The backendPort is the actual container port which is exposed by the container. The gatewayPort is the port exposed by the cell gateway.

##### 4. GRPC Ingress
GRPC ingress supports defining GRPC endpoints. This is similar to TCP ingress with optional field to define protofile. 
protofile field is resolved at build method since protofile is packed at build time.
```ballerina
cellery:GRPCIngress grpcIngress = {
    backendPort: 3306,
    gatewayPort: 31406
};

public function build(cellery:ImageName iName) returns error? {                                                                    
    grpcIngress.protoFile = "./resources/employee.proto";
    ...
}
```

#### EnvVars
A cell developer can require a set of environment parameters that should be passed to a Cell instance for it to be properly functional. 

```ballerina
// Employee Component
cellery:Component employeeComponent = {
    name: "employee",
    source: {
        image: "docker.io/celleryio/sampleapp-employee"
    },
    ingresses: {
        employee: <cellery:HttpApiIngress>{
            port: 8080,
            context: "employee",
            expose: "local"
        }
    },
    envVars: {
        SALARY_HOST: { value: "" }
    },
    labels: {
        team: "HR"
    }
};
```

Note the parameters SALARY_HOST in the Cell definition above. This parameters can be set in the run time of this Cell. 

```ballerina
public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    employeeCell.components.empComp.envVars.SALARY_HOST.value = config:getAsString("salary.host");
    return cellery:createInstance(employeeCell, iName);
}
```

##### Getting Host and Port from a component

Sometimes you may need to pass one component's host and port as an environment variable to another component for intra-cell 
communication. For this, you can use `cellery:getHost(<component>)`  and `cellery:getPort(<component>)` functions.

```ballerina
    cellery:Component myComponent = {
        name: "comp1",
        ...
        envVars: {
            COMP2_ADDRESS: {
                value: cellery:getHost(comp2) + ":" + cellery:getPort(comp2)
            },
        },
        ...
    };
```



#### Resources
Cellery allows to specify how much CPU and memory (RAM) each component needs. Resources can be specified as requests and limits. 
When component have resource requests specified, the scheduler can make better decisions about which nodes to place component on. 
When component have their limits specified, contention for resources on a node can be handled in a specified manner.

Following is sample syntax on how to define limits/requests.

```ballerina
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //Stock Component
    cellery:Component stockComponent = {
        name: "stock",
        source: {
            image: "wso2cellery/sampleapp-stock:0.3.0"
        },
        ingresses: {
            stock: <cellery:HttpApiIngress>{ port: 8080,
                context: "stock",
                definition: {
                    resources: [
                        {
                            path: "/options",
                            method: "GET"
                        }
                    ]
                },
                expose: "local"
            }
        },
        resources: {
            requests: {
                memory: "64Mi",
                cpu: "250m"
            },
            limits: {
                memory: "128Mi",
                cpu: "500m"
            }
        }
    };

    cellery:CellImage stockCell = {
        components: {
            stockComp: stockComponent
        }
    };
    return cellery:createImage(stockCell, untaint iName);
}
``` 

#### Scaling

##### 1. Auto-Scaling
Autoscale policies can be specified by the Cell developer at Cell creation time. 

If a component has a scaling policy as the `AutoScalingPolicy` then component will be scaled with horizontal pod autoscaler.

```ballerina
import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //Pet Component
    cellery:Component petComponent = {
        name: "pet-service",
        source: {
            image: "docker.io/isurulucky/pet-service"
        },
        ingresses: {
            stock: <cellery:HttpApiIngress>{ port: 9090,
                context: "petsvc",
                definition: {
                    resources: [
                        {
                            path: "/*",
                            method: "GET"
                        }
                    ]
                }
            }
        },
        scalingPolicy: <cellery:AutoScalingPolicy> {
            minReplicas: 1,
            maxReplicas: 10,
            metrics: {
                cpu: <cellery:Value>{ threshold : "500m" },
                memory: <cellery:Percentage> { threshold : 50 }
            }
        }

    };

    cellery:CellImage petCell = {
        components: {
            petComp: petComponent
        }
    };
    return cellery:createImage(petCell, untaint iName);
}
```
The Auto-scale policy defined by the developer can be overridden at the runtime by providing a different policy at the runtime.

##### 2. Zero-scaling

Zero scaling is powered by [Knative](https://knative.dev/v0.6-docs/). The zero-scaling have minimum replica count 0, and hence when the 
component did not get any request, the component will be terminated and it will be running back once a request was directed to the component. 

A zero-scaling policy can be defined as below.
```ballerina
import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //Pet Component
    cellery:Component petComponent = {
        name: "pet-service",
        source: {
            image: "docker.io/isurulucky/pet-service"
        },
        ingresses: {
            stock: <cellery:HttpApiIngress>{ port: 9090,
                context: "petsvc",
                definition: {
                    resources: [
                        {
                            path: "/*",
                            method: "GET"
                        }
                    ]
                }
            }
        },
        scalingPolicy: <cellery:ZeroScalingPolicy> {
                maxReplicas: 10,
                concurrencyTarget: 25
        }

    };

    cellery:CellImage petCell = {
        components: {
            petComp: petComponent
        }
    };
    return cellery:createImage(petCell, untaint iName);
}
```  

#### Probes
Liveness and readiness probes can be defined for a component. 
The probes can be defined by the means of tcp socket, command or as a http-get. 

##### 1. TCP Socket
```ballerina
import celleryio/cellery;
public function build(cellery:ImageName iName) returns error? {
    int salaryContainerPort = 8080;

    // Salary Component
    cellery:Component salaryComponent = {
        name: "salary",
        source: {
            image: "docker.io/celleryio/sampleapp-salary"
        },
        ingresses: {
            SalaryAPI: <cellery:HttpApiIngress>{
                port:salaryContainerPort,
                context: "payroll",
                definition: {
                    resources: [
                        {
                            path: "salary",
                            method: "GET"
                        }
                    ]
                },
                expose: "global"
            }
        },
        probes: {
            liveness: {
                initialDelaySeconds: 30,
                kind: <cellery:TcpSocket>{
                    port:salaryContainerPort
                }
            }
        }
    };
}
``` 

##### 2. Command
```ballerina
import celleryio/cellery;
public function build(cellery:ImageName iName) returns error? {
    int salaryContainerPort = 8080;

    // Salary Component
    cellery:Component salaryComponent = {
        name: "salary",
        source: {
            image: "docker.io/celleryio/sampleapp-salary"
        },
        ingresses: {
            SalaryAPI: <cellery:HttpApiIngress>{
                port:salaryContainerPort,
                context: "payroll",
                definition: {
                    resources: [
                        {
                            path: "salary",
                            method: "GET"
                        }
                    ]
                },
                expose: "global"
            }
        },
        probes: {
            readiness: {
                initialDelaySeconds: 10,
                timeoutSeconds: 50,
                kind: <cellery:Exec>{
                    commands: ["bash", "-version"]
                }
            }
        }
    };
}
```

##### 3. HTTP-GET
```ballerina
import celleryio/cellery;
public function build(cellery:ImageName iName) returns error? {
    int salaryContainerPort = 8080;

    // Salary Component
    cellery:Component salaryComponent = {
        name: "salary",
        source: {
            image: "docker.io/celleryio/sampleapp-salary"
        },
        ingresses: {
            SalaryAPI: <cellery:HttpApiIngress>{
                port:salaryContainerPort,
                context: "payroll",
                definition: {
                    resources: [
                        {
                            path: "salary",
                            method: "GET"
                        }
                    ]
                },
                expose: "global"
            }
        },
        probes: {
            readiness: {
                initialDelaySeconds: 10,
                timeoutSeconds: 50,
                kind: <cellery:HttpGet>{
                    port: salaryContainerPort,
                    path: "/healthz"
                }
            }
        }
    };
}
```

### Intra Cell Communication

Cell components can communicate with each other. This is achieved via environment variables. Two components can be linked via 
environment variables in the run method. For an example, consider the scenario below. The employee component expects 
`SALARY_HOST`, which is the hostname of salary component.

Employee component:
```ballerina
import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    int salaryContainerPort = 8080;

    // Salary Component
    cellery:Component salaryComponent = {
        name: "salary",
        source: {
            image: "docker.io/celleryio/sampleapp-salary"
        },
        ingresses: {
            SalaryAPI: <cellery:HttpApiIngress>{
                port:salaryContainerPort,
                context: "payroll",
                definition: {
                    resources: [
                        {
                            path: "salary",
                            method: "GET"
                        }
                    ]
                },
                expose: "local"
            }
        },
        labels: {
            team: "Finance",
            owner: "Alice"
        }
    };

    // Employee Component
    cellery:Component employeeComponent = {
        name: "employee",
        source: {
            image: "docker.io/celleryio/sampleapp-employee"
        },
        ingresses: {
            employee: <cellery:HttpApiIngress>{
                port: 8080,
                context: "employee",
                expose: "local",
                definition: <cellery:ApiDefinition>cellery:readSwaggerFile("./resources/employee.swagger.json")
            }
        },
        envVars: {
            SALARY_HOST: {
                value: cellery:getHost(salaryComponent)
            },
            PORT: {
                value: salaryContainerPort
            }
        },
        labels: {
            team: "HR"
        }
    };

    cellery:CellImage employeeCell = {
        components: {
            empComp: employeeComponent,
            salaryComp: salaryComponent
        }
    };

    return cellery:createImage(employeeCell, untaint iName);
}
```

The envVar `SALARY_HOST` value is provided at the build time as shown above, which enables the employee component to 
communicate with the salary component. 


### Inter Cell Communication

In addition to components within a cell, Cells themselves can communicate with each other. This is also achieved 
via envVars. Two cells can be linked via envVar in the run method. 

When a Cell Image is built it will generate a reference file describing the APIs/Ports that are exposed by itself. 
This cell reference will be available locally, either when you build or pull the image. This reference can be imported 
in another Cell definition which is depending on the former, and can be used to link the two Cells at the runtime. 

Consider following cell definition:
```ballerina
import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //Build Stock Cell
    io:println("Building Stock Cell ...");
    //Stock Component
    cellery:Component stockComponent = {
        name: "stock",
        source: {
            image: "docker.io/celleryio/sampleapp-stock"
        },
        ingresses: {
            stock: <cellery:HttpApiIngress>{ port: 8080,
                context: "stock",
                definition: {
                    resources: [
                        {
                            path: "/options",
                            method: "GET"
                        }
                    ]
                },
                expose: "local"
            }
        }
    };
    cellery:CellImage stockCell = {
        components: {
            stockComp: stockComponent
        }
    };
    return cellery:createImage(stockCell, untaint iName);
}
```

Generated reference file for above cell definition is as follows: 
```json
{
  "stock_api_url":"http://{{instance_name}}--gateway-service:80/stock"
}
```

If a cell component wants to access the stock api, it can be done as below:

```ballerina
public function build(cellery:ImageName iName) returns error? {
    //HR component
    cellery:Component hrComponent = {
        name: "hr",
        source: {
            image: "docker.io/celleryio/sampleapp-hr"
        },
        ingresses: {
            "hr": <cellery:HttpApiIngress>{
                port: 8080,
                context: "hr-api",
                definition: {
                    resources: [
                        {
                            path: "/",
                            method: "GET"
                        }
                    ]
                },
                expose: "global"
            }
        },
        envVars: {
            stock_api_url: { value: "" }
        },
        dependencies: {
            cells: {
                stockCellDep: <cellery:ImageName>{ org: "myorg", name: "stock", ver: "1.0.0" } // dependency as a struct
            }
        }
    };

    hrComponent.envVars = {
        stock_api_url: { value: <string>cellery:getReference(hrComponent, "stockCellDep").stock_api_url }
    };

    // Cell Initialization
    cellery:CellImage hrCell = {
        components: {
            hrComp: hrComponent
        }
    };
    return cellery:createImage(hrCell, untaint iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    cellery:CellImage hrCell = check cellery:constructCellImage(untaint iName);
    return cellery:createInstance(hrCell, iName, instances);
}
```
 
The `hrComponent` depends on the stockCell that is defined earlier. The dependency information are specified as a component attribute.
```ballerina
    dependencies: {
        cells: {
            stockCellDep: <cellery:ImageName>{ org: "myorg", name: "stock", ver: "1.0.0" } // dependency as a struct
        }
    }
```

`cellery:getReference(hrComponent, "stockCellDep").stock_api_url` method reads the json file and set value with place holder. 

Note the `run` method above, which takes a variable argument map for the references of the dependency cells. 
These are names of already deployed cell instances, which will be used to resolve the urls and link with this 
cell instance. As an example, if the stock cell instance name is `stock-app` then the `stockRef.stock_api_url` 
returns the host name of the running stock cell instance as `http://stock-app--gateway-service:80/stock`.

# What's Next?
- [Developing a Cell](writing-a-cell.md) - step by step explanation on how you could define your own cells.
- [Samples](https://github.com/wso2-cellery/samples/tree/master) - a collection of useful samples.

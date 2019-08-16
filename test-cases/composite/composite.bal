import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    int fooComponentPort = 9090;
    cellery:Component hrCompComponent = {
        name: "hr",
        source: {
            image: "docker.io/celleryio/sampleapp-hr"
        },
        ingresses: {
            https:<cellery:HttpsPortIngress>{port: 8443}
        },
        dependencies: {
            composites: {
                //  fully qualified dependency image name as a string
                employeeCellDep: "myorg/employee:1.0.0",
                // dependency as a struct
                stockCellDep: <cellery:ImageName>{ org: "myorg", name: "stock", ver: "1.0.0" }
            }
        },
        probes: {
            liveness: {
                initialDelaySeconds: 30,
                kind: <cellery:TcpSocket>{
                            port:fooComponentPort
                        }
            },
            readiness: {
                initialDelaySeconds: 60,
                timeoutSeconds: 50,
                kind: <cellery:HttpGet>{
                                        port: 80,
                                        path: "/",
                                        httpHeaders:{
                                            myCustomHeader: "customerHeaderValue"
                                        }
                                    }
            }
        }

    };

    hrCompComponent.envVars = {
        employee_api_url: {
            value: <string>cellery:getReference(hrCompComponent, "employeeCellDep").employee_api_url
        },
        stock_api_url: {
            value: <string>cellery:getReference(hrCompComponent, "stockCellDep").stock_api_url
        }
    };

    cellery:Component fooCompComponent = {
        name: "foo",
        source: {
            image: "docker.io/celleryio/sampleapp-foo"
        },
        ingresses: {
            http: <cellery:HttpPortIngress>{
                port: fooComponentPort
            }
        }
    };

    // Declare the composite
    cellery:Composite hrComposite = {
        components: {
            hrComp: hrCompComponent
        }
    };
    return cellery:createImage(hrComposite, untaint iName);// this will create an image with type = composite
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    cellery:Composite hrComposite = check cellery:constructImage(untaint iName);
    return cellery:createInstance(hrComposite, untaint iName, instances);
}



import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    // Web Component
    cellery:Component helloComponent = {
        name: "hello",
        source: {
            image: "wso2cellery/samples-hello-world-webapp"
        },
        ingresses: {
            webUI: <cellery:WebIngress> { // Web ingress will be always exposed globally.
                port: 80,
                gatewayConfig: {
                    vhost: "hello-world.com",
                    context: "/"
                }
            }
        },
        envVars: {
            HELLO_NAME: {value: "Cellery"}
        }
    };

    cellery:CellImage helloCell = {
        components: {
            helloComp: helloComponent
        }
    };

    return cellery:createImage(helloCell, untaint iName);
}


public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:CellImage helloCell = check cellery:constructCellImage(untaint iName);
    return cellery:createInstance(helloCell, iName, instances, startDependencies, shareDependencies);
}

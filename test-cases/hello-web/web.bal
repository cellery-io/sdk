import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    // Web Component
    cellery:Component webComponent = {
        name: "web-ui",
        source: {
            image: "docker.io/celleryio/sampleapp-employee"
        },
        ingresses: {
            webUI: <cellery:WebIngress>{    //web ingress will be always exposed globally.
                port: 8080,
                gatewayConfig: {
                    vhost: "abc.com",
                    context: "/demo" //default to “/”
                }
            }
        }
    };

    cellery:CellImage webCell = {
        components: {
            webComp: webComponent
        }
    };

    return cellery:createImage(webCell, untaint iName);
}


public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    cellery:CellImage webCell = check cellery:constructCellImage(untaint iName);
    return cellery:createInstance(webCell, iName, instances);
}

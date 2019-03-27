import ballerina/io;
import celleryio/cellery;


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

public function build(cellery:ImageName iName) returns error? {
    return cellery:createImage(webCell, iName);
}


public function run(cellery:ImageName iName, map<string> instances) returns error? {
    return cellery:createInstance(webCell, iName);
}

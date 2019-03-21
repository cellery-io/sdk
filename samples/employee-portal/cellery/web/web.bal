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
            uri: {
                vhost: "abc.com",
                context: "/demo" //default to “/”
            }
        }
    }
};


cellery:CellImage webCell = {
    components: [
        webComponent
    ]
};

public function build(cellery:StructuredName sName) returns error? {
    return cellery:createImage(webCell, sName);
}


public function run(cellery:StructuredName sName, map<string> instances) returns error? {
    return cellery:createInstance(webCell, sName);
}


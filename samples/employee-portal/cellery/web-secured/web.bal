import ballerina/io;
import celleryio/cellery;
import ballerina/config;


// Web Component
cellery:Component webComponent = {
    name: "web-ui",
    source: {
        image: "docker.io/celleryio/sampleapp-employee"
    },
    ingresses: {
        webUI: <cellery:WebIngress>{
            port: 8080,
            uri: {
                vhost: "abc.com",
                context: "/demo" //default to “/”
            },
            tls: {
                key: "",
                cert: ""
            },
            oidc: {
                context: ["/testApp"],
                provider: "https://accounts.google.com",
                clientId: "",
                clientSecret: "",
                redirectUrl: "http://pet-store.com/_auth/callback",
                baseUrl: "http://pet-store.com/items/",
                subjectClaim: "given_name"
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


public function run(cellery:StructuredName sName, map<string> instance) returns error? {
    // Read key and crt values from the environment at runtime.
    cellery:WebIngress webUI = <cellery:WebIngress>webCell.components[0].ingresses.webUI;
    webUI.tls.key = "key";
    webUI.tls.cert = "cert";
    webUI.oidc.clientId = "id";
    webUI.oidc.clientSecret = "secret";
    return cellery:createInstance(webCell, sName);
}

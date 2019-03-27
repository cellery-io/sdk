import ballerina/io;
import celleryio/cellery;
import ballerina/config;


// Web Component
cellery:Component webComponent = {
    name: "web-ui",
    source: {
        image: "docker.io/celleryio/sampleapp-petstore"
    },
    ingresses: {
        webUI: <cellery:WebIngress>{
            port: 8080,
            gatewayConfig: {
                vhost: "pet-store.com",
                context: "/items", //default to “/”
                tls: {
                    key: "",
                    cert: ""
                },
                oidc:
                {
                    nonSecureContexts: ["/testApp"], // Default [], optional field
                    provider: "https://accounts.google.com",
                    clientId: "",
                    clientSecret: "",
                    redirectUrl: "http://pet-store.com/_auth/callback",
                    baseUrl: "http://pet-store.com/items/",
                    subjectClaim: "given_name"
                }
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


public function run(cellery:ImageName iName, map<string> instance) returns error? {
    // Read key and crt values from the environment at runtime.
    cellery:WebIngress webUI = <cellery:WebIngress>webCell.components.webComp.ingresses.webUI;
    webUI.gatewayConfig.tls.key = config:getAsString("tls.key");
    webUI.gatewayConfig.tls.cert = config:getAsString("tls.cery");
    webUI.gatewayConfig.oidc.clientId = config:getAsString("google.client.key");
    webUI.gatewayConfig.oidc.clientSecret = config:getAsString("google.client.secret");
    return cellery:createInstance(webCell, iName);
}

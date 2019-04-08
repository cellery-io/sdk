import ballerina/io;
import celleryio/cellery;
import ballerina/config;


// Web Component
cellery:Component webComponent = {
    name: "web-ui",
    source: {
        image: "wso2cellery/samples-hello-world-webapp"
    },
    ingresses: {
        webUI: <cellery:WebIngress>{
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


cellery:CellImage webCell = {
    components: {
        webComp: webComponent
    }
};

public function build(cellery:ImageName iName) returns error? {
    return cellery:createImage(webCell, iName);
}


public function run(cellery:ImageName iName, map<cellery:ImageName> instance) returns error? {
    //Read TLS key file path from ENV and get the value
    string tlsKey = readFile(config:getAsString("tls.key"));
    string tlsCert = readFile(config:getAsString("tls.cert"));

    //Assign values to cell
    cellery:WebIngress webUI = <cellery:WebIngress>webCell.components.webComp.ingresses.webUI;
    webUI.gatewayConfig.tls.key = tlsKey;
    webUI.gatewayConfig.tls.cert = tlsCert;
    return cellery:createInstance(webCell, iName);
}


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

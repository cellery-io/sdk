import ballerina/io;
import celleryio/cellery;
import ballerina/config;

public function build(cellery:ImageName iName) returns error? {
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
    return cellery:createImage(webCell, untaint iName);
}


public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    //Read TLS key file path from ENV and get the value
    string tlsKey = readFile(config:getAsString("tls.key", defaultValue = "./certs/95749524_hello.com.key"));
    string tlsCert = readFile(config:getAsString("tls.cert", defaultValue = "./certs/95749524_hello.com.cert"));

    //Assign values to cell
    cellery:CellImage webCell = check cellery:constructCellImage(untaint iName);
    cellery:WebIngress webUI = <cellery:WebIngress>webCell.components.webComp.ingresses.webUI;
    webUI.gatewayConfig.tls.key = tlsKey;
    webUI.gatewayConfig.tls.cert = tlsCert;
    return cellery:createInstance(webCell, iName, instances, startDependencies, shareDependencies);
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

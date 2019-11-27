Cellery is a code-first approach to building, integrating, running and managing composite applications on Kubernetes.

# Module Overview

Cellery is a code-first approach to building, integrating, running and managing composite applications on Kubernetes.
Build, push/pull, test, deploy, update, scale and observe cells. 
Create secure APIs by default. Use your existing container images, written in any language.

Visit [Cellery website](https://wso2-cellery.github.io/) for more info.

## Samples

Sample Hello World program is as below. 
Please visit [Cellery Samples Repository](https://github.com/wso2/cellery-samples) for more examples. 

```ballerina
import ballerina/config;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    // Hello Component
    // This Components exposes the HTML hello world page
    cellery:Component helloComponent = {
        name: "hello",
        src: {
            image: "wso2cellery/samples-hello-world-webapp"
        },
        ingresses: {
            webUI: <cellery:WebIngress>{ // Web ingress will be always exposed globally.
                port: 80,
                gatewayConfig: {
                    vhost: "hello-world.com",
                    context: "/"
                }
            }
        },
        envVars: {
            HELLO_NAME: { value: "Cellery" }
        }
    };

    // Cell Initialization
    cellery:CellImage helloCell = {
        components: {
            helloComp: helloComponent
        }
    };
    return <@untainted> cellery:createImage(helloCell, iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:CellImage helloCell = check cellery:constructCellImage(<@untainted> iName);
    string vhostName = config:getAsString("VHOST_NAME");
    if (vhostName !== "") {
        cellery:WebIngress web = <cellery:WebIngress>helloCell.components["helloComp"]["ingresses"]["webUI"];
        web.gatewayConfig.vhost = vhostName;
    }

    string helloName = config:getAsString("HELLO_NAME");
    if (helloName !== "") {
        helloCell.components["helloComp"]["envVars"]["HELLO_NAME"].value = helloName;
    }
    return <@untainted> cellery:createInstance(helloCell, iName, instances, startDependencies, shareDependencies);
}
```
import ballerina/io;
import ballerina/log;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //Build Stock Cell
    io:println("Building Stock Cell ...");
    //Stock Component
    cellery:Component stockComponent = {
        name: "stock",
        source: {
            image: "wso2cellery/sampleapp-stock:0.3.0"
        },
        ingresses: {
            stock: <cellery:HttpApiIngress>{ port: 8080,
                context: "stock",
                definition: {
                    resources: [
                        {
                            path: "/options",
                            method: "GET"
                        }
                    ]
                },
                expose: "local"
            }
        },
        resources: {
            requests: {
                memory: "64Mi",
                cpu: "250m"
            },
            limits: {
                memory: "128Mi",
                cpu: "500m"
            }
        }
    };

    cellery:CellImage stockCell = {
        components: {
            stockComp: stockComponent
        }
    };
    return cellery:createImage(stockCell, untaint iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:CellImage stockCell = check cellery:constructCellImage(untaint iName);
    stockCell.components.stockComp.resources.requests= {
            memory: "256Mi",
            cpu: "256m"
        };
    return cellery:createInstance(stockCell, iName, instances, startDependencies, shareDependencies);
}

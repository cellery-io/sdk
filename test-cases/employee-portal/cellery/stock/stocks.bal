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
        }
    };

    cellery:CellImage stockCell = {
        components: {
            stockComp: stockComponent
        }
    };
    return cellery:createImage(stockCell, untaint iName);
}

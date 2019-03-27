import ballerina/io;
import celleryio/cellery;

//Stock Component
cellery:Component stockComponent = {
    name: "stock",
    source: {
        image: "docker.io/celleryio/sampleapp-stock"
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

public function build(cellery:ImageName iName) returns error? {
    //Build Stock Cell
    io:println("Building Stock Cell ...");
    return cellery:createImage(stockCell,iName);
}

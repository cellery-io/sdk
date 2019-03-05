import ballerina/io;
import celleryio/cellery;

//Stock Component
cellery:Component stock = {
    name: "stock",
    source: {
        image: "docker.io/celleryio/sampleapp-stock"
    },
    ingresses: {
        stock: new cellery:HttpApiIngress(8080,
            "stock",
            [
                {
                    path: "/options",
                    method: "GET"
                }
            ]
        )
    }
};

cellery:CellImage stockCell = new();

public function build(string orgName, string imageName, string imageVersion) {
    //Build Stock Cell
    io:println("Building Stock Cell ...");
    stockCell.addComponent(stock);
    //Expose API from Cell Gateway
    stockCell.exposeLocal(stock);
    _ = cellery:createImage(stockCell, orgName, imageName, imageVersion);
}

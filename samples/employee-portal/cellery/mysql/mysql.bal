import ballerina/io;
import celleryio/cellery;

//MySQL Component
cellery:Component mysqlComponent = {
    name: "mysql",
    source: {
        image: "mirage20/samples-productreview-mysql"
    },
    ingresses: {
        mysqlIngress: new cellery:TCPIngress(3306,31406)
    },
    parameters: {
        MYSQL_ROOT_PASSWORD: new cellery:Env(default = "root")
    }
};

cellery:CellImage mysqlCell = new();

public function build(string imageName, string imageVersion) {
    //Build MySQL Cell
    io:println("Building MySQL Cell ...");
    mysqlCell.addComponent(mysqlComponent);
    //Expose API from Cell Gateway
    mysqlCell.exposeAPIsFrom(mysqlComponent);
    _ = cellery:createImage(mysqlCell, imageName, imageVersion);
}

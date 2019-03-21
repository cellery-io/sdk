import ballerina/io;
import celleryio/cellery;

//MySQL Component
cellery:Component mysqlComponent = {
    name: "mysql",
    source: {
        image: "mirage20/samples-productreview-mysql"
    },
    ingresses: {
        mysqlIngress: <cellery:TCPIngress>{
            backendPort: 3306,
            gatewayPort: 31406,
            expose: "local"
        }
    },
    envVars: {
        MYSQL_ROOT_PASSWORD: { value: "root" }
    }
};

cellery:CellImage mysqlCell = {
    components: [
        mysqlComponent
    ]
};

public function build(cellery:StructuredName sName) {
    //Build MySQL Cell
    _ = cellery:createImage(mysqlCell, sName);
}

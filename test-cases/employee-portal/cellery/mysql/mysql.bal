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
            gatewayPort: 31406
        }
    },
    envVars: {
        MYSQL_ROOT_PASSWORD: { value: "root" }
    }
};

cellery:CellImage mysqlCell = {
    components: {
        mysqlComp: mysqlComponent
    }
};

public function build(cellery:ImageName iName) returns error? {
    //Build MySQL Cell
    return cellery:createImage(mysqlCell, untaint iName);
}

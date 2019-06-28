import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //HR component
    cellery:Component hrComponent = {
        name: "hr",
        source: {
            image: "wso2cellery/sampleapp-hr:0.3.0"
        },
        ingresses: {
            "hr": <cellery:HttpApiIngress>{
                port: 8080,
                context: "hr-api",
                definition: {
                    resources: [
                        {
                            path: "/",
                            method: "GET"
                        }
                    ]
                },
                expose: "global"
            }
        },
        envVars: {
            employee_api_url: { value: "" },
            stock_api_url: { value: "" }
        },
        dependencies: {
            cells: {
                employeeCellDep: "myorg/employee:1.0.0", //  fully qualified dependency image name as a string
                stockCellDep: <cellery:ImageName>{ org: "myorg", name: "stock", ver: "1.0.0" } // dependency as a struct
            }
        }
    };

    hrComponent.envVars = {
        employee_api_url: { value: <string>cellery:getReference(hrComponent, "employeeCellDep").employee_api_url },
        stock_api_url: { value: <string>cellery:getReference(hrComponent, "stockCellDep").stock_api_url }
    };

    // Cell Initialization
    cellery:CellImage hrCell = {
        components: {
            hrComp: hrComponent
        }
    };
    return cellery:createImage(hrCell, untaint iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    cellery:CellImage hrCell = check cellery:constructCellImage(untaint iName);
    return cellery:createInstance(hrCell, iName, instances);
}

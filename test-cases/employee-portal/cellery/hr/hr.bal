import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //HR component
    cellery:Component hrComponent = {
        name: "hr",
        source: {
            image: "docker.io/celleryio/sampleapp-hr"
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
            employeeCellDep: "myorg/employee:1.0.0", //  fully qualified dependency image name as a string
            stockCellDep: <cellery:ImageName>{ org: "myorg", name: "stock", ver: "1.0.0" } // dependency as a struct
        }
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
    //Resolve employee API URL
    cellery:Reference employeeRef = check cellery:getReference(instances.employeeCellDep);
    hrCell.components.hrComp.envVars.employee_api_url.value = <string>employeeRef.employee_api_url;

    //Resolve stock API URL
    cellery:Reference stockRef = check cellery:getReference(instances.stockCellDep);
    hrCell.components.hrComp.envVars.stock_api_url.value = <string>stockRef.stock_api_url;
    return cellery:createInstance(hrCell, iName, instances);
}

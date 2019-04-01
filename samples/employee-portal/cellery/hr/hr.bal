import ballerina/io;
import celleryio/cellery;
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
        employeeCellDep: <cellery:ImageName>{ org: "myorg", name: "employee", ver: "1.0.0" },
        stockCellDep: <cellery:ImageName>{ org: "myorg", name: "stock", ver: "1.0.0" }
    }
};

// Cell Intialization
cellery:CellImage hrCell = {
    components: {
        hrComp: hrComponent
    }
};

public function build(cellery:ImageName iName) returns error? {
    return cellery:createImage(hrCell, iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    //Resolve employee API URL
    cellery:Reference employeeRef = check cellery:getReferenceRecord(instances.employeeCellDep);
    hrCell.components.hrComp.envVars.employee_api_url.value = <string>employeeRef.employee_api_url;
    io:println(hrCell);

    //Resolve stock API URL
    cellery:Reference stockRef = check cellery:getReferenceRecord(instances.stockCellDep);
    hrCell.components.hrComp.envVars.stock_api_url.value = <string>stockRef.stock_api_url;
    return cellery:createInstance(hrCell, iName);
}

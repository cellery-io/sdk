import ballerina/io;
import celleryio/cellery;
import myorg/employee;
import myorg/stock;

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

public function run(cellery:ImageName iName, map<string> instances) returns error? {
    //Resolve employee gateway URL
    employee:EmployeeReference employeeRef = cellery:getReferenceRecord(instances.employeeCellDep);
    hrCell.hrComp.envVars.employee_api_url.value = employeeRef.gatewayHost;

    //Resolve stock gateway URL
    stock:StockReference stockRef = cellery:getReferenceRecord(instances.stockCellDep);
    hrCell.components.hrComp.envVars.stock_api_url.value = stockRef.gatewayHost;
    return cellery:createInstance(hrCell, iName);
}
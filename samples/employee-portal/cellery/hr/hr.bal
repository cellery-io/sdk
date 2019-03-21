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
            definitions: [
                {
                    path: "/",
                    method: "GET"
                }
            ],
            expose: "global"
        }
    },
    envVars: {
        employee_api_url: { value: "" },
        stock_api_url: { value: "" }
    },
    dependencies: {
        employeeCellDep: <cellery:StructuredName>{ orgName: "myorg", imageName: "employee", imageVersion: "1.0.0" },
        stockCellDep: <cellery:StructuredName>{ orgName: "myorg", imageName: "stock", imageVersion: "1.0.0" }
    }
};

// Cell Intialization
cellery:CellImage hrCell = {
    components: [
        hrComponent
    ]
};

public function build(cellery:StructuredName sName) returns error? {
    return cellery:createImage(hrCell, sName);
}

public function run(cellery:StructuredName sName, map<string> instances) returns error? {
    //Resolve employee gateway URL
    employee:EmployeeReference employeeRef = employee:getReferenceRecord(instances.employeeCellDep);
    hrCell.components[0].envVars.employee_api_url.value = employeeRef.gatewayHost;

    //Resolve stock gateway URL
    stock:StockReference stockRef = getReferenceRecord(instances.stockCellDep);
    hrCell.components[0].envVars.stock_api_url.value = stockRef.gatewayHost;
    return cellery:createInstance(hrCell, sName);
}
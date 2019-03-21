import ballerina/io;
import ballerina/config;
import celleryio/cellery;

int salaryContainerPort = 8080;
// Read API defintion from swagger file.
cellery:ApiDefinition[] employeeAPIdefn = (<cellery:ApiDefinition[]>cellery:readSwaggerFile(
                                                                        "./resources/employee.swagger.json"));

// Employee Component
cellery:Component employeeComponent = {
    name: "employee",
    source: {
        image: "docker.io/celleryio/sampleapp-employee"
    },
    ingresses: {
        employee: <cellery:HttpApiIngress>{
            port: 8080,
            context: "employee",
            definitions: employeeAPIdefn,
            expose: "local"
        }
    },
    envVars: {
        SALARY_HOST: { value: "" },
        PORT: { value: salaryContainerPort }
    },
    labels: {
        team: "HR"
    }
};

// Salary Component
cellery:Component salaryComponent = {
    name: "salary",
    source: {
        image: "docker.io/celleryio/sampleapp-salary"
    },
    ingresses: {
        SalaryAPI: <cellery:HttpApiIngress>{
            port:salaryContainerPort,
            context: "payroll",
            definitions: [
                {
                    path: "salary",
                    method: "GET"
                }
            ],
            expose: "local"
        }
    },
    labels: {
        team: "Finance",
        owner: "Alice"
    }
};

cellery:CellImage employeeCell = {
    components: [
        employeeComponent,
        salaryComponent
    ]
};

public function build(cellery:StructuredName sName) returns error? {
    return cellery:createImage(employeeCell, sName);
}

public function run(cellery:StructuredName sName, map<string> instances) returns error? {
    employeeCell.components[0].envVars.SALARY_HOST.value = cellery:getHost(untaint sName.instanceName, salaryComponent);
    io:println(employeeCell);
    //return cellery:createInstance(employeeCell, sName);
}

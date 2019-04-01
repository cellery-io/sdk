import ballerina/io;
import celleryio/cellery;

int salaryContainerPort = 8080;

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
            definition: (),
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
            definition: {
                resources: [
                    {
                        path: "salary",
                        method: "GET"
                    }
                ]
            },
            expose: "local"
        }
    },
    labels: {
        team: "Finance",
        owner: "Alice"
    }
};

cellery:CellImage employeeCell = {
    components: {
        empComp: employeeComponent,
        salaryComp: salaryComponent
    }
};

public function build(cellery:ImageName iName) returns error? {
    cellery:ApiDefinition employeeAPIdefn = (<cellery:ApiDefinition>cellery:readSwaggerFile(
                                                                        "./resources/employee.swagger.json"));
    cellery:HttpApiIngress httpAPI = <cellery:HttpApiIngress>(employeeComponent.ingresses.employee);
    httpAPI.definition = employeeAPIdefn;
    return cellery:createImage(employeeCell, iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    employeeCell.components.empComp.envVars.SALARY_HOST.value = cellery:getHost(untaint iName.instanceName,
        salaryComponent);
    return cellery:createInstance(employeeCell, iName);
}

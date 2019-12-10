import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    int salaryContainerPort = 8080;

    // Salary Component
    cellery:Component salaryComponent = {
        name: "salary",
        src: {
            image: "wso2cellery/sampleapp-salary:0.3.0"
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

    // Employee Component
    cellery:Component employeeComponent = {
        name: "employee",
        src: {
            image: "wso2cellery/sampleapp-employee:0.3.0"
        },
        ingresses: {
            employee: <cellery:HttpApiIngress>{
                port: 8080,
                context: "employee",
                expose: "local",
                definition: <cellery:ApiDefinition>cellery:readSwaggerFile("./src/employee/resources/employee.swagger.json")
            }
        },
        envVars: {
            SALARY_HOST: {
                value: cellery:getHost(salaryComponent)
            }
        },
        labels: {
            team: "HR"
        },
        dependencies:{
            components:[salaryComponent]
        }
    };

    cellery:CellImage employeeCell = {
        components: {
            empComp: employeeComponent,
            salaryComp: salaryComponent
        }
    };

    return <@untainted> cellery:createImage(employeeCell, iName);
}


public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:CellImage|cellery:Composite employeeCell = cellery:constructImage(iName);
    return <@untainted> cellery:createInstance(employeeCell, iName, instances, startDependencies, shareDependencies);
}

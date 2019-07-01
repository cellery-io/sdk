import ballerina/config;
import ballerina/io;
import ballerina/log;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    int salaryContainerPort = 8080;

    // Salary Component
    cellery:Component salaryComponent = {
        name: "salary",
        source: {
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
        source: {
            image: "wso2cellery/sampleapp-employee:0.3.0"
        },
        ingresses: {
            employee: <cellery:HttpApiIngress>{
                port: 8080,
                context: "employee",
                expose: "local",
                definition: <cellery:ApiDefinition>cellery:readSwaggerFile("./resources/employee.swagger.json")
            }
        },
        autoscaling: {
            policy: {
                minReplicas: 0,
                maxReplicas: 10,
                cpuPercentage: <cellery:CpuUtilizationPercentage>{ percentage: 50 }
            }
        }
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

    return cellery:createImage(employeeCell, untaint iName);
}


public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    cellery:CellImage employeeCell = check cellery:constructCellImage(untaint iName);
    return cellery:createInstance(employeeCell, iName, instances);
}
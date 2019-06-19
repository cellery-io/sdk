import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    int salaryContainerPort = 8080;

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
        probes: {
            liveness: {
                initialDelaySeconds: 30,
                kind: <cellery:TcpSocket>{
                    port:salaryContainerPort
                }
            }
        }
    };

    // Employee Component
    int empPort = 8080;
    cellery:Component employeeComponent = {
        name: "employee",
        source: {
            image: "docker.io/celleryio/sampleapp-employee"
        },
        ingresses: {
            employee: <cellery:HttpApiIngress>{
                port:empPort,
                context: "employee",
                expose: "local"
            }
        },
        probes: {
            liveness: {
                initialDelaySeconds: 30,
                kind: <cellery:TcpSocket>{
                    port:empPort
                }
            },
            readiness: {
                initialDelaySeconds: 10,
                timeoutSeconds: 50,
                kind: <cellery:Exec>{
                    commands: ["bin", "bash", "-version"]
                }
            }
        },
        envVars: {
            SALARY_HOST: {
                value: cellery:getHost(salaryComponent)
            }
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


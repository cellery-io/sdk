import ballerina/io;
import ballerina/config;
import celleryio/cellery;

//Employee Component
cellery:Component employeeComponent = {
    name: "employee",
    source: {
        image: "docker.io/wso2vick/sampleapp-employee"
    },
    ingresses: {
        employee: new cellery:HTTPIngress(
                      8080,
                      "employee",
                      [
                          {
                              path: "/",
                              method: "GET"
                          }
                      ]
        )
    },
    parameters: {
        SALARY: new cellery:Env(),
        PORT: new cellery:Env(default = 8080),
        BASE_PATH: new cellery:Env()
    },
    labels: {
        cellery:TEAM:"HR"
    }
};

//Salary Component
cellery:Component salaryComponent = {
    name: "salary",
    source: {
        image: "docker.io/wso2vick/sampleapp-salary"
    },
    ingresses: {
        salaryAPI: new cellery:HTTPIngress(8080, "payroll",
            [
                {
                    path: "/salary",
                    method: "GET"
                }
            ])
    },
    labels: {
        cellery:TEAM:"Finance",
        cellery:OWNER:"Alice"
    }
};

public cellery:CellImage employeeCell = new("Employee");

public function build() {

    // Build EmployeeCell
    io:println("Building Employee Cell ...");

    // Map component parameters
    cellery:setParameter(employeeComponent.parameters["SALARY"], cellery:getHost(employeeCell, salaryComponent));
    cellery:setParameter(employeeComponent.parameters["BASE_PATH"],
        cellery:getBasePath(salaryComponent.ingresses["salaryAPI"]));

    // Add components to Cell
    employeeCell.addComponent(employeeComponent);
    employeeCell.addComponent(salaryComponent);

    //Expose API from Cell Gateway
    employeeCell.exposeAPIsFrom(employeeComponent);

    _ = cellery:createImage(employeeCell);
}

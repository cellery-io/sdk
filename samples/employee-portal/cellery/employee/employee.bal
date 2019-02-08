import ballerina/io;
import ballerina/config;
import celleryio/cellery;

cellery:Component employeeComponent = {
    name: "employee",
    source: {
        image: "docker.io/celleryio/sampleapp-employee"
    },
    ingresses: {
        employee: new cellery:HTTPIngress(
                      8080,
                      "./resources/employee.swagger.json"
        )
    },
    parameters: {
        SALARY_HOST: new cellery:Env(),
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
        image: "docker.io/celleryio/sampleapp-salary"
    },
    ingresses: {
        SalaryAPI: new cellery:HTTPIngress(
                8080,

                {
                          basePath: "payment",
                          definitions:[{
                              path: "/",
                              method: "GET"
                          }]
                      }
            )
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
    cellery:setParameter(employeeComponent.parameters.SALARY_HOST, cellery:getHost(employeeCell, salaryComponent));
    cellery:setParameter(employeeComponent.parameters.BASE_PATH,
    cellery:getBasePath(salaryComponent.ingresses.SalaryAPI));

    // Add components to Cell
    employeeCell.addComponent(employeeComponent);
    employeeCell.addComponent(salaryComponent);

    //Expose API from Cell Gateway
    employeeCell.exposeAPIsFrom(employeeComponent);

    _ = cellery:createImage(employeeCell);
}

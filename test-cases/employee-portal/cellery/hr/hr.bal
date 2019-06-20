import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //HR component
    cellery:Component hrComponent = {
        name: "hr",
        source: {
            image: "wso2cellery/sampleapp-hr:0.3.0"
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
            cells: {
                employeeCellDep: "myorg/employee:1.0.0", //  fully qualified dependency image name as a string
                stockCellDep: <cellery:ImageName>{ org: "myorg", name: "stock", ver: "1.0.0" } // dependency as a struct
            }
        }
    };

    hrComponent.envVars = {
        employee_api_url: { value: <string>cellery:getReference(hrComponent, "employeeCellDep").employee_api_url },
        stock_api_url: { value: <string>cellery:getReference(hrComponent, "stockCellDep").stock_api_url }
    };

    // Cell Initialization
    cellery:CellImage hrCell = {
        components: {
            hrComp: hrComponent
        }
    };
    return cellery:createImage(hrCell, untaint iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    cellery:CellImage hrCell = check cellery:constructCellImage(untaint iName);
    return cellery:createInstance(hrCell, iName, instances);
}

// cellery test command will facilitate all flags as cellery run
public function test(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {

    cellery:ImageName[] instanceList = cellery:runInstances(iName, instances);
    string hr_cell_url = "http://" + iName.instanceName + "--gateway-service:80/hr";
    string emp_cell_url = "";
    string stock_cell_url = "";

    foreach var (k, v) in instances {
        if (k == "employeeCellDep") {
            emp_cell_url = "http://" + v.instanceName + "--gateway-service:80/employee";
        } else if (k == "stockCellDep") {
            stock_cell_url = "http://" + v.instanceName + "--gateway-service:80/stock";
        }
    }

    cellery:Test employeeExternalTest1 = {
        name: "hr-test1",
        source: {
            image: "docker.io/celleryio/sampleapp-test-hr"
        },
        envVars: {
            HR_CELL_URL: { value: hr_cell_url },
            EMP_CELL_URL: { value: emp_cell_url },
            STOCK_CELL_URL: { value: stock_cell_url }
        }
    };

    cellery:Test employeeExternalTest2 = {
        name: "hr-test2",
        source: {
            image: "docker.io/celleryio/sampleapp-test2-hr"
        },
        envVars: {
            EMP_CELL_URL: { value: emp_cell_url }
        }
    };

    cellery:TestSuite hrTestSuite = {
        tests: [employeeExternalTest1, employeeExternalTest2]
    };

    error? a = cellery:runTestSuite(iName, hrTestSuite);
    return cellery:stopInstances(iName, instanceList);
}

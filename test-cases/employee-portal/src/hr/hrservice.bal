import ballerina/http;
import ballerina/log;
import ballerina/internal;
import ballerina/config;
import ballerinax/docker;
import ballerina/io;

@final
string AUTH_HEADER = "Authorization";
@final
string BEARER_PREFIX = "Bearer";
@final
string JWT_SUB_TOKEN = "sub";
@final
string EMPLOYEE_NAME_HEADER = "x-emp-name";

endpoint http:Client employeeDetailsEp {
    url: config:getAsString("employee.api.url")
};

endpoint http:Client stockOptionsEp {
    url: config:getAsString("stock.api.url")
};

@http:ServiceConfig {
    basePath:"/"
}

@docker:Config {
    registry:"celleryio",
    name:"sampleapp-hr",
    tag:"v1.0"
}
service<http:Service> hr bind { port: 8080 } {

    @http:ResourceConfig {
        methods:["GET"],
        path:"/"
    }
    getHrDetails (endpoint caller, http:Request req) {
        http:Response res = new;

        string[] headers = req.getHeaderNames();
        foreach header in headers {
            io:println(header + ": " + req.getHeader(header));
        }

        string employeeName = req.getHeader("x-cellery-auth-subject");
        io:println("employee name: " + employeeName);

        json employeeDetails;
            match getEmployeeDetails(req) {
                http:Response response => {
                    match response.getJsonPayload() {
                        json jsonEmpDetails => {
                            employeeDetails = jsonEmpDetails;
                            io:println("employee details: ");
                            io:println(employeeDetails);
                        }
                        error e => {
                            log:printError("Error in extracting response from Employee service", err = e);
                            res.statusCode = 500;
                            caller->respond(res) but { error e1 => log:printError("Error sending response", err = e1) };
                            done;
                        }
                    }
                }
                error e => {
                    res.statusCode = 500;
                    caller->respond(res) but { error e1 => log:printError("Error sending response", err = e1) };
                    done;
                }
            }

        json stockOptionDetails;
            match getStockOptionData(req) {
                http:Response response => {
                    match response.getJsonPayload() {
                        json jsonStocks => {
                            stockOptionDetails = jsonStocks;
                            io:println("stock details: ");
                            io:println(stockOptionDetails);
                        }
                        error e => {
                            log:printError("Error in extracting response from Stock service", err = e);
                            res.statusCode = 500;
                            caller->respond(res) but { error e1 => log:printError("Error sending response", err = e1) };
                            done;
                        }
                    }
                }
                error e => {
                    res.statusCode = 500;
                    caller->respond(res) but { error e1 => log:printError("Error sending response", err = e1) };
                    done;
                }
            }
        json resp = buildResponse(employeeDetails, stockOptionDetails);
        io:println("response: ");
        io:println(resp);
        res.setJsonPayload(resp);

        caller->respond(res) but { error e => log:printError("Error sending response", err = e) };
    }
}


function buildResponse(json employeeDetails , json stockOptions)
returns json
{
    json response = {
        employee: { details: employeeDetails, stocks: stockOptions }
    };
    return response;

}

function getEmployeeDetails(http:Request clientRequest) returns http:Response|error {
    var response = employeeDetailsEp->get("/details", message = clientRequest);
    match response {
        http:Response httpResponse => {
            return httpResponse;
        }
        error e => {
            log:printError("Error in invoking EmployeeDetails service", err = e);
            return e;
        }
    }
}

function getStockOptionData(http:Request clientRequest) returns http:Response|error {
    var response = stockOptionsEp->get("/options", message = clientRequest);
    match response {
        http:Response httpResponse => {
            return httpResponse;
        }
        error e => {
            log:printError("Error in invoking StockOptions service", err = e);
            return e;
        }
    }
}

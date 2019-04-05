import ballerina/http;
import ballerina/log;
import ballerina/system;
import ballerina/io;
import ballerinax/docker;

map<string> employeeIdMap = { "bob": "0639", "alice": "0410", "jack": "0231", "peter": "0041" };
map<string> employeeDesignationMap = { "bob": "Software Engineer", "alice": "Senior Software Engineer", "jack": "Tech Lead", "peter": "Senior Tech Lead" };

@docker:Config {
    registry:"celleryio",
    name:"sampleapp-employee",
    tag:"v1.0"
}
@http:ServiceConfig {
    basePath:"/"
}
service<http:Service> employee bind { port: 8080 } {
    @http:ResourceConfig {
        methods: ["GET"],
        path:"/details"
    }
    details (endpoint caller, http:Request req) {
        //create new response
        http:Response res = new;

        string[] headers = req.getHeaderNames();
        foreach header in headers {
            io:println(header + ": " + req.getHeader(untaint header));
        }

        //check the header
        if (req.hasHeader("x-cellery-auth-subject")) {
            string empName = req.getHeader("x-cellery-auth-subject");
            log:printInfo(empName);
            if (employeeIdMap.hasKey(empName)) {
                json employeeIdJson = { id: employeeIdMap[empName], designation: employeeDesignationMap[empName] };

                string salaryServiceName = system:getEnv("SALARY");

                json salaryDetails = getSalaryDetails(empName, salaryServiceName, untaint req);
                employeeIdJson.salary = salaryDetails.salary;

                io:println(employeeIdJson);
                res.setJsonPayload(untaint employeeIdJson);
            } else {
                res.statusCode = 404;
                res.setContentType("application/json");
                res.setJsonPayload({});
            }
        } else {
            res.statusCode = 401;
            res.setContentType("application/json");
            res.setJsonPayload({});
        }
        caller->respond(res) but { error e => log:printError("Error sending response", err = e) };
    }
}

function getSalaryDetails(string empName, string salaryServiceName, http:Request req) returns (json) {
    endpoint http:Client clientEP {
        url: "http://" + salaryServiceName
    };

    json salary;
    var response = clientEP->get("/payroll/salary", message = req);

    match response {
        http:Response resp => {
            var msg = resp.getJsonPayload();
            match msg {
                json res => {
                    salary = res;
                }
                error err => log:printError(err.message);
            }
        }
        error err => log:printError(err.message);
    }
    return salary;
}

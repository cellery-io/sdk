import ballerina/http;
import ballerina/log;
import ballerinax/docker;
import ballerina/io;

map<string> employeeSalaryMap = { "bob": "$1000", "alice": "$1500", "jack": "$2000", "peter": "$2500" };

@docker:Config {
registry:"celleryio",
name:"sampleapp-salary",
tag:"v1.0"
}
service<http:Service> payroll bind { port: 8080 } {
    @http:ResourceConfig {
        methods: ["GET"]
    }
    salary (endpoint caller, http:Request req) {
        //create new response
        http:Response res = new;

        string[] headers = req.getHeaderNames();
        foreach header in headers {
            io:println(header + ": " + req.getHeader(header));
        }

        //check the header
        if (req.hasHeader("x-cellery-auth-subject")) {
            string empName = req.getHeader("x-cellery-auth-subject");
            if (employeeSalaryMap.hasKey(empName)) {
                json employeeIdJson = { salary: employeeSalaryMap[empName]};
                res.setJsonPayload(employeeIdJson);
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

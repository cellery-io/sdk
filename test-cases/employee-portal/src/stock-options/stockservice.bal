import ballerina/http;
import ballerina/log;
import ballerina/io;
import ballerinax/docker;

type StockRecord record {
    int total;
    int vestedAmount;
};
StockRecord stockRecord1 = {total: 100, vestedAmount: 90};
StockRecord stockRecord2 = {total: 120, vestedAmount: 105};
StockRecord stockRecord3 = {total: 200, vestedAmount: 160};
StockRecord stockRecord4 = {total: 75, vestedAmount: 55};
map stockOptionMap = { "bob" : stockRecord1 , "alice": stockRecord2, "jack": stockRecord3, "peter": stockRecord4 };

@docker:Config {
registry:"celleryio",
name:"sampleapp-stock",
tag:"v1.0"
}

@http:ServiceConfig {
    basePath:"/"
}

service<http:Service> stock bind { port: 8080 } {
    @http:ResourceConfig {
        methods: ["GET"]
    }
    options (endpoint caller, http:Request req) {
        http:Response res = new;

        string[] headers = req.getHeaderNames();
        foreach header in headers {
            io:println(header + ": " + req.getHeader(header));
        }

        //check the header
        if (req.hasHeader("x-cellery-auth-subject")) {
            string empName = req.getHeader("x-cellery-auth-subject");

            if (stockOptionMap.hasKey(empName)) {
                StockRecord stockRecord = check <StockRecord>stockOptionMap[empName];
                json stockResult = { options: { total: stockRecord.total, vestedAmount: stockRecord.vestedAmount} } ;    
                res.setJsonPayload(stockResult);
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
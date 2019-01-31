import ballerina/io;
import celleryio/cellery;


//HR component
cellery:Component hr = {
    name: "hr",
    source: {
        image: "docker.io/wso2vick/sampleapp-hr"
    },
    env: { employeegw_url: "", stockgw_url: "" },
    ingresses: {
        "hr": new cellery:HTTPIngress(8080, "info",
            [
                {
                    path: "/",
                    method: "GET"
                }
            ]
        )
    }

};

// Cell Intialization
cellery:CellImage hrCell = new("HR");
cellery:CellStub employeeStub = new("Employee");
cellery:CellStub stocksStub = new("Stock-Options");

public function build() {
    // Build HR cell
    io:println("Building HR Cell ...");
    hrCell.addComponent(hr);

    // Load Employee Stub
    employeeStub.addIngress("employee");
    hrCell.addStub(employeeStub);

    // Load Stock Stub
    stocksStub.addIngress("stocks");
    hrCell.addStub(stocksStub);

    // Expose API from Cell Gateway & Global Gateway
    hrCell.exposeGlobalAPI(hr);
    hrCell.declareEgress(employeeStub.name, employeeStub.getIngress("employee"), "employeegw_url");
    hrCell.declareEgress(stocksStub.name, stocksStub.getIngress("stocks"), "stockgw_url");
    _ = cellery:createImage(hrCell);

}

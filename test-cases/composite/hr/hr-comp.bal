import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    cellery:Component hrCompComponent = {
        name: "hr",
        source: {
            image: "wso2cellery/sampleapp-hr:0.3.0"
        },
        ingresses: {
            http:<cellery:HttpsPortIngress>{port: 8080}
        },
        dependencies: {
            composites: {
                //  fully qualified dependency image name as a string
                employeeCompDep: "myorg/employee-comp:1.0.0",
                // dependency as a struct
                stockCompDep: <cellery:ImageName>{ org: "myorg", name: "stock-comp", ver: "1.0.0" }
            }
        },
        labels: {
            team: "Finance",
            owner: "Alice"
        }
    };

    cellery:Reference empRef = cellery:getReference(hrCompComponent, "employeeCompDep");
    string empURL = "http://" +<string>empRef.employee_host + ":" +<string>empRef.employee_port;

    cellery:Reference stockRef = cellery:getReference(hrCompComponent, "stockCompDep");
    string stockURL = "http://" +<string>stockRef.stock_host + ":" +<string>stockRef.stock_port;

    hrCompComponent.envVars = {
        employee_api_url: {
            value: empURL
        },
        stock_api_url: {
            value: stockURL
        }
    };

    // Declare the composite
    cellery:Composite hrComposite = {
        components: {
            hrComp: hrCompComponent
        }
    };
    return cellery:createImage(hrComposite, untaint iName);// this will create an image with type = composite
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:Composite hrComposite = check cellery:constructImage(untaint iName);
    return cellery:createInstance(hrComposite, iName, instances, startDependencies, shareDependencies);
}



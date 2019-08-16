import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    cellery:Component hrCompComponent = {
        name: "hr",
        source: {
            image: "docker.io/celleryio/sampleapp-hr"
        },
        ingresses: {
            http:<cellery:HttpsPortIngress>{port: 8080}
        },
        dependencies: {
            composites: {
                //  fully qualified dependency image name as a string
                employeeCellDep: "myorg/employee-comp:1.0.0",
                // dependency as a struct
                stockCellDep: <cellery:ImageName>{ org: "myorg", name: "stock-comp", ver: "1.0.0" }
            }
        },
        labels: {
            team: "Finance",
            owner: "Alice"
        }
    };

    hrCompComponent.envVars = {
        employee_api_url: {
            value: <string>cellery:getReference(hrCompComponent, "employeeCellDep").employee_host
        },
        stock_api_url: {
            value: <string>cellery:getReference(hrCompComponent, "stockCellDep").stock_host
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

public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    cellery:Composite hrComposite = check cellery:constructImage(untaint iName);
    return cellery:createInstance(hrComposite, untaint iName, instances);
}



import ballerina/io;
import celleryio/cellery;

//Pet Component
cellery:Component petComponent = {
    name: "pet-service",
    source: {
        image: "docker.io/isurulucky/pet-service"
    },
    ingresses: {
        stock: new cellery:HTTPIngress(9090,
            "petsvc",
            [
                {
                    path: "/*",
                    method: "GET"
                }
            ]
        )
    },
    autoscaling: {
        policy: {
            minReplicas: 1,
            maxReplicas: 10,
            cpuPercentage: new cellery:CpuUtilizationPercentage(50)
        }

    }
};


//Pet Component
cellery:Component debugComponent = {
    name: "debug",
    source: {
        image: "docker.io/mirage20/k8s-debug-tools"
    }
};

cellery:CellImage petCell = new();

public function build(string orgName, string imageName, string imageVersion) {
    //Build Pet Cell
    io:println("Building Pet Service Cell ...");
    petCell.addComponent(petComponent);
    petCell.addComponent(debugComponent);
    //Expose API from Cell Gateway
    petCell.exposeGlobalAPI(petComponent);
    _ = cellery:createImage(petCell, orgName, imageName, imageVersion);
}

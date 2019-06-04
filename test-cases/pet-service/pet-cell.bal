import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //Pet Component
    cellery:Component petComponent = {
        name: "pet-service",
        source: {
            image: "docker.io/isurulucky/pet-service"
        },
        ingresses: {
            stock: <cellery:HttpApiIngress>{ port: 9090,
                context: "petsvc",
                definition: {
                    resources: [
                        {
                            path: "/*",
                            method: "GET"
                        }
                    ]
                }
            }
        },
        autoscaling: {
            policy: {
                minReplicas: 1,
                maxReplicas: 10,
                cpuPercentage: <cellery:CpuUtilizationPercentage>{ percentage: 50 }
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

    cellery:CellImage petCell = {
        components:{
            petComp:petComponent,
            debugComp:debugComponent
        }
    };
    return cellery:createImage(petCell, untaint iName);
}

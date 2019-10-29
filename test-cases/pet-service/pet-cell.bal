import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //Pet Component
    cellery:Component petComponent = {
        name: "pet-service",
        src: {
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
        scalingPolicy: <cellery:ZeroScalingPolicy> {
            maxReplicas: 10,
            concurrencyTarget: 25
        }

    };


    //Pet Component
    cellery:Component debugComponent = {
        name: "debug",
        src: {
            image: "docker.io/mirage20/k8s-debug-tools"
        },
        scalingPolicy: <cellery:AutoScalingPolicy> {
            minReplicas: 1,
            maxReplicas: 10,
            metrics: {
                cpu: <cellery:Value>{ threshold : "500m" },
                memory: <cellery:Percentage> { threshold : 50 }
            }
        }
    };

    cellery:CellImage petCell = {
        components: {
            petComp: petComponent,
            debugComp: debugComponent
        }
    };
    return <@untainted> cellery:createImage(petCell, iName);
}

// The Cellery Lifecycle Run method which is invoked for creating a Cell Instance.
//
// iName - The Image name
// instances - The map dependency instances of the Cell instance to be created
// return - The Cell instance
public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:CellImage petCell = check cellery:constructCellImage(iName);
    return <@untainted> cellery:createInstance(petCell, iName, instances, startDependencies, shareDependencies);
}
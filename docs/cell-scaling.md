# Scaling up/down cells
Each component within the cells can be scaled up or down. Cellery supports auto scaling with [Horizontal pod autoscaler](#scale-with-hpa), 
and [Zero-scaling](#zero-scaling).

This README includes,
* [Horizontal pod autoscaler](#scale-with-hpa)
    * [Enable HPA](#enable-hpa)
    * [Syntax](#syntax-for-scaling-with-hpa)
    * [Export Policy](#export-policy)
    * [Apply Policcy](#apply-policy)
* [Zero-scaling](#zero-scaling)
    * [Enable zero-scaling](#enable-zero-scaling)
    * [Syntax](#syntax-for-zero-scaling)

A [component](https://github.com/wso2-cellery/spec/tree/master#component) can have either Autoscaling policy or zero scaling policy. Based on that, the underneath autoscaler will be determined. 
Generally the autoscaling policy has minimum replica count greater than 0, and hence the [horizontal pod autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
will be used to scale up and down the components. The zero-scaling has minimum replica count `0`, and hence when the 
component did not get any request, the component will be terminated and it will be running back once a request was directed to the component. 

## Scale with HPA
If a component has a scaling policy as Autoscaling policy as explained [here](cellery-syntax.md#autoscaling), the specified 
component will be scaled with [horizontal pod autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/). 

### Enable HPA
By default [local](setup/local-setup.md), and [existing kubernetes cluster](setup/existing-cluster.md) will not have the 
[metrics server](https://github.com/kubernetes-incubator/metrics-server) deployed for 
[horizontal pod autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) to work. [GCP](setup/gcp-setup.md) has it by default.
Therefore, if you are using [local](setup/local-setup.md), or [existing kubernetes cluster](setup/existing-cluster.md), 
then you can enable by following below command. This also can be performed by [modify setup](setup/modify-setup.md).

```bash
    cellery setup modify --hpa=enable
```

### Syntax for scaling with HPA
```ballerina
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
        resources: {
            limits: {
                memory: "128Mi",
                cpu: "500m"
            }
        },
        scaling: {
            policy: <AutoScalingPolicy> {
                minReplicas: 1,
                maxReplicas: 10,
                metrics: {
                    cpu: <cellery:Value>{ threshold : "500m" },
                    memory: <cellery:Percentage> { threshold : 50 }
                }
            },
            override: true
        }
    };
```

The above component `pet-service` has an autoscaling policy with minimum replica count `1` and maximum replica count `10`. 
Further, threshold values for cpu and memory is provided to decide upong scaling decisions. For detailed syntax, refer [here](cellery-syntax.md).

### Export policy
Once the above component is wrapped into a cell and deployed in Cellery runtime, the build time auto scaling policy will be applied. 
Often the build time autoscaling policy is not sufficient, and depends on the runtime environment and available resources, 
the devops may require to re-evaluate the autoscaling policies. Therefore, the exporting policy will be helpful to understand the 
current autoscaling policy that is applied to the component. This can be performed as shown below.
```bash
    cellery export-policy autoscale pet-be-instance -f pet-be-policy.yaml
```
  
### Apply policy
Once the policy is [exported](#export-policy), the policy can be evaluated, and modified based on the requirement. The 
modified policy can be applied to the running cell instance. Note, if the component is defined with scaling policy with 
`override` parameter `false` as shown [syntax](#syntax), the policy cannot be applied in the runtime. 
```bash
 cellery apply-policy autoscale pet-fe-modified.yaml pet-be
```

You also can selectively apply the autoscaling policy to selected components via below command. 
```bash
 cellery apply-policy autoscale pet-fe-modified.yaml pet-be -c controller, catalog
```

## Zero-scaling
Zero scaling is powered by [Knative](https://knative.dev/v0.6-docs/). The zero-scaling have minimum replica count 0, and hence when the 
component did not get any request, the component will be terminated and it will be running back once a request was directed to the component. 

### Enable zero-scaling
By default, the cellery installations have zero-scaling disabled. Therefore, if you want to zero scale your components, 
you have to enable zero-scaling as shown below.
```bash
 cellery setup modify --scale-to-zero=enable
``` 

### Syntax for zero-scaling
```ballerina
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
        scaling: {
            policy: <ZeroScalingPolicy> {
                maxReplicas: 10,
                concurrencyTarget: 50
            }
        }
    };
```
The above component `pet-service` has a zero scaling policy (minimum replica count `0`) with maximum replica count `10`. 
Further, scaling with be performed based on concurrent requests for the component, and `pet-service` has concurrency threshold `50`.
For detailed syntax, refer [here](cellery-syntax.md).  

Just [build and run the cell](writing-a-cell.md) with this scaling configuration for the component, to try out the zero scaling functionality. 

#What's Next?
- [Developing and runing a Cell](writing-a-cell.md) - step by step explanation on how you could define your own cells.
- [How to code cells?](cellery-syntax.md) - explains how Cellery cells are written.
- [Update cells](cell-update-and-adv-deployment.md) - update a running cell instance using a new cell version.
- [Observe cells](cellery-observability.md) - provides the runtime insight of cells and components.


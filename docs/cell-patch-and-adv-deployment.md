# Component Patching and Advanced Deployment Patterns

Cellery supports in-place patching of components in running cells/composites as well as advanced deployment patterns Blue-Green and Canary. 

This README includes,

- [Patching Components](#patching-components)
- [Blue/Green Deployment of Cell/Composite Instances](#bluegreen-and-canary-deployment-of-cell-instances)
- [Canary Deployment of Cell/Composite Instances](#bluegreen-and-canary-deployment-of-cell-instances)

## Patching Components
Components of a running cell/composite instance can be patched by this method. This will terminate the selected component and spin up a new updated component. This is an in-place update mechanism.

#### Note:
__This updating mechanism only considers changes done to docker images which are encapsulated in components.__
 
Let us assume, there is a cell/composite `pet-be` running in the current runtime, created from the cell image `wso2cellery/pet-be-cell:1.0.0`. 
And now, some changes are made to the application source and hence the users will have to update the running instance with those changes.
The new application binary is packed to a docker image with a patch version. 

As this operation simply updates the currently running instance, any client cells/composites invoking this instance may not be aware of this. 
Therefore, this should be only permitted when there is no API changes. 
 
Below steps should be followed to perform the patch operation for a component in a running cell/composite. 
 
1) Update the component `controller` of the currently running `pet-be` instance with below command. This will update its components respectively.
```
cellery update pet-be controller --container-image mycontainerorg/hello:1.0.1
```
3) Now execute `kubectl get pods` and you can see the pods of the `pet-be` are getting initialized. And finally, older pods are getting terminated.
```
$ kubectl get pods
NAME                                             READY   STATUS            RESTARTS   AGE
pet-be--catalog-deployment-54b8cd64-knhnc        0/2     PodInitializing   0          4s
pet-be--catalog-deployment-67b8565469-fq86w      2/2     Running           0          26m
pet-be--controller-deployment-6f89fdb47c-rn4mn   2/2     Running           0          24m
pet-be--controller-deployment-75f5db95f4-2dt96   0/2     PodInitializing   0          4s
pet-be--customers-deployment-7997974649-22hft    2/2     Running           0          26m
pet-be--customers-deployment-7d8df7fb84-h48xs    0/2     PodInitializing   0          4s
pet-be--gateway-deployment-7f787575c6-vmg4p      2/2     Running           0          26m
pet-be--orders-deployment-7d874dfd98-vnhdw       0/2     PodInitializing   0          4s
pet-be--orders-deployment-7d9fd8f5ff-4czdx       2/2     Running           0          26m
pet-be--sts-deployment-7f4f56b5d5-bjhww          3/3     Running           0          26m
pet-fe--gateway-deployment-67ccf688fb-dnhhw      2/2     Running           0          4h6m
pet-fe--portal-deployment-69bb57c466-25nqd       2/2     Running           0          4h6m
pet-fe--sts-deployment-59dbb995c7-g7tc7          3/3     Running           0          4h6m
```
Refer to [CLI docs](cli-reference.md#cellery-update) for a complete guide on performing updates on cell instances.

## Blue/Green and Canary Deployment of Cell/Composite Instances
Blue-Green and Canary are advanced deployment patterns which can used to perform updates to running cell/composite instances. 
However, in contrast to the component update method described above, this update does not happen in place and a new instance needs to be used to re-route traffic explicitly. 
The traffic can be either switched 100% (Blue-Green method) or partially (Canary method) to a instance created with a new cell/composite image. 

Let us assume that the `pet-be` is an instance of `wso2cellery/pet-be-cell:1.0.0`, and we are planning to switch traffic to a new cell instance created from the image ` wso2cellery/pet-be-cell:2.0.0`.
Therefore, as a first step a new pet-be instance `pet-be-v2` should be started. Canary deployment can be achieved by having 50% traffic routed to the `pet-be` and `pet-be-v2` 
cell instances. Then, we can  completely switch 100% traffic to `pet-be-v2` and still have the both cell instances running as per the blue-green deployment pattern. Finally, terminate `pet-be`.

- Route the 50% of the traffic to the new `pet-be-v2` cell instance. 
```
$ cellery route-traffic pet-be -p pet-be-v2=50
```

- The traffic can be completely switched to 100% to the `pet-be-v2` as shown below. 
```
$ cellery route-traffic pet-be -p pet-be-v2=100
```
- Terminate the old instance `pet-be` (once all clients have switched to the new instance), and only have the `pet-be-v2` cell running. 
```
cellery terminate pet-be
```
- Additionally, you can use the flag `--enable-session-awareness` to enable session aware routing based on the user name. 
An instance will be selected and will be propagated via the header `x-instance-id`. Its the cell component's responsibility to forward this header if this option is to be used.

#### Note:
The above commands will apply to all cell/composite instances which has a dependency on `pet-be`. If required, route-traffic command can be applied to only a selected set of instances
using the `-s/--source` option:
```
$ cellery route-traffic -s pet-fe pet-be -p pet-be-v2=50
```
Refer to [CLI docs](cli-reference.md#cellery-route-traffic) for a complete guide on managing advanced deployments with cell instances.

## Try with sample
[Pet-store application sample](https://github.com/wso2-cellery/samples/tree/master/cells/pet-store) walks through the cell upate scenario. 
Find more information on the steps [here](https://github.com/wso2-cellery/samples/blob/master/docs/pet-store/update-cell.md).

# What's Next?
- [Developing and runing a Cell](writing-a-cell.md) - step by step explanation on how you could define your own cells.
- [CLI Commands](cli-reference.md) - reference for CLI commands.
- [How to code cells?](cellery-syntax.md) - explains how Cellery cells are written.
- [Scale up/down](cell-scaling.md) - scalability of running cell instances with zero scaling and horizontal autoscaler.
- [Observe cells](cellery-observability.md) - provides the runtime insight of cells and components.


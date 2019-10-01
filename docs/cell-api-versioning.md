# Cell Api Versioning

A cell defines a standard contract for the external parties to communicate with it. Hence, it is accessible only via 
the cell gateway. Each http API exposed via the cell gateway can optionally specify a version. 

[Cli List Ingress](cli-reference.md#cellery-list-ingresses) command can be used to find the API versions of HTTP ingresses, 
either for a cell image or a running cell instance. 

# Exposing Apis for External Access

Cell APIs can be selectively published to the Cellery global gateway to be accessed by external clients. This is done by 
marking the relevant HTTP ingress as `expose: global` and by using a global API publisher configuration. 

If a set of cell APIs are to be exposed globally each with a unique context, each ingress should be marked as to be exposed 
as global.
Ex.: 
```ballerina
cellery:HttpApiIngress helloAPI = {
    // ...
    expose: "global"
};
```
 
If cell APIs are to be exposed with a common root context, the global publisher configuration can be used to specify it.  
Each context specified in the HTTP ingress will be used as a sub context of the root context specified in the publisher. 
Ex.:
```ballerina
cellery:CellImage hrCell = {
    globalPublisher: {
        apiVersion: "1.0.1",
        context: "myorg"
    },
    // ...
};
```

Refer [Exposing Ingresses](cellery-syntax.md#expose) for further details and syntax. 

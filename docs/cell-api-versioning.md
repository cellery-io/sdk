# Cell Api Versioning

A cell defines a standard contract for the external parties to communicate with it. Hence, a it is accessible only via 
the cell gateway. Each http API exposed via the cell gateway can optionally specify a version. 

[Cli List Ingress](cli-reference.md#cellery-list-ingresses) command can be used to find the API versions of HTTP ingresses, 
either for a cell image or a running cell instance. 

# Exposing Apis for External Access

Cell APIs can be selectively published to the Cellery global gateway to be accessed by external clients. This is done by 
marking the relevant HTTP ingress as global, or by using a global API publisher configuration. 

If only a selective set of cell APIs are to be exposed globally, each relevant ingress should be marked as global. If all 
cell APIs are to be exposed, the global publisher configuration can be used to specify it. The global publisher configuration 
provides the capability to use a single root context and a version for all published APIs. Each context specified in the 
HTTP ingress will be used as a sub context of the root context specified in the publisher. Refer 
[Exposing Ingresses](cellery-syntax.md#expose) for further details and syntax. 

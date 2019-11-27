# Kubernetes Resources Overview

This document provides a description of resources deployed to your
Kubernetes cluster when running a cell or composite. Cellery uses following CustomResourceDefinitions under `mesh.cellery.io` API group 
in order to create cells and composites.

* [Cell](#Cell)
* [Composite](#Composite)
* [Component](#Component)
* [Gateway](#Gateway)
* [TokenService](#TokenService)

Run `kubectl get crds | grep cellery` to check whether those resource are present in your cluster. 

## Cell

Cell resource declares an individual cell running inside the cluster which contains set of components exposed via a gateway. The incoming traffic to a cell is only allowed from the gateway. 
The cell spec has following major configurations, 

* gateway: [Gateway](#Gateway) template for specifying the cell gateway
* components: List of [Component](#Component) templates 
* sts: [TokenService](#TokenService) template for defining security policies

See [Todo-API](https://github.com/wso2/cellery-controller/blob/master/samples/todo-api/todos.yaml) for example configuration

## Composite

Composite resource declares a set of components which does not have any restrictions as in Cells. 
The composite spec has following configuration, 

* components: List of [Component](#Component) templates 


## Component

Component is the smallest deployable unit in Cellery which is mapped to a Microservice.
The component spec has following major configurations, 

* type: Describes the runtime behaviour of this component
* template: Standard Kubernetes pod template
* ports: List of port mappings which is used for exposing the containers via a service
* scalingPolicy: Describes autoscaling configurations for this component
* volumeClaims: List of persistent volume claims used by this component 
* configurations: List of configurations used by this component  
* secrets: List of secrets used by this component

Example component configuration with all parameters

```yaml
apiVersion: mesh.cellery.io/v1alpha2
kind: Component
metadata:
  name: component1
spec:
  # +optional k8s resource this component is matched to. Acceptable values are Deployment, Job, StatefulSet.
  # Default to Deployment if not specified
  type: StatefulSet
  # +optional scaling configuration for this component. If nothing specified, the system specify the default replicas
  # based on 'default-replicas' system configuration. If all multiple autoscaling is specified, priority is given in
  # following order,
  # 1. hpa
  # 2. kpa
  # 3. replicas
  scalingPolicy:
    # +optional set the fixed number of replicas without any autoscaling
    replicas: 1
    # +optional use k8s standard horizontal pod autoscaler
    hpa:
      # +optional specify whether the scaling policy can be overridden or not
      overridable: true|false
      minReplicas: 1
      maxReplicas: 10
      # standard k8s hpa metrics
      metrics:
      - type: Resource
        resource:
          name: cpu
          target:
            type: Utilization
            averageUtilization: 50
    # +optional use knative pod autoscaler which supports scale-to-zero
    kpa:
      minReplicas: 0
      maxReplicas: 5
      # max request concurrency per replica
      concurrency: 10
  # pod template for this component. all undocumented fields have their standard meanings.
  # see https://github.com/kubernetes/kubernetes/blob/v1.14.6/pkg/apis/core/types.go#L2569
  template:
    containers:
      # +optional unique name to identify the container.
    - name: container-core
      image: docker.io/celleryio/core
      # ports specified in this section are override by $.spec.components[n].spec.ports.
      # please use $.spec.components[n].spec.ports to specify the port mapping.
      ports:
      - containerPort: 8080
      - containerPort: 8081
      # list of volume mounts for this container.
      volumeMounts:
        # name of the volume which must match one of the following
        # 1. volumes specified in $.spec.components[n].spec.template.volumes
        # 2. volumeClaims specified in $.spec.components[n].spec.volumeClaims
        # 3. configurations specified in $.spec.components[n].spec.configurations
        # 4. secrets specified in $.spec.components[n].spec.secrets
      - name: vol1
        # mount path to the container
        mountPath: /etc/config
      - name: pvc-data
        mountPath: /etc/data
    - name: container-sidecar
      image: docker.io/celleryio/sidecar
    # specify a volume which can be access by the containers. use this configuration if you want to specify volumes
    # other than persistent volumes,configmaps and secrets
    # 1. use $.spec.components[n].spec.volumeClaims to specify volumeClaims
    # 2. use $.spec.components[n].spec.configurations to specify configurations
    # 3. use $.spec.components[n].spec.secrets to specify secrets
    volumes:
    - name: vol1
      emptyDir: {}
  # list of port mappings which exposes container ports via the component service
  ports:
    # +optional unique name to identify the port mapping. A generated name in following format
    # will use if not specified.
    # <protocol>-<port>-<targetPort>
  - name: api
    # +optional protocol exposed by this port which can be one of HTTP, GRPC, TCP. Defaults to TCP if not specified.
    protocol: HTTP
    # exposed port from the component service
    port: 80
    # +optional container name where the port is exposed. If not specified, the mapping will apply to container[0]
    targetContainer: hr-core
    # target port of the container
    targetPort: 8080
  - name: stream
    protocol: GRPC
    port: 8081
    targetContainer: hr-core
    targetPort: 8081
  # list of volumeClaims required by the component
  volumeClaims:
    # unique name to identify the volume claim. This name can be used for accessing this volume from the component volumeMounts
  - name: pvc-data
    # specify whether this volume claim needs to be shared between replicas. This parameter is only works for StatefulSets
    shared: false
    # template for the volumeClaim
    template:
      metadata:
        name: pvc-data
      spec:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName: local-storage
        resources:
          requests:
            storage: 2Gi
  # list of configurations required by the component. Each configuration get mapped to a configmap
  configurations:
  - metadata:
      # unique name to identify the configuration. This name can be used for accessing this configuration from the component volumeMounts
      name: config1
    data:
      # key value pairs specifying each configurations
      key1: value1
      key2: |
        config file with
        multiple lines
  # list of secrets required by the component.
  secrets:
  - metadata:
      # unique name to identify the secret. This name can be used for accessing this secret from the component volumeMounts
      name: secret1
    data:
      # key value pairs specifying secrets. Secrets can be specified in three different ways.
      # 1. Plain text mode. Example: key1
      # 2. Base64 encoded mode. Example: key2
      # 3. Encrypted mode. This mode uses cellery key-pair to decrypt the secret. Example: key3
      key1: value1
      key2: "@base64:LS0tLS1CRUdJ=="
      key3: "@encrypted:AQCSiNN87la=="
```
 

## Gateway


Gateway declares a cell gateway which is used for specifying routing to the components.

Example gateway configuration with all parameters

```yaml
apiVersion: mesh.cellery.io/v1alpha2
kind: Gateway
metadata:
  name: gateway1
spec:
  # ingress configuration of the gateway
  ingress:
    # +optional specify the built-in extensions which adds different functionality to the ingress gateway
    extensions:
      # +optional publish http api's to the global gateway.
      apiPublisher:
        # +optional enable authentication of the global API
        authenticate: false
        # service name of the backend
        backend: backend1
        # context for the global API
        context: context1
        # +optional semantic version of the published api in MAJOR.MINOR format. if not specified v0.1 is used
        version: v1.0
      # +optional create k8s cluster ingress exposing this gateway
      clusterIngress:
        host: "foo.com"
        tls:
          secret: ing-secret
          key: "@encrypted:AQCSiNN87la=="
          cert: "@base64:LS0tLS1CRUdJ=="
      # +optional enable open id connect for the gateway
      oidc: ...
    # +optional list of route rules for exposing HTTP traffic
    http:
      # defines the base path of the url. The routing rule uses this context to identify the backend and rewrite the
      # base path with '/'. For example, http://<host>:<port>/context1/foo/bar?q=true will
      # rewrite to http://<host>:<port>/foo/bar?q=true
    - context: context1
      # +api version for the cell gateway
      version: 
      # +optional definitions of the api's
      definitions:
      - path: /
        method: GET
      destination:
        # component name of the backend
        host: my-component1
      # exposed http port of the backend
        port: 8080
      # +optional enable authentication on the published api
      authenticate: true
      # +optional enable published api as global
      global: true
    # +optional list of route rules for exposing GRPC traffic
    grpc:
      # opens a port with specified port number where the gateway is accepting GRPC traffic
    - port: 8080
      destination:
        # component name of the backend
        host: my-component1
        # exposed grpc port of the backend
        port: 8080
    # +optional list of route rules for exposing TCP traffic
    tcp:
      # opens a port with specified port number where the gateway is accepting TCP traffic
    - port: 8080
      destination:
        # component name of the backend
        host: my-component1
        # exposed tcp port of the backend
        port: 8080
  scalingPolicy:
    # +optional set the fixed number of replicas without any autoscaling
    replicas: 1
    # +optional use k8s standard horizontal pod autoscaler
    hpa:
      # +optional specify whether the scaling policy can be overridden or not
      overridable: true|false
      minReplicas: 1
      maxReplicas: 10
      # standard k8s hpa metrics
      metrics:
      - type: Resource
        resource:
          name: cpu
          target:
            type: Utilization
            averageUtilization: 50

```
 


## TokenService

TokenService declares a security policies for a cell which is used for token management and defining OPA policies.

Example tokenservice configuration with all parameters


```yaml
apiVersion: mesh.cellery.io/v1alpha2
kind: TokenService
metadata:
  name: tokenservice1
spec:
  # list of paths which are ignored by the token service
  unsecuredPaths: ["/foo", "/bar"]
  # list of opa policies required for the cell
  opa:
    # unique key for the policy
    key: sample1
    # policy specification in Rego query language
    regoPolicy: |
      package cellery.io

      default allow = true


```
 

## Cellery CLI Commands

* [setup](#cellery-setup) - create/manage cellery runtime installations.
* [init](#cellery-init) - initialize a cellery project.
* [build](#cellery-build) - build a cell image.
* [run](#cellery-run) - create cell instance(s). 
* [view](#cellery-view) - view cell and component dependencies.
* [list](#cellery-list) - list information about cell instances/images.
* [delete](#cellery-delete) - Delete cell images.
* [login](#cellery-login) - login to cell image repository.
* [push](#cellery-push) - push a built image to cell image repository.
* [pull](#cellery-pull) - pull an image from cell image repository.
* [terminate](#cellery-terminate) - terminate a cell instance.
* [status](#cellery-status) - check status of cell instance.
* [logs](#cellery-logs) - display logs of one/all components of a cell instance.
* [inspect](#cellery-inspect) - list the files included in a cell image. 
* [extract-resources](#cellery-extract-resources) - extract packed resources in a cell image.
* [update](#cellery-update) - perform a patch update on a particular cell instance.
* [route-traffic](#cellery-route-traffic) - route a percentage of traffic to a new cell instance.
* [export-policy](#cellery-export-policy) - export a policy from cellery run time.
* [apply-policy](#cellery-apply-policy) - apply a policy to a cellery instance.

#### Cellery Setup
Cellery setup command install and manage cellery runtimes. For this purpose it supports several sub commands. Please 
refer the [setup command readme](cli-setup-command.md) for complete instructions.

##### Cellery setup status:

Display status of cluster with a status list of system components

Ex:

 ```
    cellery setup status
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery Init

This will initialize a new cellery project in the current directory with the given name which includes an auto-generated cell definition. 
The project name also can be provided as inline param which will initialize the cellery project as given.

Ex:

 ```
    cellery init 
    cellery init my-first-project
 ```
 
[Back to Command List](#cellery-cli-commands)

#### Cellery Build

Build an immutable cell image.

###### Parameters: 

* _Cell file: The .bal which has the cell definition_
* _Cell image name: This is the image name, and it should be in format 
<ORGANIZATION_NAME>/<IMAGE_NAME>:\<VERSION>_

Ex: 

 ```
    cellery build my-project.bal wso2/my-cell:1.0.0
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery Run

Create a running instance from a cell image.

###### Parameters: 

* Cell image name: name of a built Cell image

###### Flags (Optional): 

* _-y, --assume-yes : Flag to enable/disable prompting for confirmation before starting instance(s)_
* _-e, --env : Set an environment variable for the cellery run method in the Cell file_
* _-l, --link : Link an instance with a dependency alias_
* _-n, --name : Name of the cell instance_
* _-s, --share-instances : Share all instances among equivalent Cell Instances_
* _-d, --start-dependencies : Start all the dependencies of this Cell Image in order_

Ex: 

 ```
    cellery run wso2/my-cell:1.0.0
    cellery run wso2/my-cell:1.0.0 -n my-cell-inst 
    cellery run wso2/my-cell:1.0.0 -l dependencyKey:dependentInstance
    cellery run wso2/my-cell:1.0.0 -e config=value 
    cellery run wso2/my-cell:1.0.0 -d 
    cellery run wso2/my-cell:1.0.0 -s -d
    cellery run wso2/my-cell:1.0.0 -y
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery View

View the cell image with inter component dependencies, and inter cell dependencies.

 ```
    cellery view wso2/my-cell:1.0.0
 ```
[Back to Command List](#cellery-cli-commands)

#### Cellery List

List running instances/cell images. This command can take four forms, for listing components, images, ingresses or instances.

##### Cellery List Components:

List the components which the cell image/instance encapsulate.

###### Parameters: 

* _cell instance name/cell image name: Either the cell instance name and cell image name._

Ex:
 ```
   cellery list components my-cell-inst
   cellery list components cellery-samples/employee:1.0.0
 ```
 
##### Cellery List Images

Lists the available cell images.

Ex:
 ```
   cellery list images
 ```

##### Cellery List Ingresses

List the exposed APIs of a cell image/instance.

###### Parameters: 

* _cell instance name/cell image name: Either the cell instance name and cell image name._

Ex: 
 ```
   cellery list ingresses my-cell-inst
   cellery list ingresses cellery-samples/employee:1.0.0
 ```

###### Cellery List Instances

List all running cells.

Ex:

 ```
    cellery list instances 
 ```

[Back to Command List](#cellery-cli-commands)

###### Cellery List dependencies

List the dependency information for a given cell instance. The dependency cell image, version and the cell instance name
will be provided as the output. 

###### Parameters: 

* _cell instance name: A valid cell instance name._

Ex:

 ```
    cellery list dependencies my-cell-inst 
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery delete

Delete cell images. This command will delete one or more cell images from cellery local repository. Users can also delete all cell images by executing the command with "--all" flag.

##### Parameters:

* _cell image names: This is a space separated list of existing cell images and/or regular expression of cell images._

Ex:
 ```
   cellery delete cellery-samples/employee:1.0.0 cellery-samples/hr:1.0.0
   cellery delete --regex 'cellery-samples/.*:1.0.0'
   cellery delete cellery-samples/employee:1.0.0 --regex '.*/employee:.*'
   cellery delete --all
 ```

 [Back to Command List](#cellery-cli-commands)

#### Cellery login

Log in the user to the cellery image repository, which is docker hub, and caches the credentials in the key ring in their machine, therefore user doesn't need to repeat typing the credentials.

Ex: 

 ```
    cellery login
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery Push

Push the cell image to the docker hub user account.

###### Parameters:

* _cell image name: This is the image name, and it should be in format <ORGANIZATION_NAME>/<IMAGE_NAME>:\<VERSION>_

Ex:

 ```
    cellery push wso2/my-cell:1.0.0
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery Pull

Pull the cell image from docker registry and include in the cellery local repository.

###### Parameters:

* _cell image name: This is the image name, and it should be in format <ORGANIZATION_NAME>/<IMAGE_NAME>:\<VERSION>_

Ex: 
 ```
   cellery pull wso2/my-cell:1.0.0
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery Terminate

Terminate running cell instances within cell runtime.

###### Parameters:

* _cell instance names: Names of the instances running in the cellery system_

Ex: 
 ```
   cellery terminate employee
   cellery terminate pet-fe pet-be
   cellery terminate --all
 ```
 
 [Back to Command List](#cellery-cli-commands)

#### Cellery Status

Check for the runtime status of the cell instance.

###### Parameters:

* _cell instance name: Name of the instance running in the cellery system_

Ex: 
 ```
   cellery status my-cell-inst
 ```
 
[Back to Command List](#cellery-cli-commands)

#### Cellery Logs

Fetch logs of all components or specific component within the cell instance and print in the console.

###### Parameters:

* _cell instance name: Name of the instance running in the cellery system_

###### Flags (Optional):

* _-c, --component: Name of the component of which the logs are required_

Ex: 
 ```
   cellery logs my-cell-inst 
   cellery logs my-cell-inst -c my-comp
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery Inspect

List the files included in a cell image.

###### Parameters:

* _cell image name: This is the image name, and it should be in format <ORGANIZATION_NAME>/<IMAGE_NAME>:\<VERSION>_

Ex:
 ```
   cellery inspect wso2/my-cell:1.0.0
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery Extract Resources

This will extract the resources folder of the cell image. This is useful to see the swagger definitions of the cell APIs therefore users can generate client code to invoke the cell APIs.

###### Parameters:

* _cell image name: This is the image name, and it should be in format <ORGANIZATION_NAME>/<IMAGE_NAME>:\<VERSION>_

###### Flags (Optional):

* _-o, --output: The directory into which the resources should be extracted_

Ex:
 ```
   cellery extract-resources wso2/my-cell:1.0.0 
   cellery extract-resources wso2/my-cell:1.0.0 -o /my/output/resource
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery Update

Perform a patch update on a running cell instance, using the new cell image provided. This is done as a rolling update, hence only changes to docker images encapsulated within components will be applied.

###### Parameters:

* _cell instance name: The name of a running cell instance, which should be updated._
* _cell image name: The name of the new cell image, which will be used to perform a rolling update on the running instance's components._

OR
* _cell instance name: The name of a running cell instance, which should be updated._
* _target component name: The name of the component of the running instance to update._

###### Flags (Mandatory):

* _-i, --container-image: container image name which will be used to update the target component. Only applicable when a target component is specified._

###### Flags (Optional):

* _-e, --env: environment variable name value pairs separated by a '=' sign. Only applicable when a target component is specified._

Ex:
 ```
   cellery update myhello cellery/sample-hello:1.0.3
   cellery update myhello controller --container-image mycellorg/hellocell:1.0.0
   cellery update myhello controller --container-image mycellorg/hellocell:1.0.0 --env foo=bar --env name=alice
 ```
 
[Back to Command List](#cellery-cli-commands)

#### Cellery Route Traffic

This is used to direct a percentage of traffic originating from one cell instance to another. This command is used in advanced deployment patterns such as Blue-Green and Canary. 


###### Flags (Mandatory):

* _-d, --dependency: Existing dependency instance, which is currently receiving traffic from the relevant source instance(s)._
* _-t, --target: The new target instance to which traffic should be routed to._

###### Flags (Optional):

* _-s, --source: The source instance which is generating traffic for the dependency instance specified in the Parameters above. 
If this is not given all instances which are currently depending on the provided dependency instance will be considered._
* _-p, --percentage: The percentage of traffic to be routed to the target instance. If not specified, this will be considered to be 100%._

Ex:
 ```
   cellery route-traffic --source hr-client-inst1 --dependency hr-inst-1 --target hr-inst-2 --percentage 20 
   cellery route-traffic --dependency hr-inst-1 --target hr-inst-2 --percentage 20
   cellery route-traffic --dependency hr-inst-1 --target hr-inst-2
 ```

[Back to Command List](#cellery-cli-commands)

#### Cellery Export Policy

Export a policy from the Cellery runtime as a file system artifact. This can be either exported to a file specified by the CLI user, or a file starting with cell instance name.

##### Cellery Export Policy Autoscale:

Export a set of autoscale policies which is applicable to a given cell instance.

###### Parameters: 

* _cell instance name: A valid cell instance name._

###### Flags (Optional):

* _-f, --file: File name to which the autoscale policy should be exported._


Ex:
 ```
   cellery export-policy autoscale mytestcell1 -f myscalepolicy.yaml
   cellery export-policy autoscale mytestcell1
 ```
 
 [Back to Command List](#cellery-cli-commands)
 
 #### Cellery Apply Policy
 
Apply a policy/set of policies included in a file to the Cellery runtime targeting a running cell instance. 
 
 ##### Cellery Apply Policy Autoscale:
 
 Apply a file containing a set of autoscale policies to the given cell instance.
 
 ###### Parameters: 
 
 * _autoscale policy file: A file containing a valid autoscale policy/set of autoscale policies._
 * _instance name: The target instance to which the autoscale policies should be applied._
 
 ###### Flags (Optional):
 
 * _-c, --components: Comma separated list of components of the provided instance, to which the autoscale policy should be applied._
 * _-g, --gateway: Flag to indicate that the given autoscale policy should be applied to only the gateway of the target instance._
 
 Ex:
  ```
    cellery apply-policy autoscale myscalepolicy.yaml myinstance --components comp1,comp2
    cellery apply-policy autoscale myscalepolicy.yaml myinstance --gateway
    cellery apply-policy autoscale myscalepolicy.yaml myinstance
  ```
  
 ###### Sample autoscale policy:
  ```yaml
    type: AutoscalePolicy
    rules:
    - overridable: true
      policy:
        maxReplicas: 4
        metrics:
        - resource:
            name: cpu
            targetAverageUtilization: 40
          type: Resource
        minReplicas: "1"
      target:
        name: controller
        type: component
    ---
    type: AutoscalePolicy
    rules:
      - overridable: false
        policy:
          maxReplicas: 2
          metrics:
            - resource:
                name: cpu
                targetAverageUtilization: 50
              type: Resource
          minReplicas: "1"
        target:
          type: gateway
    ---
  ```
  * If the target type is specified as 'component', that implies that the policy should be applied to the component denoted by the target name.
  * If the target type is 'gateway' it should be applied to the gateway of the relevant cell instance.
  * The flag 'overridable' implies whether the existing policy can be overriden by the same command repeatedly. 
  
  [Back to Command List](#cellery-cli-commands)

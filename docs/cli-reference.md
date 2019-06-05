## Cellery CLI Commands

* [setup](#cellery-setup) - create/manage cellery runtime installations.
* [init](#cellery-init) - initialize a cellery project.
* [build](#cellery-build) - build a cell image.
* [run](#cellery-run) - create cell instance(s). 
* [list](#cellery-list) - list running instances/cell images.
* [delete](#cellery-delete) - Delete cell images.
* [login](#cellery-login) - login to cell image repository.
* [push](#cellery-push) - push a built image to cell image repository.
* [pull](#cellery-pull) - pull an image from cell image repository.
* [terminate](#cellery-terminate) - terminate a cell instance.
* [status](#cellery-status) - check status of cell instance.
* [logs](#cellery-logs) - display logs of one/all components of a cell instance.
* [inspect](#cellery-inspect) - list the files included in a cell image. 
* [extract-resources](#cellery-extract-resources) - extract packed resources in a cell image.

#### Cellery Setup
Cellery setup command install and manage cellery runtimes. For this purpose it supports several sub commands. Please 
refer the [setup command readme](cli-setup-command.md) for complete instructions.

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

Terminate the running cell instance within cell runtime.

###### Parameters:

* _cell instance name: Name of the instance running in the cellery system_

Ex: 
 ```
   cellery terminate my-cell-inst
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

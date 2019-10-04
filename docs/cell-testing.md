# Cellery Testing

Cellery supports writing integration tests in two ways:
 1) [Inline tests](#in-line-tests)
 2) [Docker image based tests](#docker-image-based-tests)
 
## In-line tests
 
 Cellery supports writing in-line integration tests using Ballerina and [Telepresence](https://www.telepresence.io/).
 Inline tests can be written according to the standards of [Testerina](https://v1-0.ballerina.io/learn/how-to-test-ballerina-code/) which is the Ballerina test framework. 
 
 Telepresence makes the developer feel as micro services run in their local machine despite actually they are
  running in a remote kubernetes cluster. Testerina provides the smooth experience of
   writing integration tests to these local services which actually run in a kubernetes cluster.  
 
Writing in-line tests is as easy as writing an integration test to a service which runs in developer machine
. Cellery provides a set of helper functions which eases writing tests.
 
 #### Helper Functions

| Function        | Signature                                                                                                                                                    | Summary                                                                                                            |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------|
| [```getCellImage```](#functiongetcellimage)   | <code>public function getCellImage() returns (ImageName&#124; error) </code>                                                                                               | Gives information about the instance being tested                                                                  |
| [```getDependencies```](#functiongetdependencies) | <code>public function getDependencies() returns (map &#124; error)</code>                                                                                                    | Gives a map of dependency instances of the cell being tested                                                       |
| [```getGatewayHost```](#functiongetgatewayhost)  | <code>public function getGatewayHost(InstanceState[] iNameList,  string alias = "", string kind = "Cell")  returns (Reference&#124;error)</code>                               | Returns the services(endpoint URLs) exposed from the cell being tested and dependency cells                        |
| [```run```](#functionrun)           | <code>public function run(cellery:ImageName iName, map instances,  boolean startDependencies, boolean shareDependencies)  returns (cellery:InstanceState[]&#124;error?)</code> | Starts the instances and return an array of instance with their states                                             |
| [```stopInstances```](#functionstopinstances)   | <code>public function stopInstances(InstanceState[] instances) returns (error?)</code>                                                                                  | Checks for the state of the instances and stops each of them only if the instance was created by the test function |


 #### Function```getCellImage```
 
 Returns an ImageName record of the cell which is being tested. This record consists of information about Cell
 /Composite and instance information about the actual instance which is running. Below depicts how this is used to
  retrieve instance information.
  
  ```ballerina
cellery:ImageName iName = <cellery:ImageName>cellery:getCellImage();
```

[back to helper functions](#helper-functions)

#### Function```getDependencies```
  
Returns a map of ImageName records of the dependencies of the cell being tested.
  
  ```ballerina
map<cellery:ImageName> instances = <map<cellery:ImageName>>cellery:getDependencies();
```
[back to helper functions](#helper-functions)
#### Function```getGatewayHost```

Returns the endpoints which are exposed from the Cell which is being tested and from the dependency Cells.

* If alias is given - Returns the endpoints which are exposed by the dependency cell with given alias name
* If alias is not given - Returns the endpoints of the cell being tested 

```ballerina
cellery:Reference petBeEndpoints = <cellery:Reference>cellery:getGatewayHost(instanceList, alias = "petStoreBackend");

cellery:Reference allEndpoints = <cellery:Reference>cellery:getGatewayHost(instanceList);
```
[back to helper functions](#helper-functions)

#### Function```run```

The `run()` function defined in the cell file can be used to
 create the instances for the purpose of testing. This returns an error if the root instance is already
  available in the runtime and it has to be handled when writing tests. Please note that ```@test:BeforeSuite``` is
   used to make sure this runs before the test suite. 

```ballerina
# Handle creation of instances for running tests
@test:BeforeSuite
function setup() {
   cellery:ImageName iName = <cellery:ImageName>cellery:getCellImage();
 
   cellery:InstanceState[]|error? result = run(iName, {}, true, true);
   if (result is error) {
       cellery:InstanceState iNameState = {
           iName : iName,
           isRunning: true
       };
       instanceList[instanceList.length()] = iNameState;
   } else {
       instanceList = <cellery:InstanceState[]>result;
   }
}

```
[back to helper functions](#helper-functions)

#### Function```stopInstances```

Stops and destroys the instances created for the purpose of running tests. Instances that were already available before
 starting tests are kept without stopping.

```ballerina
# Handle deletion of instances for running tests
@test:AfterSuite
public function cleanUp() {
   error? err = cellery:stopInstances(instanceList);
}

```
[back to helper functions](#helper-functions)

For more information and samples about writing in-line tests please refer [here](https://github.com/wso2-cellery/samples/blob/master/docs/pet-store/test-be-cell.md)

 ## Docker image based tests

 
This mechanism gives developers the flexibility of running any test suite which is packaged as a docker image
, on top of Cells they have developed. Developers can choose their own technologies and tools seamlessly and package
 them as a docker image in order to develop their test suite. 
 
 Writing docker image based tests for Cells is straight forward since the actual test
  is in the docker images, and since developers can select any framework or language to develop their suite. Also, they
   can use any existing test cases and wrap as docker images.
   
Along with ```build``` and ```run``` methods, ```test``` method includes the instructions that needs to be executed during the Cellery
 test.
 
 ```ballerina
public function build(cellery:ImageName iName) returns error? {
    ...
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    ...
}

public function test(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    cellery:Test petBeTest = {
        name: "pet-be-test",
        source: {
            image: "docker.io/wso2cellery/pet-be-tests"
        },
        envVars: {
            PET_BE_CELL_URL: { value: <string>cellery:resolveReference(iName).controller_api_url }
        }
    };
    cellery:TestSuite petBeTestSuite = {
        tests: [petBeTest]
    };

    cellery:ImageName[] instanceList = cellery:runInstances(iName, instances);
    error? a = cellery:runTestSuite(iName, petBeTestSuite);
return cellery:stopInstances(iName, instanceList);
} 
```

* Above sample cell file shows the ```cellery test``` method, where there are two tests are defined, and added into
 the test suite. These tests are be executed sequentially one after the other in the provided order.

* The ```cellery test``` supports all the input params as same as the run method. The tests could be executed on
 already running instance, or the Cellery test framework starts a new instance and run the test against it. For
  example, if the test should be executed against ```test-be cell```, then it first checks whether there is an
   instance already exists in the runtime and if so, the tests are executed against that instance. Otherwise, new
    instances with the provided name is started, and those are be terminated once the tests are finished.
  
  For more information and samples about writing docker image based tests please refer [here](https://github.com/wso2-cellery/samples/blob/master/docs/pet-store/test-be-cell.md)
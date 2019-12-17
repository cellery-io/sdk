# Cellery Testing
 
 Cellery supports writing integration tests using Ballerina and [Telepresence](https://www.telepresence.io/).
 Tests can be written according to the standards of [Testerina](https://v1-0.ballerina.io/learn/how-to-test-ballerina-code/) which is the Ballerina test framework. 
 
 Telepresence makes the developer feel as is the micro services are running in their local machine despite the fact 
 that they are actually running in a remote kubernetes cluster. Testerina provides the smooth experience of
   writing integration tests to these local services which actually run in a kubernetes cluster.  
 
Writing integration tests with Ballerina is as easy as writing an integration test to a service which runs in developer 
machine. Cellery provides a set of helper functions which eases writing tests.
 
 #### Helper Functions

| Function        | Signature                                                                                                                                                    | Summary                                                                                                            |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------|
| [```getTestConfig```](#functiongettestconfig)   | <code>public function getTestConfig() returns TestConfig </code> | Gives information about the instance being tested |                                                         |
| [```runInstances```](#functionruninstances)           | <code>public function runInstances(testConfig)</code> | Starts the instances and skips already running ones |
| [```getInstanceEndpoints```](#functiongetinstanceendpoints)  | <code>public function getInstanceEndpoints(string alias = "")  returns Reference</code> | Returns the services(endpoint URLs) exposed from the cell being tested and dependency cells |
| [```stopInstances```](#functionstopinstances)   | <code>public function stopInstances() returns (error?)</code>  | Checks for the state of the instances and stops each of them only if the instance was created by the test function |
| [```runDockerTest```](#functionrundockertest)   | <code>public function runDockerTest(string imageName, map<Env> envVars) returns (error?)</code> | Runs a docker image based test |


 #### Function```getTestConfig```
 
 Returns a TestConfig record related to the cell which is being tested. This record consists of the image name, 
 instance name, and condition to start and share dependencies which are passed by the user to the CLI command. Below 
 depicts how this is used to retrieve configuration relate to the tests. The return value should be passed to 
 [```runInstances``` helper function](#functionruninstances) to start the instances.
  
  ```ballerina
cellery:TestConfig testConfig = <@untainted>cellery:getTestConfig();
```

[back to helper functions](#helper-functions)

#### Function```runInstances```

This starts the required instances and reuses already existing for testing. Please note that ```@test:BeforeSuite``` is used to make sure this runs before the test suite. 

```ballerina
# Handle creation of instances for running tests
@test:BeforeSuite
function setup() {
   cellery:runInstances(testConfig);
}

```
[back to helper functions](#helper-functions)

#### Function```getInstanceEndpoints```

Returns the endpoints which are exposed from the cell under test or dependency instance  
Cells.

* If alias is given - Returns the endpoints which are exposed by the dependency cell with given alias name
* If alias is not given - Returns the endpoints of the cell being tested 

```ballerina
cellery:Reference petBeEndpoints = <cellery:Reference>cellery:getInstanceEndpoints(alias = "petStoreBackend");

cellery:Reference allEndpoints = <cellery:Reference>cellery:getInstanceEndpoints();
```
[back to helper functions](#helper-functions)

#### Function```stopInstances```

Stops and destroys the instances created for the purpose of running tests. Instances that were already available 
before starting tests are kept without stopping.

```ballerina
# Handle deletion of instances for running tests
@test:AfterSuite
public function cleanUp() {
   error? err = cellery:stopInstances();
}

```
[back to helper functions](#helper-functions)

For more information and samples about writing integration tests please refer [here](https://github.com/wso2/cellery-samples/blob/master/docs/pet-store/test-be-cell.md)

#### Function```runDockerTest```
 
This helper function gives developers the flexibility of running any test suite which is packaged as a docker image, on top of Cells they have developed. Developers can choose their own technologies and tools seamlessly and package
 them as a docker image in order to develop their test suite. 
 
 Writing docker image based tests for Cells is straightforward since the actual test
  is in the docker image, and since developers can select any framework or language to develop their suite. Also, they
   can use any existing test cases and wrap as docker images.
   
Docker image based tests are also written as ballerina tests using this helper function. 
 
 ```ballerina
@test:Config {}
function testDocker() {
    map<cellery:Env> envVars = {PET_BE_CELL_URL: { value: PET_BE_CONTROLLER_ENDPOINT }};
    error? err = cellery:runDockerTest("docker.io/wso2cellery/samples-pet-store-order-tests", envVars);
    test:assertFalse(err is error, msg = "Docker image test failed.\n Reason: \n" + err.toString() + "\n");
}
```

* Above sample ballerina test Cell shows the way of defining docker image based test for the Cell.

The ```cellery test``` command supports all the input params as same as the run method. The tests could be executed on
 already running instance, or the Cellery test framework starts a new instance and run the test against it. For
  example, if the test should be executed against ```test-be cell```, then it first checks whether there is an
   instance already exists in the runtime and if so, the tests are executed against that instance. Otherwise, new
    instances with the provided name is started, and those are be terminated once the tests are finished.
  
  For more information and samples about writing docker image based tests please refer [here](https://github.com/wso2/cellery-samples/blob/master/docs/pet-store/test-be-cell.md)
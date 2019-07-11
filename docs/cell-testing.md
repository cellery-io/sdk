# Cellery Testing

Cellery focuses on the integration testing of the cell, where the tests are executed against the running cells. Along with `build` and `run` methods,
`test` method includes the instructions that needs to be executed  during the `cellery test`. 

```ballerina
import celleryio/cellery;

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

- Above sample cell file shows the cellery `test` method, where there are two tests are defined, and added into the test suite. 
These tests will be executed sequentially one after the other in the provided order. 

- The actual test is in the docker images, and developers can select any framework and language to develop the test. Also, 
can use any existing test cases and wrap as docker images to execute the tests. 

- The `cellery test` will support all the input params as same as the `run` method. The tests could be executed on already running instance, 
or the cellery test framework will start a new instance and run the test against it. For example, if the test should be executed against `test-be` cell, then it will first check whether
there is an instance already exists in the runtime and if so, the test will be executed against that instance. Otherwise, 
new instances with the provided name will be started, and those will be terminated once the tests are executed. 
    ```
    $ cellery run wso2cellery/pet-be-cell:latest -n pet-be -y
    $ cellery test wso2cellery/pet-be-cell:latest -n pet-be
 
    OR

    $ cellery test wso2cellery/pet-be-cell:latest -n pet-be 

    OR 
    $ cellery test wso2cellery/pet-be-cell:latest -n pet-be -d 
    ```
- Once  the tests are executed, the logs are stored in the same location from where the command is executed. 

# What's Next?
- [Pet store sample test](https://github.com/wso2-cellery/samples/blob/master/docs/pet-store/test-be-cell.md) - takes through the sample cellery test.
- [Developing and runing a Cell](writing-a-cell.md) - step by step explanation on how you could define your own cells.
- [How to code cells?](cellery-syntax.md) - explains how Cellery cells are written.
- [Cell update and advanced deployments](cell-update-and-adv-deployment.md) - update components of running cell instance using a new cell image and create advanced deployments.
- [Observe cells](cellery-observability.md) - provides the runtime insight of cells and components.

## Developing & running a cell with resources and/or tests

In this section let's focus on initialize, build, run and push operations for pet-be cell. Pet-be cell requires a 
swagger file when building the image. Therefore, this has to be written as a Ballerina project. Follow the instructions listed below to create the cell.

1. Execute `cellery init --project` command from the command prompt, and provide the project name as `pet-be`. 
    ```
    $ cellery init --project
    ? Project name:  [my-project] pet-be
    Created new Ballerina project at pet-be

    Next:
        Move into the project directory and use `ballerina add <module-name>` to
        add a new Ballerina module.
    Added new ballerina module at 'src/pet_be'


    ✔ Initialized Ballerina project: pet-be

    What's next?
    --------------------------------------------------------
    Execute the following command to build the image:
      $ cellery build pet-be organization/image_name:version
    --------------------------------------------------------

    ```
The above step will auto generate a Ballerina project with the name pet-be in the present working directory. The structure of the Ballerina project generated is shown below.

    pet-be/
    ├── Ballerina.toml
    └── src
        └── pet_be
            ├── pet_be.bal
            ├── resources
            └── tests
                ├── pet_be_test.bal
                └── resources

All files related to the cell are created within the `pet_be` Ballerina module. Tests has to be written for the 
Ballerina module `pet_be`. The cell file is identical to the file that was created when [developing & running your first cell](https://github.com/wso2/cellery/blob/master/docs/writing-a-cell.md). 

2. Modify the files to add components and tests required for the pet-be cell. The final project after modifications 
can be found in the [pet-be sample](https://github.com/wso2/cellery-samples/tree/master/cells/pet-store/pet-be).

3. Build the Cellery image for pet-be project by executing the `cellery build` command as shown below. For building 
the cell you should give the path to the `pet-be` Ballerina project as the argument to the CLI command.
Note `CELLERY_HUB_ORG` is your organization name in [Cellery hub](https://hub.cellery.io/).
    ```
    $ cellery build hello-world-cell.bal <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
    ✔ Creating temporary executable main bal file
    Compiling source
        celleryio/pet_be:0.6.0

    Creating balos
        target/balo/pet_be-2019r3-any-0.1.0.balo

    Generating executables
        target/bin/pet_be.jar

    Running executables

    ✔ Executing ballerina build
    ✔ Generating metadata
    ✔ Creating the image zip file


    ✔ Successfully built image: <CELLERY_HUB_ORG>/pet-be-cell:1.0.0

    What's next?
    --------------------------------------------------------
    Execute the following command to run the image:
      $ cellery run <CELLERY_HUB_ORG>/pet-be-cell:1.0.0
    --------------------------------------------------------
    ```

4. Execute `cellery run`, `cellery list instances`, `cellery view`, `cellery push` and `cellery terminate` commands 
as you already did when [developing & running your first cell](https://github.com/wso2/cellery/blob/master/docs/writing-a-cell.md).
    
### What's Next?
- [Cell Specification](https://github.com/wso2/cellery-spec/blob/master/README.md) - Key concepts of Cellery.
- [How to code cells?](cellery-syntax.md) - explains how Cellery cells are written.
- [CLI commands](cli-reference.md) - reference for CLI commands.
- [Cell API versioning](cell-api-versioning.md) - api versioning for cells. 
- [Samples](https://github.com/wso2/cellery-samples/tree/master) - a collection of useful samples.

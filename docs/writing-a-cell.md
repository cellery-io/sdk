## Developing & running your first cell

In this section let's focus on initialize, build, run and push the same hello world cell explained above. 
Follow the instructions listed below to create your first cell. To create a cell that requires additional resources, you can create a Ballerina project for the cell. To get started with a Ballerina project, follow instructions given [here]().

1. Execute `cellery init` command from the command prompt, and provide the project name as `hello-world-cell`. 
    ```
    $ cellery init
    ? Project name:  [my-project] hello-world-cell
    
    
    ✔ Initialized project in directory: /Users/cellery/hello-world-cell
    
    What's next?
    --------------------------------------------------------
    Execute the following command to build the image:
      $ cellery build helloworld/helloworld.bal [repo/]organization/image_name:version
    --------------------------------------------------------
    ```
2. The above step will auto generate a Cellery file in the location: hello-world-cell/hello-world-cell.bal with below content. 
This file is implemented in [Ballerina](https://ballerina.io/) - A Cloud Native Programming Language. However, Ballerina language is used
for simple assignments to define your applications in Cellery, and hence you do not need to learn or have extensive knowledge in 
Ballerina language to use Cellery. Basically, the cell file consists of three methods, `build` and `run`, where `build` 
method will be invoked during `cellery build` operation, 
`run` method will be invoked during `cellery run` operation. In each of these method, you define
what should be occurred for your component/cell/composite.   

3. You define how your [Cell](cellery-syntax.md#cell) or [Composite](cellery-syntax.md#composite) should be packaged, 
and the list of components inside it, during the `build` method. At the end of the method, you will be creating the Cell or Composite image via `cellery:createImage()`
function. Below shown cell `helloCell` consists of one [component](cellery-syntax.md#component) defined as `helloComponent` and it has one [web ingress](cellery-syntax.md#2-web-ingress) with default vhost `hello-world.com`.
An environment variable `HELLO_NAME` is expected by `helloComponent` to render the webpage, and it is assigned `Cellery` as default. 
 
    ```ballerina
    import ballerina/config;
    import celleryio/cellery;
    
    public function build(cellery:ImageName iName) returns error? {
        // Hello Component
        // This Components exposes the HTML hello world page
        cellery:Component helloComponent = {
            name: "hello",
            src: {
                image: "wso2cellery/samples-hello-world-webapp"
            },
            ingresses: {
                webUI: <cellery:WebIngress>{ // Web ingress will be always exposed globally.
                    port: 80,
                    gatewayConfig: {
                        vhost: "hello-world.com",
                        context: "/"
                    }
                }
            },
            envVars: {
                HELLO_NAME: { value: "Cellery" }
            }
        };
    
        // Cell Initialization
        cellery:CellImage helloCell = {
            components: {
                helloComp: helloComponent
            }
        };
        return <@untainted> cellery:createImage(helloCell, untaint iName);
    }
    
    public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[] | error?) {
        cellery:CellImage helloCell = check cellery:constructCellImage(untaint iName);
        string vhostName = config:getAsString("VHOST_NAME");
        if (vhostName !== "") {
            cellery:WebIngress web = <cellery:WebIngress>helloCell.components["helloComp"]["ingresses"]["webUI"];
            web["gatewayConfig"]["vhost"] = vhostName;
        }
    
        string helloName = config:getAsString("HELLO_NAME");
        if (helloName !== "") {
            helloCell.components["helloComp"]["envVars"]["HELLO_NAME"].value = helloName;
        }
        return <@untainted> cellery:createInstance(helloCell, iName, instances);
    }
    ```
4. The `run` method will be invoked during the `cellery run` operation, and users can pass environment variables to 
modify how the cell/composite instance should be created. As shown in the above example, the `vhost` parameter in the web ingress and the environment variable `HELLO_NAME` can be modified by resolving the environmental
variables `VHOST_NAME` and `HELLO_NAME` in the runtime. This is performed via having `config:getAsString("<ENV_NAME>")` - a utility function available in [Ballerina](https://ballerina.io/) within the 
`run` function. The environment variables can be optionally passed via `cellery run <image-name> -e VHOST_NAME=modified.hello.com -e HELLO_NAME=ModifiedCellery`. 

5. Build the Cellery image for hello world project by executing the `cellery build` command as shown below. 
Note `CELLERY_HUB_ORG` is your organization name in [Cellery hub](https://hub.cellery.io/).
    ```
    $ cd hellow-world-cell
    $ cellery build hello-world-cell.bal <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
    Hello World Cell Built successfully.
    
    ✔ Building image <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
    ✔ Saving new Image to the Local Repository
    
    
    ✔ Successfully built cell image: <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
    
    What's next?
    --------------------------------------------------------
    Execute the following command to run the image:
      $ cellery run <CELLERY_HUB_ORG>/helloworld:1.0.0
    --------------------------------------------------------
    ```

6. Note that in the cell file's run method at step-2, it's looking for runtime parameters `VHOST_NAME` and `HELLO_NAME`, 
and if it's available then it'll will be using those as vhost and greeting name. 
Therefore run the built Cellery image with `cellery run` command,  and pass `my-hello-world.com` for `VHOST_NAME`, 
and your name for `HELLO_NAME` as shown below. Environment variables can be passed into the Cellery file as explained [here](https://github.com/wso2/cellery-spec).
    ```
    $ cellery run <CELLERY_HUB_ORG>/hello-world-cell:1.0.0 -e VHOST_NAME=my-hello-world.com -e HELLO_NAME=WSO2 -n my-hello-world
       ✔ Extracting Cell Image  <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
       
       Main Instance: my-hello-world
       
       ✔ Reading Cell Image  <CELLERY_HUB_ORG/hello-world-cell:1.0.0
       ✔ Validating environment variables
       ✔ Validating dependencies
       
       Instances to be Used:
       
         INSTANCE NAME              CELL IMAGE                      USED INSTANCE   SHARED
        ---------------- ----------------------------------------- --------------- --------
         my-hello-world    <CELLERY_HUB_ORG>/hello-world-cell:1.0.0   To be Created    -
       
       Dependency Tree to be Used:
       
        No Dependencies
       
       ? Do you wish to continue with starting above Cell instances (Y/n)?
       
       ✔ Starting main instance my-hello-world
       
       
       ✔ Successfully deployed cell image:  <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
       
       What's next?
       --------------------------------------------------------
       Execute the following command to list running cells:
         $ cellery list instances
       --------------------------------------------------------
    ```
    
5. Now your hello world cell is deployed, you can run the `cellery list instances` command to see the status of the deployed cell.
    ```
    $ cellery list instances
                        INSTANCE                                   CELL IMAGE                   STATUS                            GATEWAY                            COMPONENTS           AGE
       ------------------------------------------ -------------------------------------------- -------- ----------------------------------------------------------- ------------ ----------------------
        hello-world-cell-1-0-0-676b2131   sinthuja/hello-world-cell:1.0.0              Ready    sinthuja-hello-world-cell-1-0-0-676b2131--gateway-service   1            10 minutes 1 seconds
    ```
6. Execute `cellery view` to see the components of your cell. This will open a HTML page in a browser and you can visualize the components and dependent cells of the cell image.
    ```
    $ cellery view <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
    ```
    ![hello world cell view](images/hello-web-cell.png)
    
7. Access url [http://my-hello-world.com/](http://my-hello-world.com/) from browser. You will see updated web page with greeting 
param you passed for HELLO_NAME in step-4.

8. As a final step, let's push your first cell project to your docker hub account. To perform this execute `cellery push` 
as shown below.
    ```
    $ cellery push <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
    ✔ Connecting to registry-1.docker.io
    ✔ Reading image <CELLERY_HUB_ORG>/hello-world-cell:1.0.0 from the Local Repository
    ✔ Checking if the image <CELLERY_HUB_ORG>/hello-world-cell:1.0.0 already exists in the Registry
    ✔ Pushing image <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
    
    Image Digest : sha256:8935b3495a6c1cbc466ac28f4120c3836894e8ea1563fb5da7ecbd17e4b80df5
    
    ✔ Successfully pushed cell image: <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
    
    What's next?
    --------------------------------------------------------
    Execute the following command to pull the image:
      $ cellery pull <CELLERY_HUB_ORG>/hello-world-cell:1.0.0
    --------------------------------------------------------
    ```
 Congratulations! You have successfully created your own cell, and completed getting started!
 
### Clean up
You can terminate the cells that are started during this guide.

1) List the cells that are running in the current setup by `cellery list instances`.
    ```
    $ cellery list instances
         INSTANCE                      CELL IMAGE          STATUS               GATEWAY               COMPONENTS            AGE
     ---------------- ----------------------------------- -------- --------------------------------- ------------ -----------------------
      hello            wso2cellery/hello-world-cell:0.2.1    Ready    hello--gateway-service            1            1 hours 2 minutes
      my-hello-world   <ORGNAME>/hello-world-cell:1.0.0      Ready    my-hello-world--gateway-service   1            27 minutes 42 seconds
    ```
2) Execute terminate command for each cell instances that you want to clean up as shown below.
    ```
    $ cellery terminate hello
    $ cellery terminate my-hello-world
    ```
    OR Execute below command to terminal all running cell instances.
    ```
    $ cellery term --all
    ```
### What's Next?
- [Developing & running a cell with resources and/or tests](https://github.com/wso2/cellery/blob/master/docs/writing-a-cell-bal-project.md) - Step by step explanation on how you could define cells that has resources and/or tests
- [Cell Specification](https://github.com/wso2/cellery-spec/blob/master/README.md) - Key concepts of Cellery.
- [How to code cells?](cellery-syntax.md) - explains how Cellery cells are written.
- [CLI commands](cli-reference.md) - reference for CLI commands.
- [Cell API versioning](cell-api-versioning.md) - api versioning for cells. 
- [Samples](https://github.com/wso2/cellery-samples/tree/master) - a collection of useful samples.

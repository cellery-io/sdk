## Developing & running your first cell

In this section let's focus on initialize, build, run and push the same hello world cell explained above. 
Follow the instructions listed below to create your first cell.

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
This file is implemented in [Ballerina](https://ballerina.io/) - A Cloud Native Programming Language. 
The cell `helloCell` consists of one component defined as `helloComponent` and it has one web ingress with default 
vhost `hello-world.com`.
An environment variable `HELLO_NAME`with default value `Cellery` is used by `helloComponent` to render the webpage. 
By passing the  parameters in the runtime, the vhost entry and the environment variable HELLO_NAME can be modified. 

    ```ballerina
    import ballerina/config;
    import celleryio/cellery;
    
    public function build(cellery:ImageName iName) returns error? {
        // Hello Component
        // This Components exposes the HTML hello world page
        cellery:Component helloComponent = {
            name: "hello",
            source: {
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
        return cellery:createImage(helloCell, untaint iName);
    }
    
    public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
        cellery:CellImage helloCell = check cellery:constructCellImage(untaint iName);
        string vhostName = config:getAsString("VHOST_NAME");
        if (vhostName !== "") {
            cellery:WebIngress web = <cellery:WebIngress>helloCell.components.helloComp.ingresses.webUI;
            web.gatewayConfig.vhost = vhostName;
        }
    
        string helloName = config:getAsString("HELLO_NAME");
        if (helloName !== "") {
            helloCell.components.helloComp.envVars.HELLO_NAME.value = helloName;
        }
        return cellery:createInstance(helloCell, iName, instances);
    }
    ```

3. Build the Cellery image for hello world project by executing the `cellery build` command as shown below. 
Note `DOCKER_HUB_ORG` is your organization name in docker hub.
    ```
    $ cd hellow-world-cell
    $ cellery build hello-world-cell.bal <DOCKER_HUB_ORG>/hello-world-cell:1.0.0
    Hello World Cell Built successfully.
    
    ✔ Building image <DOCKER_HUB_ORG>/hello-world-cell:1.0.0
    ✔ Saving new Image to the Local Repository
    
    
    ✔ Successfully built cell image: <DOCKER_HUB_ORG>/hello-world-cell:1.0.0
    
    What's next?
    --------------------------------------------------------
    Execute the following command to run the image:
      $ cellery run <DOCKER_HUB_ORG>/helloworld:1.0.0
    --------------------------------------------------------
    ```

4. Note that in the cell file's run method at step-2, it's looking for runtime parameters `VHOST_NAME` and `HELLO_NAME`, 
and if it's available then it'll will be using those as vhost and greeting name. 
Therefore run the built Cellery image with ‘cellery run’ command, 
and pass `my-hello-world.com` for `VHOST_NAME`, and your name for `HELLO_NAME` as shown below. 
Environment variables can be passed into the Cellery file as explained [here](https://github.com/wso2-cellery/spec).
    ```
    $ cellery run <DOCKER_HUB_ORG>/hello-world-cell:1.0.0 -e VHOST_NAME=my-hello-world.com -e HELLO_NAME=WSO2 -n my-hello-world
       ✔ Extracting Cell Image  <DOCKER_HUB_ORG/hello-world-cell:1.0.0
       
       Main Instance: my-hello-world
       
       ✔ Reading Cell Image  <DOCKER_HUB_ORG/hello-world-cell:1.0.0
       ✔ Validating environment variables
       ✔ Validating dependencies
       
       Instances to be Used:
       
         INSTANCE NAME              CELL IMAGE                      USED INSTANCE   SHARED
        ---------------- ----------------------------------------- --------------- --------
         my-hello-world    <DOCKER_HUB_ORG>/hello-world-cell:1.0.0   To be Created    -
       
       Dependency Tree to be Used:
       
        No Dependencies
       
       ? Do you wish to continue with starting above Cell instances (Y/n)?
       
       ✔ Starting main instance my-hello-world
       
       
       ✔ Successfully deployed cell image:  <DOCKER_HUB_ORG>/hello-world-cell:1.0.0
       
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
    $ cellery view <DOCKER_HUB_ORG>/hello-world-cell:1.0.0
    ```
    ![hello world cell view](docs/images/hello-web-cell.png)
    
7. Access url [http://my-hello-world.com/](http://my-hello-world.com/) from browser. You will see updated web page with greeting 
param you passed for HELLO_NAME in step-4.

8. As a final step, let's push your first cell project to your docker hub account. To perform this execute `cellery push` 
as shown below.
    ```
    $ cellery push <DOCKER_HUB_ORG>/hello-world-cell:1.0.0
    ✔ Connecting to registry-1.docker.io
    ✔ Reading image <DOCKER_HUB_ORG>/hello-world-cell:1.0.0 from the Local Repository
    ✔ Checking if the image <DOCKER_HUB_ORG>/hello-world-cell:1.0.0 already exists in the Registry
    ✔ Pushing image <DOCKER_HUB_ORG>/hello-world-cell:1.0.0
    
    Image Digest : sha256:8935b3495a6c1cbc466ac28f4120c3836894e8ea1563fb5da7ecbd17e4b80df5
    
    ✔ Successfully pushed cell image: <DOCKER_HUB_ORG>/hello-world-cell:1.0.0
    
    What's next?
    --------------------------------------------------------
    Execute the following command to pull the image:
      $ cellery pull <DOCKER_HUB_ORG>/hello-world-cell:1.0.0
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
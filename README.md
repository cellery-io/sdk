![Cellery Logo](docs/images/cellery-logo.svg)

[![GitHub (pre-)release](https://img.shields.io/github/release/wso2-cellery/sdk/all.svg)](https://github.com/wso2-cellery/sdk/releases)
[![GitHub (Pre-)Release Date](https://img.shields.io/github/release-date-pre/wso2-cellery/sdk.svg)](https://github.com/wso2-cellery/sdk/releases)
[![GitHub last commit](https://img.shields.io/github/last-commit/wso2-cellery/sdk.svg)](https://github.com/wso2-cellery/sdk/commits/master)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Cellery is a code-first approach to building, integrating, running and managing composite microservice applications on Kubernetes. 
Build, push/pull, run, scale and observe composites. Secure APIs by default. Code in any language.

----

## Getting Started
### Pre requisites 
1. kubectl 
2. Google Cloud SDK (required to install cellery mesh in GCP)
3. Ballerina version 0.990.3
4. VirtualBox (for the local installer)

### How to Install

#### Linux
Download the [cellery-ubuntu-x64-0.1.0.deb](https://github.com/wso2-cellery/sdk/releases) and install it using dpkg command as shown below.
```
dpkg -i cellery-ubuntu-x64-0.1.0.deb
```
#### Mac OS
Download [cellery-0.1.0.pkg](https://github.com/wso2-cellery/sdk/releases) and install it by following macOS package installation steps.

### Set up Cellery Runtime
Once the cellery is installed, the users can install the cellery mesh runtime locally on your machine as an Virtual 
Machine or in GCP. The steps to setup cellery runtime is provided below.  

1. As a first step, user has to execute `cellery setup` command to configure Cellery runtime. This 
will prompt a list of selections. There are three options to select from; create, manage and switch. By selecting 
`create ` section users can setup the Cellery runtime. 
```
$ cellery setup
[Use arrow keys]
? Setup Cellery runtime
    Manage
  ➤ Create
    Switch
    EXIT
```

2. When `create` is selected 2 options will be prompted; `Local` and `GCP`.
cellery setup. You can select either options based on your installation requirement.
```
$ ✔ Create
[Use arrow keys]
? Select an environment to be installed
  ➤ Local
    GCP
    BACK
```

#### Local Setup
As mentioned in setup guide above, select Local to setup the local environment. 
This will download and install a pre-configured cellery runtime. Once the installation process is completed, 
users can start working on a cellery project.

Next add /etc/host entries to access cellery hosts such as APIM, Observability, etc
```
127.0.0.1 wso2-apim cellery-dashboard wso2sp-observability-api wso2-apim-gateway
```

#### GCP
First users will have to create a GCP project and enable relevant APIs. After that user needs to install GCloud SDK and 
point the current project to the already configured GCP project.

1. Use gloud init command to point the GCP project that should be used for cellery runtime installation.
```
gcloud init
```
2. After initializing the project user needs to create the GCP API credentials and copy the credential json file into 
$HOME/.cellery/gcp folder.
3. Then run `cellery setup` command and select `create` then `GCP` from the menu as mentioned in the 3rd step at Setup Runtime.

```
cellery setup
$ ✔ Create
[Use arrow keys]
? Select an environment to be installed
    Local
  ➤ GCP
    BACK
```

This will start the process of creating a Cellery runtime in GCP.
```
✔ Creating GKE client
✔ Creating GCP cluster
⠸ Updating kube config clusterFetching cluster endpoint and auth data.
⠼ Updating kube config clusterkubeconfig entry generated for cellery-cluster520.
✔ Updating kube config cluster
✔ Creating sql instance
✔ Creating gcp bucket
✔ Uploading init.sql file to dcp bucket
✔ Updating bucket permission
✔ Importing sql script
✔ Updating sql instance
✔ Creating NFS server
✔ Deploying Cellery runtime

✔ Successfully installed Cellery runtime.

What's next ?
======================
To create your first project, execute the command:
  $ cellery init
```
When the process is completed Cellery will point to the newly created GCP cluster and user can start working on a 
Cellery project. 

Next get the Nginx Ip and add it to the /etc/hosts.
Run the following kubectl command to get the Ip address
```
kubectl get ingress -n cellery-system
```
Then update the /etc/hosts file with that Ip as follows
```
<IP_ADDRESS> wso2-apim cellery-dashboard wso2sp-observability-api wso2-apim-gateway
```

### Build and Deploy your first Cell 
Now we can deploy a hello world cell as a the first project. Follow the instructions listed below to deploy the hello world cell.

1. Execute cellery init command from the command prompt, and provide the project name as ‘helloworld’. 
```
$ cellery init
? Project name:  [my-project] helloworld


✔ Initialized project in directory: /Users/sinthu/wso2/sources/dev/git/mesh-observability/docker/helloworld

What's next?
--------------------------------------------------------
Execute the following command to build the image:
  $ cellery build helloworld/helloworld.bal [repo/]organization/image_name:version
--------------------------------------------------------
```
2. The above step will auto generate a cellery file in the location: helloworld/helloworld.bal with below content. 
As you can see there is one component defined in the cell as ‘helloWorldComp’, and has defined one ingress with context 
path ‘hello’. Therefore the hello API can be invoked within the cellery runtime by other cells. Further within the build 
method of the cell file, the hello ingress is also exposed as global API therefore the same API can be accessed external 
of cellery runtime. 

```ballerina
import ballerina/io;
import celleryio/cellery;

cellery:Component helloWorldComp = {
    name: "hello-world",
    source: {
        image: "sumedhassk/hello-world:1.0.0"
    },
    ingresses: {
        hello: new cellery:HttpApiIngress(
                   9090,

                   "hello",
                   [
                       {
                           path: "/*",
                           method: "GET"
                       }
                   ]

        )
    }
};

cellery:CellImage helloCell = new();

public function build(string orgName, string imageName, string imageVersion) {
    helloCell.addComponent(helloWorldComp);

    helloCell.exposeGlobal(helloWorldComp);

    var out = cellery:createImage(helloCell, orgName, imageName, imageVersion);
    if (out is boolean) {
        io:println("Hello World Cell Built successfully.");
    }
}
```

3. Build the cellery image for hello world project by executing the cellery build command as shown below.
```
$ cellery build helloworld.bal myorg/helloworld:1.0.0
Hello World Cell Built successfully.

✔ Building image myorg/helloworld:1.0.0
✔ Saving new Image to the Local Repository


✔ Successfully built cell image: myorg/helloworld:1.0.0

What's next?
--------------------------------------------------------
Execute the following command to run the image:
  $ cellery run myorg/helloworld:1.0.0
--------------------------------------------------------
```

4. Run the built cellery image with ‘cellery run’ command. 
```
$ cellery run myorg/helloworld:1.0.0
Running cell image: myorg/helloworld:1.0.0
cell.mesh.cellery.io/helloworld created


✔ Successfully deployed cell image: myorg/helloworld:1.0.0

What's next?
--------------------------------------------------------
Execute the following command to list running cells:
  $ cellery list instances
--------------------------------------------------------
```

5. Now the hello world cell is deployed, you can run the cellery list instances command to see the status of the deployed cell.
Wait until the cell becomes into ‘Ready’ state.
```
$ cellery list instances
NAME         STATUS     GATEWAY                       SERVICES   AGE
helloworld   Ready   helloworld--gateway-service   1          3m
```

6. Login to API Manager’s store application with below details.
```
URL: https://wso2-apim/store/
Username: admin
Password: admin
```

7. Click on the API with name ‘helloworld_global_1_0_0_hello - 1.0.0’ which the global API published by the hello world 
cell that you deployed.The subscribe and generate the token as described in WSO2 APIM documentation.  

8. Now you can invoke the API externally from your machine as shown below. The <access_token> is the token that you 
generated in step - 7, and replace your taken instead of <access_token>. The context of the API can be derived from 
the API that was published in the API Manager, and the ‘sayHello’ is the resource that was implemented in the actual 
hello world service.
```
$ curl https://wso2-apim-gateway/helloworld/hello/sayHello -H "Authorization: Bearer <access_token>" -k
Hello, World!
```

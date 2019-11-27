GCP setup
---
## Prerequisites
### Mandatory
- [Kubectl version 1.13/1.14/1.15](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Cellery](../../README.md#download-cellery)

### Optional
- [Ballerina 1.0.3](https://ballerina.io/downloads/)
If Ballerina 1.0.3 is not installed, Cellery will execute ballerina using Docker.

To create a GCP based cellery installation, you need to have GCP account and [Gloud SDK](https://cloud.google.com/sdk/docs/) installed in your machine. 
Follow the below steps to install cellery into your GCP.

i. Use gloud init command and create a project if not exists. Make sure gcloud is configured to the project which you want to install cellery runtime, and also billing is enabled.
```
   gcloud init
```
ii. Make sure zone and region is set correctly for the project. Execute below mentioned command.
```
   gcloud config list --format json
```
The expected output from the command should be as below with zone, and region fields.
```
    {
      "compute": {
        "region": "us-central1",
        "zone": "us-central1-c"
      },
      "core": {
        "account": "testcellery@gmail.com",
        "disable_usage_reporting": "True",
        "project": "cellery-gcp-test"
      }
    }
```

iii. If region or zone is not set as above, then please go to [GCP console](https://console.cloud.google.com/compute/settings?_ga=2.20830648.-1274616255.1554447482), and select default zone and region for your project created. 
   OR 
   Use gcloud CLI to set the project zone and the region.
   ```
   gcloud config set project <project name>
   gcloud config set compute/region <region>
   gcloud config set compute/zone <zone>
   ```

iv. Cellery uses some APIs to create resources in the GCP, therefore we need to enable the below listed APIs. 
You can enable this via the GCP Dashboard by going to menu pane, and selecting APIs and Services > Dashboard options. 
    - Cloud Engine API
    - Kubernetes API
    - Cloud Filestore API
    - Cloud SQL Admin API

   Or you can execute commands via gcloud as below.
```
    gcloud services enable container.googleapis.com file.googleapis.com sqladmin.googleapis.com
```

v. Since cellery creates the resources in the GCP, it needs the API key to access the above mentioned APIs. Hence create the key for default service account via 
selecting IAM & Admin > Service Account from left menu pane. And then select the default service account > Create Key > JSON options, and download the JSON file. Copy the 
downloaded JSON file into directory `$HOME/.cellery/gcp folder`.

vi. Now we are ready install cellery into GCP, either using interactive or non-interactive. 

### Interactive Method

    ```
    cellery setup
    $ ✔ Create
    [Use arrow keys]
    ? Select an environment to be installed
        Local
      ➤ GCP
        BACK
    ```
    OR 
    ```
    cellery setup create gcp [--complete]
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
    ✔ Uploading init.sql file to GCP bucket
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
    
When the process is completed Cellery will point to the newly created GCP cluster and user can start working on a Cellery project. 

Once the installation is completed [configure the host entries](#configure-host-entries).

### Inline Method

With the non-interactive method, creating the GCP setup with CLI commands with one go is supported. This supports both complete and basic setups as shown below:

| Setup Option | Package | Command <img width=1100/> | Description |
|--------------|------|---------|-------------|
| GCP | Basic | `cellery setup create gcp -y` | Creates basic GCP setup. This will spawn a GCP kubernetes cluster and create resources for the cellery runtime. You will require GCloud SDK as pre-requisite to perform this operation. Please check [GCP](#2.-gcp) for the steps.| 
| GCP | Complete | `cellery setup create gcp --complete -y` | Creates complete GCP setup. This will spawn a GCP kubernetes cluster and create resources for the cellery runtime. You will require GCloud SDK as pre-requisite to perform this operation. Please check [GCP](#2.-gcp) for the steps| 

Once the installation is completed [configure the host entries](#configure-host-entries)

## Configure host entries

Get the Nginx Ip and add it to the /etc/hosts. Run the following kubectl command to get the IP address.
```
 kubectl get ingress -n cellery-system
```
Then update the /etc/hosts file with that Ip as follows.  
```
 <IP Address> wso2-apim cellery-dashboard wso2sp-observability-api wso2-apim-gateway cellery-k8s-metrics idp.cellery-system pet-store.com hello-world.com my-hello-world.com
```

## Trying Out

Once the installation process is completed, you can try out [quick start with cellery](../../README.md#quickstart-guide).

## Cleaning Up

Please refer readme for [managing cellery runtimes](./manage-setup.md) for details on how to clean up the setup.

## What's Next?

- [Developing a Cell](writing-a-cell.md) - step by step explanation on how you could define your own cells.
- [Cell Specification](https://github.com/wso2/cellery-spec/blob/master/README.md) - Key concepts of Cellery.
- [How to code cells?](../cellery-syntax.md) - explains how Cellery cells are written.
- [CLI commands](../cli-reference.md) - reference for CLI commands.
- [Samples](https://github.com/wso2/cellery-samples/tree/master) - a collection of useful samples.

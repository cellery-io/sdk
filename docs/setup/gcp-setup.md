### 2. GCP setup
To create a GCP based cellery installation, you need to have GCP account and [Gloud SDK](https://cloud.google.com/sdk/docs/) installed in your machine. 
Follow the below steps to install cellery into your GCP.

1. Use gloud init command and create a project if not exists. Make sure gcloud is configured to the project which you want to install cellery runtime, and also billing is enabled.
    ```
    gcloud init
    ```
2. Make sure zone and region is set correctly for the project. Execute below mentioned command.
    ```
    gcloud config list --format json
    ```
    The expected output from the command should be as below with zone, and region fields.
    ```json
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
3. If region or zone is not set as above, then please go to [GCP console](https://console.cloud.google.com/compute/settings?_ga=2.20830648.-1274616255.1554447482), and select default zone and region for your project created. 
   OR 
   Use gcloud CLI to set the project zone and the region.
   ```
   gcloud config set project <project name>
   gcloud config set compute/region <region>
   gcloud config set compute/zone <zone>
   ```
4. Cellery uses some APIs to create resources in the GCP, therefore we need to enable the below listed APIs. 
You can enable this via the GCP Dashboard by going to menu pane, and selecting APIs and Services > Dashboard options. 
    - Cloud Engine API
    - Kubernetes API
    - Cloud Filestore API
    - Cloud SQL Admin API

    Or you can execute commands via gcloud as below.
    ```
    gcloud services enable container.googleapis.com file.googleapis.com sqladmin.googleapis.com
    ```
5. Since cellery creates the resources in the GCP, it needs the API key to access the above mentioned APIs. Hence create the key for default service account via 
selecting IAM & Admin > Service Account from left menu pane. And then select the default service account > Create Key > JSON options, and download the JSON file. Copy the 
downloaded JSON file into directory `$HOME/.cellery/gcp folder`.

6. Now we are ready install cellery into GCP. Run `cellery setup` command and select Create > GCP > Basic or Complete from the menu OR 
executing inline command with `cellery setup create gcp [--complete]`.

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
    When the process is completed Cellery will point to the newly created GCP cluster and user can start working on a 
    Cellery project. 

7. Get the Nginx Ip and add it to the /etc/hosts. Run the following kubectl command to get the IP address.
    ```
    kubectl get ingress -n cellery-system
    ```
    Then update the /etc/hosts file with that Ip as follows.  
    ```
    <IP Address> wso2-apim cellery-dashboard wso2sp-observability-api wso2-apim-gateway cellery-k8s-metrics idp.cellery-system pet-store.com hello-world.com my-hello-world.com
    ```
8. As the installation process is completed, you can [quick start with cellery](../../README.md#quick-start-with-cellery).

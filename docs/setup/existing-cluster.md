Existing Kubernetes cluster
---
Cellery runtime can be installed onto an existing k8s clusters. Cellery runtime requires MySQL server to store the control plane state, and MySQL server can be started within kubernetes with persistent volume mounted or not. 

Hence we have persistence mode and non-persisted mode. Also it requires network or local file system access to store the runtime artifacts of the cellery runtime.

Cellery installer is tested on K8s environments with following component versions.

- Kubernetes:  1.11.x  
- Docker : docker-ce 18.06.0
- MySQL Server: 5.7
- NFS Server : NFSv3 compatible

## Tested Kubernetes Providers
1. [GCP GKE](#1.-gcp-gke)

2. [Docker for Desktop on MacOS](#2.-docker-for-desktop-on-macos) 

3. [Kube Admin](#3.-kube-admin)

4. [Minikube](#4.-minikube)

### 1. GCP GKE
Cellery system can be installed into the existing GCP setup in both modes; [Persisted volume](#1.-persistent-volume) and [Non-persisted volume](#2.-non-persistent-volume).
NFS share and set of MySQL databases are required to proceed with the Cellery installation if users opted to go with [Persisted volume](#1.-persistent-volume).  

Follow below steps to configure your GCP setup to install cellery system on to it. 

1. Create GKE k8s cluster.  
**Note: Follow steps - 2, 3, 4 only if you want to create cellery system with [persisted volume](#1.-persistent-volume)** 
2. Create NFS server and  MySQL server in the same GCP compute region and the zone i.
3. User can find the [SQL script](https://raw.githubusercontent.com/wso2-cellery/distribution/master/installer/k8s-artefacts/mysql/dbscripts/init.sql) file in the Cellery distribution repository. 
Please note to replace the database username (`DATABASE_USERNAME`) and the password (`DATABASE_PASSWORD`) accordingly.
4. Import the script that was prepared in step-2 into MySQL database.

### 2. Docker for Desktop on macOS
Cellery system can be installed into the docker for desktop setup in both modes; [Persisted volume](#2.-non-persistent-volume) and [Non-persisted volume](#non-persistent-volume).
User needs to increase the Docker Desktop resources to support Cellery runtime. Minimum resources required by Cellery runtime:

 * CPU : 4 core or more
 * Memory : 8 GB or more

**Note:**
 * If users wants to install cellery with persisted volume, then add /var/tmp/cellery to the Docker Desktop file sharing to support persistence deployments.
 * User may need to restart the Docker Desktop to update the ingress-nginx EXTERNAL-IP after deploying the Cellery runtime.  ( This is a known issue in Docker Desktop.)

### 3. Kube Admin
Tested on Ubuntu 18.04. Cellery supports both [persistent](#1.-persistent-volume) and [non-persistent](#2.-non-persistent-volume) runtime deployment on kubeadm based k8s. 

**Note:**
To run in persistence mode follow the below instructions
 * Set umask to 0000 if it is not already set
 * Create a folder in **/var/tmp/cellery** and give full executable permission to create and modify the artifacts which are shared with the cellery runtime.

### 4. Minikube
Cellery only supports [non-persistence mode](#2.-non-persistent-volume) deployment on Minikube.


## Cellery setup with existing kubernetes cluster
### Interactive setup
In this cellery installation option cellery CLI uses the default kubernetes cluster configured in the $HOME/.kube/config file.
As mentioned above this can be installed with [persistent volume](#2.-non-persistent-volume) and [non-persistent](#2.-non-persistent-volume) volume. 

   i. Execute `cellery setup` command to configure Cellery runtime. This 
    will prompt a list of selections. By selecting `create ` section users can setup the Cellery runtime: 
   ```
     $ cellery setup
     [Use arrow keys]
     ? Setup Cellery runtime
         Manage
       ➤ Create
         Modify
         Switch
         EXIT
   ```
   
   ii. From the selections available for environment type, select `Existing cluster` to proceed with local installation:
   ```
     $ ✔ Create
     [Use arrow keys]
     ? Select an environment to be installed
         Local
         GCP
       ➤ Existing cluster
         BACK
   ```

#### 1. Persistent volume
Cellery needs persistent volume to keep MySQL server files and WSO2 APIM deployable artifacts. 
This option enables to save the state of the cellery system, therefore restarting the cellery runtime will not

Once the option `Existing cluster` is selected, the CLI will prompt to select whether to use Persistent Volume or not:

```
    $ cellery setup
    ✔ Create
    ✔ Existing cluster
    [Use arrow keys]
    ? Select the type of runtime
     ➤ Persistent volume
       Non persistent volume
       BACK
   ```

##### 1.1. Access to NFS
If the user has access to an NFS server he/she can use it as the persistent volume, else he/she can proceed with file system mount by default. 
And based on this, user should select `yes` or `no` for the using NFS server option.  

**Note: If you are trying this on docker for desktop, and you don't have NFS, then you will be required add /var/tmp/cellery to the [Docker Desktop](#2.-docker-for-desktop-on-macos) file sharing as mentioned.**
   
   ```
    $ cellery setup 
     ✔ Create
     ✔ Existing cluster
     ✔ Persistent volume
     Use the arrow keys to navigate: ↓ ↑ → ←
     ? Use NFS server:
      ▸ Yes
        No
        BACK
   ```
##### 1.2. Access to MySQL Server
The user can provide database username/password of the MySQL instance that's running on his environment with this step as shown below.     
   ```
     $ cellery setup
      ✔ Create
      ✔ Existing cluster
      ✔ Persistent volume
      ✔ Yes
      ? NFS server ip:  192.168.2.1
      ? File share name:  data
      ? Database host:  192.168.2.100
      
      ? Mysql credentials required
      Username: mysqlroot
      Password:
      Confirm Password:
   ```
   
Once above are performed, there will be an option to select `Basic` or `Complete` [installation packages](../../README.md#cellery-runtime-packages). 
Now continue to [configure host entries](#configure-host-entries) to complete the setup. 

#### 2. Non-Persistent Volume
This mode allows users to start cellery system without any need for access to NFS/File system or MySQL database storage. 
But this will not preserve the state of the cellery system, and once the cellery system is restarted any changes made during runtime will be lost, and observability and APIM changes also will be not be stored.
This is ideal for development and quick test environments. 

   i. Select the option `Non persistent volume` and continue.
   ```
     $ cellery setup
     ✔ Create
     ✔ Existing cluster
     [Use arrow keys]
     ? Select the type of runtime
         Persistent volume
       ➤ Non persistent volume
   ```
   
   ii. The select the setup type and continue.

   ```
    $ cellery setup
    ✔ Create
    ✔ Existing cluster
    ✔ Non persistent volume
     [Use arrow keys]
     ? Select the type of runtime
       ➤ Basic
         Complete
   ```
   
### Non-Interactive Mode
|Persistent volume	|Access to NFS storage	| Command | Description|
|-------------------|-----------------------|---------|------------| 
|No	|N/A| cellery setup create existing [--complete] [--loadbalancer]| By default basic setup will be created, if `--complete` flag is passed, then complete setup will be created. If k8s cluster supports cloud based loadbalancer (e.g: GCP, Docker-for-mac), users have to pass `--loadbalancer` flag.| 
|Yes| No |cellery setup create existing --persistent [--complete] [--loadbalancer]| In this case, the file system should be mounted or should be accessible by the k8s cluster. By default basic setup will be created, if `--complete` flag is passed, then complete setup will be created. If k8s cluster supports cloud based loadbalancer (e.g: GCP, Docker-for-mac), users have to pass `--loadbalancer` flag.| 
|Yes |Yes| cellery setup create existing [--complete] [--dbHost <DB_HOST> --dbUsername <DB_USER_NAME> --dbPassword <DB_PASSWORD> --nfsServerIp <IP_ADDRESS> --nfsFileShare <FILE_SHARE>] [--loadbalancer]| In this case, the external database and NFS server available and k8s cluster can be connected to those to provide the persisted functionality. This is the recommended mode for the production deployment.| 

## Configure host entries 
Once the setup is complete, cellery system hostnames should be mapped with the ip of the ingress. 

For this purpose, its currently assumed that the K8s cluster has a nginx ingress controller functioning. 

Run the following kubectl command to get the IP address. 

   ```
    kubectl get ingress -n cellery-system
   ```
    
**Note: Due to some known issue in the nginx-ingress with docker for desktop, users will need to restart the docker for 
desktop for the IPs to appear in the `kubectl get ingress` command. **

   Then update the /etc/hosts file with that Ip as follows. 
     
   ```
    <IP Address> wso2-apim cellery-dashboard wso2sp-observability-api wso2-apim-gateway cellery-k8s-metrics idp.cellery-system pet-store.com hello-world.com my-hello-world.com
   ```
**Note:**

**1. IP Address for docker for desktop is `127.0.0.1`, and minikube is `192.168.99.100`.** 

**2. In some pre-configured setups, (ex.: setups created with kubeam command line tool), it might be required to specifically find the publicly exposed IP(s) of the nodes and update the ingress-nginx kubernetes service's `externalIPs` section by using `kubectl edit` command.**

## Trying Out
Once the installation process is completed, you can try out [quick start with cellery](../../README.md#quick-start-with-cellery).

## Cleaning Up
Please refer readme for [managing cellery runtimes](./manage-setup.md) for details on how to clean up the setup.

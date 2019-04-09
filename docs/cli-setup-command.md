## Cellery setup
This command allows to configure and manage the cellery runtimes. This command can be executed both interactive mode and inline mode. 
Mainly there are four operations performed in the cellery setup as below.  

1. [Create](#create): This allows users to create the cellery runtime. The cellery runtime can be installed as Local, 
GCP or any existing kubernetes cluster.  
2. [Manage](#manage): This option allows the users to start/stop the cellery runtime, and cleanup the installed cellery runtime. 
3. [Modify](#modify): This option allows the users to modify the current setup. By default cellery run time can be installed in two packages; Basic and Complete.
Users can add/remove components (Observability and APIM) selectively to these packages based on their requirement.  
4. [Switch](#switch): This option shows the list of current kubernetes clusters that are configured, and allows users to switch the context between the clusters. By performing this,
the current context of `kubectl` also will be updated, and cellery commands will be performed on the switched kubernetes context. 
    
### Create

### Manage

### Modify

### Switch

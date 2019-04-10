# Switch between runtimes 

Manage command is used to uninstall Cellery from system.  
This will remove the configurations from  `cellery-system`,`istio-system` namespace along with persisted data. 

## Usage

### Interactive CLI mode

1. Run `cellery setup` command

    ```bash
    cellery setup
    ```

1. Select `Switch` option

    ```text
    cellery setup
    [Use arrow keys]
    ? Setup Cellery runtime
        ➤Manage
        Create
        Modify
        Switch
        EXIT
    ```
1. Select the runtime that needs to be switched. 
    ```text
   cellery setup
   ✔ Switch
   [Use arrow keys]
   ? Select a Cellery Installed Kubernetes Cluster
     ➤ docker-desktop
       docker-for-desktop
       gke_vick-team_us-west1-b_cellery-cluster87
       kubernetes-admin@kubernetes
   ↓   kubernetes-admin@kubernetes-remote
    ```

1. CLI will switch the kubernetes cluster context.
    ```text
    cellery setup
    ✔ Switch
    Selected cluster: docker-desktop
    
    ✔ Successfully configured Cellery.
    
    What's next ?
    ======================
    To create your first project, execute the command:
      $ cellery init
    ```

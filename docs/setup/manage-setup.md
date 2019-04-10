# Cleanup Cellery system components

Manage command is used to uninstall Cellery from system.  
This will remove the configurations from  `cellery-system`,`istio-system` namespace along with persisted data. 

## Usage

### Interactive CLI mode

1. Run `cellery setup` command

    ```bash
    cellery setup
    ```

1. Select `Manage` option

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
1. Select the runtime that needs to be uninstalled. 
    ```text
    ✔ Manage
    [Use arrow keys]
    ? Select a runtime
        Local
        GCP
        Existing cluster
      ➤ BACK
    ```

1. Select `cleanup` option.
    ```text
    cellery setup
    ✔ Manage
    ✔ Existing cluster
    [Use arrow keys]
    ? Select `cleanup` to remove an existing GCP cluster
      ➤ cleanup
        BACK
    ```
    
1. Confirm clean up option.
    ```text
    cellery setup
    ✔ Manage
    ✔ Existing cluster
    ✔ cleanup
    Use the arrow keys to navigate: ↓ ↑ → ←
    ? Do you want to delete the cellery runtime (This will delete all your cells and data):
      ▸ Yes
        No
    ```   
    
1. Select whether to remove `istio` or not.
    ```text
    cellery setup
    ✔ Manage
    ✔ Existing cluster
    ✔ cleanup
    ✔ Yes
    Use the arrow keys to navigate: ↓ ↑ → ←
    ? Remove istio:
      ▸ Yes
        No
    ```     
    
1. `Cellery` will be uninstalled from the selected environment. If local vm setup is used clean up, will remove the vm.
     ```text
    cellery setup
    ✔ Manage
    ✔ Existing cluster
    ✔ cleanup
    ✔ Yes
    ✔ Yes
    ✔ Cleaning up cluster 
    ```

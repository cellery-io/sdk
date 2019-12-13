# Enabling/Disabling Cellery system components

Cellery allows you to enable/disable different system components so that you can modify your runtime for your requirements.

**Note:** You may need to redeploy the cells which uses OIDC Dynamic client registration when you enable/disable API Management

## Usage

### Interactive method

#### Start modify setup
1. Run `cellery setup` command.

    ```bash
    cellery setup
    ```
2. Select modify option.

    ```text
    cellery setup
    [Use arrow keys]
    ? Setup Cellery runtime
        Create
        Manage
      ➤ Modify
        Switch
        EXIT
    ```
#### Enable/Disable APIM
If you want to enable/disable API Management with Global gateway, go into `API Manager` option and select `enable` or `disable`. Select `Yes, Apply changes` to apply the change. 
See [inline commands](#inline-method) to perform this with one step. 
```text
    cellery setup
    ✔ Modify
    ✔ API Manager
    ✔ Disable
    [Use arrow keys]
    ? Have you finished modifying the runtime
      ➤ Yes, Apply changes
        No, Modify another component
```

#### Enable/Disable Observability
If you want to enable/disable Observability, go into `Observability` option and select `enable` or `disable`. Select `Yes, Apply changes` to apply the change.
See [inline commands](#inline-method) to perform this with one step.
```text
 cellery setup
 ✔ Modify
 ...
 ✔ Observability
 [Use arrow keys]
 ? Observability
   ➤ Enable
     BACK
```
   
#### Enable/Disable Autoscaler
If you want to enable/disable Autoscaler, go into `Zero-Scaling` and/or `Horizontal Pod Autoscaler` options and select `enable`
or `disable`. Select `Back` after the selection. You can select either `Zero-Scaling` or `Horizontal Pod Autoscaler` or enable both at the same time. 
Once this is enabled, based on the configurations provided in the components, the scaling policies will be applied. Select `Yes, Apply changes` to apply the change.
and come to modify command main menu options. See [inline commands](#inline-method) to perform this with one step. 
```text
 cellery setup
 ✔ Modify
 ....
 ✔ Autoscaler
 [Use arrow keys]
 ? Select a runtime component
 ➤ Scale-to-Zero
   Horizontal Pod Autoscaler
   BACK
```
    
#### Complete modify setup
1. Once you select a change you will be asked whether to apply the change or whether to modify another component. Select `Yes, Apply changes` to apply the changes.
```text
 cellery setup
 ✔ Modify
 ✔ API Manager
 ✔ Disable
 [Use arrow keys]
 ? Have you finished modifying the runtime
   ➤ Yes, Apply changes
     No, Modify another component
 ```
    
2. The prompt will be appearing with components that is being enabled and disabled, and asking a confirmation to continue. 
Validate your changes by reviewing the list, and select `yes` to continue. 
```text
 ✔ Modify
 ✔ API Manager
 ✔ Disable
 ✔ No, Modify another component
 ✔ Observability
 ✔ Disable
 ✔ No, Modify another component
 ✔ Autoscaler
 ✔ Scale-to-Zero
 ✔ Disable
 ✔ Yes, Apply changes
 Following modifications to the runtime will be applied
 Disabling : API Manager, Observability, Scale-to-Zero
 Use the arrow keys to navigate: ↓ ↑ → ←
 ? Do you wish to continue:
   ▸ Yes
     No
``` 
   
### Inline method

1. Run `cellery setup modify --help` to see the available options

    ```text
    Modify Cellery runtime

    Usage:
      cellery setup modify <command> [flags]

    Examples:
      cellery setup modify --apim=enable --observability=disable

    Flags:
          --apim string            enable or disable API Management in the runtime
      -h, --help                   help for modify
          --hpa string             enable or disable hpa in the runtime
          --observability string   enable or disable observability in the runtime
          --scale-to-zero string   enable or disable scale to zero in the runtime

    Global Flags:
      -k, --insecure   Allow insecure server connections when using SSL
      -v, --verbose    Run on verbose mode
    ```

2. Use appropriate flags to enable/disable system components

    ```text
    # Enable API Management and Observability
    cellery setup modify --apim=enable --observability=enable

    # Enable horizontal autoscaler and disable scale to zero
    cellery setup modify --hpa=enable --scale-to-zero=disable
    ```

## What's Next?
- [Developing a Cell](../writing-a-cell.md) - step by step explanation on how you could define your own cells.
- [Scale up/down](../cell-scaling.md) - scalability of running cell instances with zero scaling and horizontal autoscaler.
- [Observability](../cellery-observability.md) - provides the runtime insight of cells and components.
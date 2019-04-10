# Enabling/Disabling Cellery system components

Cellery allows you to enable/disable different system components so that you can modify your runtime for your requirements.
This will only modify the configurations on your `cellery-system` namespace.


## Usage

### Interactive CLI mode

1. Run `cellery setup` command

    ```bash
    cellery setup
    ```
2. Select modify option

    ```text
    cellery setup
    [Use arrow keys]
    ? Setup Cellery runtime
        Manage
        Create
      ➤ Modify
        Switch
        EXIT
    ```
3. Select whether you want to enable API Management with Global gateway
    ```text
    cellery setup
    ✔ Modify
    [Use arrow keys]
    ? API management and global gateway
      ➤ Enable
        Disable
        BACK
    ```

3. Select whether you want to enable Observability
    ```text
    cellery setup
    ✔ Modify
    ✔ Enable
    [Use arrow keys]
    ? Observability
      ➤ Enable
        Disable
        BACK
    ```
 
### Non-interactive CLI mode

1. Run `cellery setup modify --help` to see the available options

    ```text
    cellery setup modify --help
    Modify Cellery runtime
    
    Usage:
      cellery setup modify <command> [flags]
    
    Flags:
          --apimgt          enable API Management in the runtime
      -h, --help            help for modify
          --observability   enable Observability in the runtime
    ```

2. User appropriate flag to enable/disable system components

    ```text
    # Enable API Management and Observability
    cellery setup modify --apimgt --observability
    
    # Enable Observability without API Management
    cellery setup modify --observability
    
    ```

### Feature comparison with system components

| Flag --apimgt | Flag --observability | API Management  | Identity Provider | Observability |
|:-------------:|:--------------------:|:---------------:|:-----------------:|:-------------:|
| Disabled      | Disabled             | ✘               | ✔                 | ✘             |
| Disabled      | Enabled              | ✘               | ✔                 | ✔             |
| Enabled       | Disabled             | ✔               | ✔                 | ✘             |
| Enabled       | Enabled              | ✔               | ✔                 | ✔             |

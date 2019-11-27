## Cellery runtime installation options

### Basic vs. complete installations
When it comes to installing Cellery, you could choose between the basic or complete installation options. When you run the `cellery setup` command, 
you can select to either install the basic or complete version.
The basic installation would require less resources in comparison to the complete installation.
The following table offers a comparison between what you would get.

| Packages | Components | Supported Functionality | 
|----------|------------|-------------------------|
| Basic | <ul><li>Cell controller</li><li>Light weight Identity Provider</li></ul>|<ul><li>HTTP(S) cells with local APIs</li><li>Full support for web cells</li><li>Inbuilt security for inter cell and intra cell communication</li></ul> |
| Complete | <ul><li>Cell controller</li><li>Global API manager</li><li>Observability portal and components</li></ul>| <ul><li>Full HTTP(S) cells with local/global APIs</li><li>Full support for web cells</li><li>Inbuilt security for inter cell and intra cell communication</li><li>API management functionality</li><li>observability of cells with rich UIs</li></ul> |

### Local, existing Kubernetes cluster and GCP based installations

You could opt to install Cellery either in local mode, or on to an existing Kubernetes cluster or on Google Cloud.
Note that you could combine either the basic or complete installations along with these options.

When you run `cellery setup` in interactive mode, and select `Create`, you can select one of the following options:

* `Local` -> to install locally via a pre-built VM 
* `Existing cluster` -> to install into an existing K8s cluster (e.g. Docker for Desktop)
* `GCP` -> to install on Google Cloud

The following links explain in detail the setup options for different environments.

* [1. Local setup](setup/local-setup.md)
    * [1.1. Interactive method](setup/local-setup.md#interactive-method)
    * [1.2. Inline method](setup/local-setup.md#inline-method)
* [2. Existing Cluster](setup/existing-cluster.md)
    * [2.1. Interactive method](setup/existing-cluster.md#interactive-method)
    * [2.2. Inline method](setup/existing-cluster.md#inline-method)
* [3. GCP setup](setup/gcp-setup.md) 
    * [3.1. Interactive method](setup/gcp-setup.md#interactive-method)
    * [3.2. Inline method](setup/gcp-setup.md#inline-method)

## What's Next?
- [Developing a Cell](writing-a-cell.md) - step by step explanation on how you could define your own cells
- [Cell Specification](https://github.com/wso2/cellery-spec/blob/master/README.md) - Key concepts of Cellery.
- [How to code cells?](cellery-syntax.md) - explains how Cellery cells are written.
- [CLI commands](cli-reference.md) - reference for CLI commands.
- [Samples](https://github.com/wso2/cellery-samples/tree/master) - a collection of useful samples.
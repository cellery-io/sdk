## Cellery runtime installation options

### Basic vs. complete installations
When it comes to installing Cellery, you could choose between the basic or complete installation options.
When you run the `cellery setup` command, you can choose to either install the basic or complete version.
The basic installation would require less resources in comparison to the complete installation.
The following table offers a comparison between what you would get.

| Packages | Components | Supported Functionality | 
|----------|------------|-------------------------|
| Basic | <ul><li>Cell controller</li><li>Light weight Identity Provider</li></ul>|<ul><li>HTTP(S) cells with local APIs</li><li>Full support for web cells</li><li>Inbuilt security for inter cell and intra cell communication</li></ul> |
| Complete | <ul><li>Cell controller</li><li>Global API manager</li><li>Observability portal and components</li></ul>| <ul><li>Full HTTP(S) cells with local/global APIs</li><li>Full support for web cells</li><li>Inbuilt security for inter cell and intra cell communication</li><li>API management functionality</li><li>observability of cells with rich UIs</li></ul> |

### Local, existing Kubernetes cluster and GCP based installations

You could opt to install Cellery either in local mode, on to an existing Kubernetes cluster or on Google Cloud.
Note that you could combine either the basic or complete installations along with these options.

When you run `cellery setup` in interactive mode, and select `Create`, you can select one of the following options:

* `Local` -> to install locally via a pre-built VM 
* `Existing cluster` -> to install into an existing K8s cluster (e.g. Docker for Desktop)
* `GCP` -> to install on Google Cloud

The following links explain in detail the setup options for different environments.

* [1. Local setup](docs/setup/local-setup.md)
* [2. Existing Cluster](docs/setup/existing-cluster.md)
* [3. GCP setup](docs/setup/gcp-setup.md) 
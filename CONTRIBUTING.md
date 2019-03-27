## Contribute to Cellery
The Cellery Team is pleased to welcome all contributors willing to join with us in our journey. 
Cellery project is divided into few repositories as explained below. 
- [wso2-cellery/sdk](https://github.com/wso2-cellery/sdk/) - This repository contains the cellery specification 
implementation, 
CLI implementation, and installers for different operating systems.  
- [wso2-cellery/distribution](https://github.com/wso2-cellery/distribution/) - This repository contains kubernetes 
artifacts 
for cellery mesh runtime, and docker image generation for Global and cell API Manager.   
- [wso2-cellery/mesh-controller](https://github.com/wso2-cellery/mesh-controller/) - This repository includes the 
controller 
for cell CRD in kubernetes.  
- [wso2-cellery/mesh-security](https://github.com/wso2-cellery/mesh-security) - This includes cell based control plane 
components such as STS, and other related security functionality for overall cellery mesh.  
- [wso2-cellery/mesh-observability](https://github.com/wso2-cellery/mesh-observability) - This repository includes the 
observability related components such as WSO2 stream processor extensions, siddhi applications for tracing, telemetry 
processing, dependency model generation, observability portal, etc, and docker images for observability control plane.  

### Build from Source

#### Prerequisites 
- JDK 1.8 
- Ballerina 0.990.3 
- Go 1.11.2 or higher
- Apache Maven 3.5.2 or higher
- NPM 6.9.0+
- Git

#### Steps
1. Clone the repository to GOPATH.
```
$ mkdir -p $GOPATH/src/github.com/wso2-cellery/
$ cd $GOPATH/src/github.com/wso2-cellery/
$ git clone https://github.com/celleryio/sdk.git
```
2. Building and installing the Ballerina language extensions.
```
$ cd $GOPATH/src/github.com/wso2-cellery/sdk
$ make install-lang
```
3. Building and installing the Cellery CLI.
```
$ cd $GOPATH/src/github.com/wso2-cellery/sdk
$ make install-cli
```

### Issue Management
We use GitHub issues to track all of our bugs and feature requests. Please feel free to open an issue about any 
question, bug report or feature request that you have in mind. It will be ideal to report bugs in the relevant 
repository as mentioned in above, but if you are not sure about the repository, you can create issues to [wso2-cellery/sdk]
(https://github.com/wso2-cellery/sdk/) repository, and we’ll analyze the issue and then move it to relevant repository. 
We also welcome any external contributors who are willing to contribute. You can join a conversation in any existing issue and even send PRs to contribute
Each issue we track has a variety of metadata which you can select with labels:

- Type: This represents the kind of the reported issues such as Bug, New Feature, Improvement, etc. 
- Priority: This represents the importance of the issue, and it can be scaled from High to Normal.
- Severity: This represents the impact of the issue in your current system. If the issue is blocking your system, 
and it’s having an catastrophic effect, then you can mark is ‘Blocker’. The ‘Blocker’ issues are given high priority 
as well when we are resolving the issues. 

Additional to the information provided above, the issue template added to the repository will guide you to describe 
the issue in detail therefore we can analyze and work on the resolution towards it. Therefore we appreciate to fill the 
fields mostly as possible when you are creating the issue. We will evaluate issues, and based on the label provided 
details and labels, and will allocate to the milestones. 


## Cellery Commands
### Init
Initializes a cellery project. In the command prompt, provide the project name.
Usage : 
```
cellery init
```

### Build
Build an immutable cell image with required dependencies.  
Usage : 
```
cellery build <BAL_FILE_NAME> -t <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION>
``` 
Example : 
```
cellery build my-project.bal wso2/my-cell:1.0.0
```

### Run
Use a cellery image to create a  running instance.You can also run the cell image with a name by passing -n parameter, 
and this is an optional field. 
Usage : 
```
cellery run <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION>  
cellery run <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION> -n <NAME>
``` 
Example : 
``` 
cellery run wso2/my-cell:1.0.0 
cellery run wso2/my-cell:1.0.0 -n my-cell-name1
```

### Run linking to dependencies
Use a cellery image to create a running instance specifying an instance name.
Input the same instance name to establish the link between the running cell instance and another cell which depends on it.  
Usage : 
```
cellery run <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION> -l <ALIAS>:<NAME-1>
```
Example : 
```
cellery run wso2/cell2:1.0.1 -n mycell-2 -l cell-1:mycell-1
```

### Run with dependencies
Use a cellery image to create a running instance specifying an instance name along with its dependencies. (The images will be pulled from the Registry if required)
Input the same instance name to establish the link between the running cell instance and another cell which depends on it.
Any unspecified (no links provided) can be shared (if the Cell Image is the same) by using the `--share-instances` flag.
You will be prompted the confirm the instances and the dependency tree before starting the instances.
Usage : 
```
cellery run <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION> --with-dependencies
cellery run <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION> --with-dependencies --share-instances
cellery run <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION> --with-dependencies -l <PARENT-INSTANCE>.<ALIAS>:<NAME-1>
```
Example : 
```
cellery run wso2/cell2:1.0.1 -n mycell-2 -l cell-1:mycell-1
cellery run wso2/cell2:1.0.1 -n mycell-2 --with-dependencies
cellery run wso2/cell2:1.0.1 -n mycell-2 --with-dependencies --share-instances
```

### List instances
List all running cells.  
Usage : 
```
cellery list instances
```

### Login
Login to the Cellery Registry and save the credentials.
Usage :
```
cellery login [REGISTRY]
```
Example :
```
cellery login
```

### Logout
Remove the saved credentials.
Usage :
```
cellery logout [REGISTRY]
```
Example :
```
cellery logout
```

### Push
Push cell image to the remote repository.  
Usage : 
```
cellery push <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION>
```
Example : 
```
cellery push wso2/my-cell:1.0.0
```

### Pull 
Pulls cell image from the remote repository.  
Usage : 
```
cellery pull <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION>
```
Example : 
```
cellery pull wso2/my-cell:1.0.0
```

### List images
List cell images that was pulled and built in the current machine.  
Usage : 
```
cellery list images
```

### View image
View the graphical representation of a Cell Image.
Usage :
```
cellery view <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION>
```
Example :
```
cellery view wso2/my-cell:1.0.0
```

### Terminate
Terminate a running cell instance.
Usage : 
```
cellery terminate <CELL_NAME>
```
Example : 
```
cellery terminate hello-cell
```

### Status 
Performs a health check of a cell.  
Usage : 
```
cellery status <CELL_NAME>
```
Example : 
```
cellery status hello-cell
```

### List ingresses
List the exposed APIs of a cell instance.  
Usage : 
```
cellery list ingresses <CELL_NAME>
```
Example : 
```
cellery list ingresses hello-cell
```

### Logs 
Displays logs for either the cell instance, or a component of a running cell instance.  
Usage : 
```
cellery logs <CELL_NAME> / cellery logs <CELL_NAME> -c <COMPONENT_NAME>
```
Example: 
```
cellery logs hello-cell / cellery logs hello-cell -c my-component
```

### List Components
Lists the components which the cell encapsulates.  
Usage : 
```
cellery list components <CELL_NAME>
```  
Example : 
```
cellery list components hello-cell
```

### inspect
List the files (directory structure) of a cell images.  
Usage : 
```
cellery inspect <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION>
```
Example : 
```
cellery inspect wso2/my-cell:1.0.0
```

### Extract-resources
Extract the resource files of a pulled image to the provided location. This is useful when cells are packaged with 
swagger files, therefore any components or micro services that uses the cell can generate the code from the swagger.  
Usage : 
```
cellery extract-resources <ORGANIZATION_NAME>/<IMAGE_NAME>:<VERSION> ./resources
```
Example : 
```
cellery extract-resources cellery-samples/employee:1.0.0 ./resources
```

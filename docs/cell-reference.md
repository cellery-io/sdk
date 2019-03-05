## Cellery Language Syntax
The Cellery Language is a subset of Ballerina and Ballerina extensions. Therefore the language syntax of Cellery 
resembles with normal Ballerina language syntax.

#### Cell
A cell is a collection of components, APIs, Ingresses and Policies. A Cell object is initialized as follows;
```
cellery:CellImage cellObj = new ();
```
Components can be added to a cell by calling addComponent(Component comp) method.
```
cellObj.addComponent(comp);
```
#### Component
A component represents an implementation of the business logic (in a docker image or a legacy system) and a collections 
of network accessible entry points (Ingresses) and parameters. A sample component with inline definitions would be as follows:
```
cellery:Component stock = {
   name: "stock",
   source: {
       image: "docker.io/celleryio/sampleapp-stock"
   },
   ingresses: {
       stock: new cellery:HTTPIngress(8080,
           "stock",
           [
               {
                   path: "/options",
                   method: "GET"
               }
           ]
       )
   }
};
```

#### Ingresses
An Ingress represents an entry point into a Cell or Component. An ingress can be an HTTP or a TCP endpoint.

##### HTTP Ingresses
HTTP ingress supports defining HTTP definitions inline or as a swagger file.
A sample ingress instance  with inline API definition would be as follows:
```
cellery:Ingress ingressA = new cellery:HTTPIngress(
    8080,      //port
    "foo",        //context
     [        // Definitions
            {
                    path: "*",
                    method: "GET,POST,PUT,DELETE"
            }
        ]
};
```
A sample ingress instance  with swagger 2.0 definition would be as follows:
```
cellery:Ingress ingressA = new cellery:HTTPIngress(
    8080,
    "foo",
    “./resources/employee.swagger.json”
);
```
##### TCP Ingresses
TCP ingress supports defining TCP endpoints. A sample TCP ingress would be as following:
```
cellery:Ingress ingressA = new cellery:TCPIngress(
    8080,      // port
    9090,      // target port
};
```

The port is the actual container port which is exposed by the container. The targetPort is the port exposed by the cell gateway.

#### Parameters
A cell developer can require a set of parameters that should be passed to a Cell instance for it to be properly functional. 

```
cellery:Component employeeComponent = {
   name: "employee",
   source: {
       image: "docker.io/celleryio/sampleapp-employee"
   },
   ingresses: {
       employee: new cellery:HTTPIngress(
                     8080,
                     "employee",
                     "./resources/employee.swagger.json"
       )
   },
   parameters: {
       SALARY_HOST: new cellery:Env(),
       PORT: new cellery:Env(default = 8080)
   },
   labels: {
       cellery:TEAM:"HR"
   }
};
```

Note the parameters SALARY_HOST and PORT in the Cell definition above. This parameters can be set in the build time of this Cell:

```
public function build(string orgName, string imageName, string imageVersion) {
   // Build EmployeeCell
   io:println("Building Employee Cell ...");
   // Map component parameters
   cellery:setParameter(employeeComponent.parameters.SALARY_HOST, "http://salaryservice.com");
   // Add components to Cell
   employeeCell.addComponent(employeeComponent);
   //Expose API from Cell Gateway
   employeeCell.exposeLocal(employeeComponent);

   _ = cellery:createImage(employeeCell, orgName, imageName, imageVersion);
}
```

#### APIs
An API represents a defined set of functions and procedures that the services of the Components inside a Cell exposes 
as resources (i.e ingresses). An Ingress can be exposed as an API using exposeLocal method.
```
cellery:Component stock = {
   name: "stock",
   source: {
       image: "docker.io/celleryio/sampleapp-stock"
   },
   ingresses: {
       stock: new cellery:HTTPIngress(8080,
           "stock",
           [
               {
                   path: "/options",
                   method: "GET"
               }
           ]
       )
   }
};

cellery:CellImage stockCell = new();

public function build(string orgName, string imageName, string imageVersion) {
   //Build Stock Cell
   io:println("Building Stock Cell ...");
   stockCell.addComponent(stock);
   //Expose API from Cell Gateway
   stockCell.exposeLocal(stock);
   _ = cellery:createImage(stockCell, orgName, imageName, imageVersion);
}
```

#### Autoscaling

Autoscale policies can be specified by the Cell developer at Cell creation time. 

Cell component with scale policy at component level:

```
import ballerina/io;
import celleryio/cellery;

cellery:AutoscalingPolicy policy1 = {
   minReplicas: 1,
   maxReplicas: 2,
   metrics: [
       { new cellery:CpuUtilizationPercentage(50) }
   ]
};

//Stock Component
cellery:Component stock = {
   name: "stock",
   source: {
       image: "docker.io/celleryio/sampleapp-stock"
   },
   ingresses: {
       stock: new cellery:HTTPIngress(8080,
           "stock",
           [
               {
                   path: "/options",
                   method: "GET"
               }
           ]
       )
   },
   autoscaling: { policy: policy1, overridable: false }
};

cellery:CellImage stockCell = new();

public function build(string orgName, string imageName, string imageVersion) {
   //Build Stock Cell
   io:println("Building Stock Cell ...");
   stockCell.addComponent(stock);
   //Expose API from Cell Gateway
   stockCell.exposeLocal(stock);
   _ = cellery:createImage(stockCell, orgName, imageName, imageVersion);
}
```

The autoscale policy defined by the developer can be overriden at the runtime by providing a different policy at the runtime. 

### Intra Cell Communication

Cell components can communicate with each other. This is achieved via parameters. Two components can be linked via 
parameters in the build method. For an example, consider the scenario below. The employee component expects 
two parameters SALARY_HOST and PORT, which are the hostname and port of salary component.

Employee component:
```
cellery:Component employeeComponent = {
   name: "employee",
   source: {
       image: "docker.io/celleryio/sampleapp-employee"
   },
   ingresses: {
       employee: new cellery:HTTPIngress(
                     8080,
                     "employee",
                     "./resources/employee.swagger.json"
       )
   },
   parameters: {
       SALARY_HOST: new cellery:Env(),
       PORT: new cellery:Env(default = 8080)
   },
   labels: {
       cellery:TEAM:"HR"
   }
};
```

Salary component: 
```
cellery:Component salaryComponent = {
   name: "salary",
   source: {
       image: "docker.io/celleryio/sampleapp-salary"
   },
   ingresses: {
       SalaryAPI: new cellery:HTTPIngress(
               8080,
               "payroll",
               [{
                     path: "salary",
                     method: "GET"
               }]
           )
   }
};
```

These two parameters are provided in the build method as shown below, which enables the employee component to 
communicate with the salary component. 
```
public function build(string orgName, string imageName, string imageVersion) {
   // Build EmployeeCell
   io:println("Building Employee Cell ...");
   // Map component parameters
   cellery:setParameter(employeeComponent.parameters.SALARY_HOST, cellery:getHost(imageName, salaryComponent));
   // Add components to Cell
   employeeCell.addComponent(employeeComponent);
   //Expose API from Cell Gateway
   employeeCell.exposeLocal(employeeComponent);

   _ = cellery:createImage(employeeCell, orgName, imageName, imageVersion);
}
```

### Inter Cell Communication

In addition to components within a cell, Cells themselves can communicate with each other. This is also achieved 
via parameters. Two cells can be linked via parameters in the run method. 

When a Cell Image is built it will generate a reference file describing the APIs that are exposed by itself. 
This cell reference will be installed locally, either when you build or pull the image. This reference can be imported 
in another Cell definition which is depending on the former, and can be used to link the two Cells at the runtime. 

Sample generated reference file: 
```
# Reference to the stock cell image instance which can be used to get details about the APIs
# exposed by the the cell. This information is only relevant for calls within the cellery system.
# + instanceName - The name of the instance of the cell image to be referred to

public type StockReference object {
   string cellName = "stock";
   string cellVersion = "1.0.0";
   string instanceName = "";

   public function __init(string instanceName) {
       self.instanceName = instanceName;
   }

   # Get the Host name of the Cell Gateway. This hostname can be used when accessing the APIs.
   #
   # + return - The host name of the Cell Gateway
   public function getHost() returns string {
           return self.instanceName + "--gateway-service";
   }

   # Get the complete URL of the "stock" API exposed by the cell gateway.
   # This URL can be used when accessing the "stock" API.
   #
   # + return - The URL of the "stock" API
   public function getStockApiUrl() returns string {
       return self.getStockApiProtocol() + "://" + self.getHost() + ":"
           + self.getStockApiPort()
           + self.getStockApiContext();
   }

   # Get the protocol of the "stock" API exposed by the cell gateway.
   # This protocol can be used when accessing the "stock" API.
   #
   # + return - The protocol of the "stock" API
   public function getStockApiProtocol() returns string {
       return "http";
   }

   # Get the port which exposes the "stock" API exposed by the cell gateway.
   # This port can be used when accessing the "stock" API.
   #
   # + return - The port which exposes the "stock" API
   public function getStockApiPort() returns int {
       return 80;
   }

   # Get the context of the "stock" API exposed by the cell gateway.
   # This context path can be used when accessing the "stock" API.
   #
   # + return - The context of the "stock" API
   public function getStockApiContext() returns string {
       return "stock";
   }
};
```

If a cell component wants to access the StockAPI, it can be done as below:

```
import myorg/stock;
import ballerina/io;
import celleryio/cellery;

//HR component
cellery:Component hrComponent = {
  name: "hr",
  source: {
      image: "docker.io/wso2vick/sampleapp-hr"
  },
  ingresses: {
      hr: new cellery:HTTPIngress(8080, "info",
          [
              {
                  path: "/",
                  method: "GET"
              }
          ]
      )
  },
  parameters: {
      stockgw_url: new cellery:Env()
  }
};

// Cell Initialization
cellery:CellImage hrCell = new();

// build method for cell
public function build(string orgName, string imageName, string imageVersion) {
  // Build HR cell
  io:println("Building HR Cell ...");
  hrCell.addComponent(hrComponent);
  // Expose API from Cell Gateway & Global Gateway
  hrCell.exposeGlobalAPI(hrComponent);
  _ = cellery:createImage(hrCell);
}
```

Note the run method below, which takes a variable argument list for the references of the dependency cells. 
These are names of already deployed cell instances, which will be used to retrieve the parameters and link with this 
cell instance. As an example, the stockRef.getHost() method in the example below returns the host name of the running 
stock cell instance, which is then used to link this HR cell instance with it.
```
public function run(string imageName, string imageVersion, 
string instanceName, string... dependenciesRef) {
   stock:StockReference stockRef = new(dependenciesRef[0]);
   cellery:setParameter(hrComponent.parameters.stockgw_url, stockRef.getHost());
   hrCell.addComponent(hrComponent);
   _ = cellery:createInstance(hrCell, imageName, imageVersion, instanceName);
}
```

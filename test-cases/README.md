## Samples
### HR Application
Employee Portal sample demonstrates an application which uses the cell-based architecture. This application contains 
four microservices deployed across three different cells based on their responsibilities, as explained below.

**HR Cell**  
- Contains one microservice written in ballerina.
- This microservice invokes Employee cell and Stock cell to build the employee information and returns the result.
- Exposes an API at the global Gateway which can be consumed by any API consumer from outside Cellery.

**Employee cell**  
- Contains two microservices (employee and salary) written in ballerina.
- Employee microservice returns employee details such as name, designation.
- Salary microservice returns employee salary details.
- Exposes a cell level API, when invoked returns the aggregated response from employee and salary response.

**Stock cell**  
- Contains one microservice written in ballerina.
- This microservice returns the stock related data for the employee.
- Exposes a cell level API, when invoked returns the response from stock microservice.

The detailed steps to run this sample is available in [employee-portal](employee-portal/README.md).

### Product Review Application
A multi-cell application which demonstrates multiple protocol support. 

See [product-review](product-review/README.md) for more information.

### Auto Scaling
This is a single cell application which demonstrates the autoscaling of cell components. 
While Cellery is capable of using CPU, memory usage, etc. to autoscale cell components, 
this sample focuses on the former and hence is written to artificially hog cpu for demonstration purposes. 

See [pet-service](pet-service/README.md) for more information.


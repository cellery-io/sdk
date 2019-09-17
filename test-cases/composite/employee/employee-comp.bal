import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
   int containerPort = 8080;

      // Salary Component
      cellery:Component salaryComponent = {
          name: "salary",
          source: {
              image: "wso2cellery/sampleapp-salary:0.3.0"
          },
          ingresses: {
             http:<cellery:HttpsPortIngress>{port: containerPort}
          },
          labels: {
              team: "Finance",
              owner: "Alice"
          }
      };

      // Employee Component
      cellery:Component employeeComponent = {
          name: "employee",
          source: {
              image: "wso2cellery/sampleapp-employee:0.3.0"
          },
          ingresses: {
              http:<cellery:HttpsPortIngress>{port: containerPort}
          },
          envVars: {
              SALARY_HOST: {
                  value: cellery:getHost(salaryComponent)
              }
          },
          labels: {
              team: "HR"
          },
          dependencies:{
              components:[salaryComponent]
          }
      };

      cellery:Composite employeeComposite = {
          components: {
              empComp: employeeComponent,
              salaryComp: salaryComponent
          }
      };

      return cellery:createImage(employeeComposite, untaint iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:Composite employeeComposite = check cellery:constructImage(untaint iName);
    return cellery:createInstance(employeeComposite, iName, instances, startDependencies, shareDependencies);
}

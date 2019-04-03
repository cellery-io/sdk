//   Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

// Cell file for Pet Store Sample Backend.
// This Cell encompasses the components which deals with the business logic of the Pet Store

import celleryio/cellery;

// Orders Component
// This component deals with all the orders related functionality.
cellery:Component ordersComponent = {
    name: "orders",
    source: {
        image: "wso2cellery/samples-pet-store-orders"
    },
    ingresses: {
        orders: <cellery:HttpApiIngress>{
            port: 80
        }
    }
};

// Customers Component
// This component deals with all the customers related functionality.
cellery:Component customersComponent = {
    name: "customers",
    source: {
        image: "wso2cellery/samples-pet-store-customers"
    },
    ingresses: {
        customers: <cellery:HttpApiIngress>{
            port: 80
        }
    }
};

// Catalog Component
// This component deals with all the catalog related functionality.
cellery:Component catalogComponent = {
    name: "catalog",
    source: {
        image: "wso2cellery/samples-pet-store-catalog"
    },
    ingresses: {
        catalog: <cellery:HttpApiIngress>{
            port: 80
        }
    }
};

// Controller Component
// This component deals depends on Orders, Customers and Catalog components.
// This exposes useful functionality from the Cell by using the other three components.
cellery:Component controllerComponent = {
    name: "controller",
    source: {
        image: "wso2cellery/samples-pet-store-controller"
    },
    ingresses: {
        controller: <cellery:HttpApiIngress>{
            port: 80,
            context: "controller",
            expose: "global"
        }
    },
    envVars: {
        CATALOG_HOST: { value: "" },
        CATALOG_PORT: { value: 80 },
        ORDER_HOST: { value: "" },
        ORDER_PORT: { value: 80 },
        CUSTOMER_HOST: { value: "" },
        CUSTOMER_PORT: { value: 80 }

    }
};

// Cell Initialization
cellery:CellImage petStoreBackendCell = {
    components: {
        catalog: catalogComponent,
        customer: customersComponent,
        orders: ordersComponent,
        controller: controllerComponent
    }
};

// The Cellery Lifecycle Build method which is invoked for building the Cell Image
//
// iName - The Image name
// return - The created Cell Image
public function build(cellery:ImageName iName) returns error? {
    cellery:ApiDefinition controllerApiDef = (<cellery:ApiDefinition>cellery:readSwaggerFile(
                                                                         "./components/controller/resources/pet-store.swagger.json"
    ));
    cellery:HttpApiIngress controllerApi = <cellery:HttpApiIngress>(controllerComponent.ingresses.controller);
    controllerApi.definition = controllerApiDef;

    return cellery:createImage(petStoreBackendCell, iName);
}

// The Cellery Lifecycle Run method which is invoked for creating a Cell Instance.
//
// iName - The Image name
// instances - The map dependency instances of the Cell instance to be created
// return - The Cell instance
public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    petStoreBackendCell.components.controller.envVars.CATALOG_HOST.value = cellery:getHost(untaint iName.instanceName,
        catalogComponent);
    petStoreBackendCell.components.controller.envVars.ORDER_HOST.value = cellery:getHost(untaint iName.instanceName,
        ordersComponent);
    petStoreBackendCell.components.controller.envVars.CUSTOMER_HOST.value = cellery:getHost(untaint iName.instanceName,
        customersComponent);

    return cellery:createInstance(petStoreBackendCell, iName);
}

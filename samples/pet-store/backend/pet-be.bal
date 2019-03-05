
//   Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//   http://www.apache.org/licenses/LICENSE-2.0
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

// Cellery file for building backend Pet-store sample cell.

import ballerina/io;
import celleryio/cellery;

// Pet-store Components

int orderPort = 80;
int customerPort = 80;
int catalogPort = 80;
int controllerPort = 80;

// Orders Component
cellery:Component ordersComponent = {
    name: "orders",
    source: {
        image: "celleryio/samples-pet-store-orders"
    },
    ingresses: {
        orders: new cellery:HTTPIngress(orderPort,
            "orders",
            [
                {
                    path: "/orders",
                    method: "GET"
                }
            ]
        )
    }
};

// Customers Component
cellery:Component customersComponent = {
    name: "customers",
    source: {
        image: "celleryio/samples-pet-store-customers"
    },
    ingresses: {
        customers: new cellery:HTTPIngress(customerPort,
            "customers",
            [
                {
                    path: "/customers",
                    method: "GET"
                }
            ]
        )
    }
};

// Catalog Component
cellery:Component catalogComponent = {
    name: "catalog",
    source: {
        image: "celleryio/samples-pet-store-catalog"
    },
    ingresses: {
        catalog: new cellery:HTTPIngress(catalogPort,
            "catalog",
            [
                {
                    path: "/catalog",
                    method: "GET"
                }
            ]
        )
    }
};

// Controller Component
cellery:Component controllerComponent = {
    name: "controller",
    source: {
        image: "celleryio/samples-pet-store-controller"
    },
    ingresses: {
        controller: new cellery:HTTPIngress(
                      controllerPort,
                      "controller",
                      "./components/controller/resources/pet-store.swagger.json"
        )
    },
    parameters: {
        CATALOG_HOST: new cellery:Env(),
        CATALOG_PORT: new cellery:Env(),
        ORDER_HOST: new cellery:Env(),
        ORDER_PORT: new cellery:Env(),
        CUSTOMER_HOST: new cellery:Env(),
        CUSTOMER_PORT: new cellery:Env()

    }
};

cellery:CellImage petStoreCell = new();

public function build(string orgName, string imageName, string imageVersion) {
    //Build Pet-store Cell
    io:println("Building Pet-store Cell ...");

    petStoreCell.addComponent(controllerComponent);
    petStoreCell.addComponent(ordersComponent);
    petStoreCell.addComponent(catalogComponent);
    petStoreCell.addComponent(customersComponent);

    cellery:setParameter(controllerComponent.parameters.CATALOG_HOST, cellery:getHost(imageName, catalogComponent));
    cellery:setParameter(controllerComponent.parameters.CATALOG_PORT, catalogPort);
    cellery:setParameter(controllerComponent.parameters.ORDER_HOST, cellery:getHost(imageName, ordersComponent));
    cellery:setParameter(controllerComponent.parameters.ORDER_PORT, orderPort);
    cellery:setParameter(controllerComponent.parameters.CUSTOMER_HOST, cellery:getHost(imageName, customersComponent));
    cellery:setParameter(controllerComponent.parameters.CUSTOMER_PORT, customerPort);

    //Expose API from Cell Gateway
    petStoreCell.exposeLocal(controllerComponent);

    _ = cellery:createImage(petStoreCell, orgName, imageName, imageVersion);
}

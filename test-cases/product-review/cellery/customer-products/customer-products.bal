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

import ballerina/io;
import celleryio/cellery;

cellery:ApiDefinition[] customerAPIDef = [{
    path: "/*",
    method: "GET"
}];

cellery:ApiDefinition[] productsAPIDef = [{
    path: "/*",
    method: "GET"
}];

// customer Component
cellery:Component customers = {
    name: "customers",
    source: {
        image: "celleryio/samples-productreview-customers"
    },
    ingresses: {
        customerAPI: new cellery:HttpApiIngress(
                         8080,
                         "customers-1",
                         customerAPIDef
        )
    },
    parameters: {
        CATEGORIES_HOST: new cellery:Env(),
        CATEGORIES_PORT: new cellery:Env(default = 8000)
    }
};

// products Component
cellery:Component products = {
    name: "products",
    source: {
        image: "celleryio/samples-productreview-products"
    },
    ingresses: {
        productsAPI: new cellery:HttpApiIngress(
                         8080,
                         "products-1",
                         productsAPIDef
        )
    }
};

// categories Component
cellery:Component categories = {
    name: "categories",
    source: {
        image: "celleryio/samples-productreview-categories"
    },
    ingresses: {
        categoriesGRPC: new cellery:GRPCIngress(
                            8000,
                            8000
        )
    }
};

cellery:CellImage productCell = new();

public function build(string orgName, string imageName, string imageVersion) {
    io:println("Building product cell ...");

    // Map component parameters
    customers.parameters.CATEGORIES_HOST.value = cellery:getHost(untaint imageName, categories);

    // Add components to Cell
    productCell.addComponent(categories);
    productCell.addComponent(products);
    productCell.addComponent(customers);

    // Expose API from Cell Gateway
    productCell.exposeLocal(customers);
    productCell.exposeLocal(products);
    productCell.exposeLocal(categories);

    _ = cellery:createImage(productCell, orgName, imageName, imageVersion);
}

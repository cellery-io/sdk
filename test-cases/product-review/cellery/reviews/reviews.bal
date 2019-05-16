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

////Reviews Component
cellery:Component reviewsComponent = {
    name: "reviews",
    source: {
        image: "celleryio/samples-productreview-reviews"
    },
    ingresses: {
        controller: <cellery:HttpApiIngress>{
            port: 8080,
            context: "reviews-1",
            definition: {
                resources: [
                    {
                        path: "/*",
                        method: "GET"
                    }
                ]
            },
            expose: "global"
        }
    },
    envVars: {
        PORT: { value: 8080 },
        PRODUCTS_HOST: { value: "" },
        PRODUCTS_PORT: { value: 80 },
        PRODUCTS_CONTEXT: { value: "" },
        CUSTOMERS_HOST: { value: "" },
        CUSTOMERS_PORT: { value: 80 },
        CUSTOMERS_CONTEXT: { value: "" },
        RATINGS_HOST: { value: "" },
        RATINGS_PORT: { value: 80 },
        DATABASE_HOST: { value: "empty" },//TODO Remove host
        DATABASE_PORT: { value: 31406 },
        DATABASE_USERNAME: { value: "root" },
        DATABASE_PASSWORD: { value: "root" },
        DATABASE_NAME: { value: "reviews_db" }
    },
    dependencies: {
        customerProduct: <cellery:ImageName>{ org: "myorg", name: "products", ver: "1.0.0" },
        database : <cellery:ImageName>{ org: "myorg", name: "database", ver: "1.0.0" }
    }
};


// Rating Component
cellery:Component ratingComponent = {
    name: "ratings",
    source: {
        image: "celleryio/samples-productreview-ratings"
    },
    ingresses: {
        controller: <cellery:HttpApiIngress>{
            port: 8080,
            context: "ratings-1",
            definition: {
                resources: [
                    {
                        path: "/*",
                        method: "GET"
                    }
                ]
            },
            expose: "local"
        }
    },
    envVars: {
        PORT: { value: 8080 }
    }
};

cellery:CellImage reviewCell = {
    components: {
        reviews: reviewsComponent,
        rating: ratingComponent
    }
};

public function build(cellery:ImageName iName) returns error? {
    return cellery:createImage(reviewCell, iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    cellery:Reference customerProductRef = check cellery:getReference(instances.customerProduct);
    ComponentApi customerComp = parseApiUrl(<string>customerProductRef["customers-1_api_url"]);
    reviewsComponent.envVars.CUSTOMERS_HOST.value = customerComp.url;
    reviewsComponent.envVars.CUSTOMERS_PORT.value = customerComp.port;
    reviewsComponent.envVars.CUSTOMERS_CONTEXT.value = customerComp.path;

    ComponentApi productComp = parseApiUrl(<string>customerProductRef["products-1_api_url"]);
    reviewsComponent.envVars.PRODUCTS_HOST.value = productComp.url;
    reviewsComponent.envVars.PRODUCTS_PORT.value = productComp.port;
    reviewsComponent.envVars.PRODUCTS_CONTEXT.value = productComp.path;

    reviewsComponent.envVars.RATINGS_HOST.value = cellery:getHost(untaint iName.instanceName,
        ratingComponent);

    //TODO Add host
    cellery:Reference databaseRef = check cellery:getReference(instances.database);
    reviewsComponent.envVars.DATABASE_PORT.value = <string> databaseRef["mysql_tcp_port"];

    return cellery:createInstance(reviewCell, iName);
}

type ComponentApi record {
    string url;
    string port;
    string path?;
};

function parseApiUrl (string apiUrl) returns ComponentApi {
    string[] array = apiUrl.split (":");
    string url = array[1].replaceAll("/", "");
    string port = array[2].split("/")[0];
    string path = array[2].split("/")[1];
    return {url: url, port: port , path: path};
}

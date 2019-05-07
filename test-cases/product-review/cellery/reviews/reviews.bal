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
    name: "pet-service",
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
        PRODUCTS_PORT: { value: "" },
        PRODUCTS_CONTEXT: { value: "" },
        CUSTOMERS_HOST: { value: "" },
        CUSTOMERS_PORT: { value: "" },
        CUSTOMERS_CONTEXT: { value: "" },
        RATINGS_HOST: { value: "" },
        RATINGS_PORT: { value: "" },
        DATABASE_HOST: { value: "" },
        DATABASE_PORT: { value: "" },
        DATABASE_USERNAME: { value: "root" },
        DATABASE_PASSWORD: { value: "root" },
        DATABASE_NAME: { value: "reviews_db" }
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
    return cellery:createInstance(reviewCell, iName);
}

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

//Reviews Component
cellery:Component reviewsComponent = {
    name: "pet-service",
    source: {
        image: "celleryio/samples-productreview-reviews"
    },
    ingresses: {
        reviewsAPI: new cellery:HttpApiIngress(8080,
            "reviews-1",
            [
                {
                    path: "/*",
                    method: "GET"
                }
            ]
        )
    },
    parameters: {
        PORT: new cellery:Env(default = "8080"),
        PRODUCTS_HOST: new cellery:Env(),
        PRODUCTS_PORT: new cellery:Env(),
        PRODUCTS_CONTEXT: new cellery:Env(),
        CUSTOMERS_HOST: new cellery:Env(),
        CUSTOMERS_PORT: new cellery:Env(),
        CUSTOMERS_CONTEXT: new cellery:Env(),
        RATINGS_HOST: new cellery:Env(),
        RATINGS_PORT: new cellery:Env(),
        DATABASE_HOST: new cellery:Env(),
        DATABASE_PORT: new cellery:Env(),
        DATABASE_USERNAME: new cellery:Env(default = "root"),
        DATABASE_PASSWORD: new cellery:Env(default = "root"),
        DATABASE_NAME: new cellery:Env(default = "reviews_db")
    }
};


// Rating Component
cellery:Component ratingComponent = {
    name: "ratings",
    source: {
        image: "celleryio/samples-productreview-ratings"
    },
    ingresses: {
        ratingsAPI: new cellery:HttpApiIngress(8080,
            "reviews-1",
            [
                {
                    path: "/*",
                    method: "GET"
                }
            ]
        )
    },
    parameters: {
        PORT: new cellery:Env(default = "8080")
    }
};

cellery:CellImage reviewCell = new();

public function build(string orgName, string imageName, string imageVersion) {
    reviewCell.addComponent(reviewsComponent);
    reviewCell.addComponent(ratingComponent);
    reviewCell.exposeGlobal(reviewsComponent);
    _ = cellery:createImage(reviewCell, orgName, imageName, imageVersion);
}

//TODO: Implement run method

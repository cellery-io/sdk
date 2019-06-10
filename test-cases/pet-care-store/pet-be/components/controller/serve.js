/*
 *  Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const express = require("express");
const axios = require("axios");

const service = express();
const port = process.env.SERVICE_PORT || 3004;

const CATALOG_HOST = process.env.CATALOG_HOST;
const CATALOG_PORT = process.env.CATALOG_PORT;
const CUSTOMERS_HOST = process.env.CUSTOMER_HOST;
const CUSTOMERS_PORT = process.env.CUSTOMER_PORT;
const ORDERS_HOST = process.env.ORDER_HOST;
const ORDERS_PORT = process.env.ORDER_PORT;

const CATALOG_SERVICE_URL = "http://" + CATALOG_HOST + ":" + CATALOG_PORT;
const CUSTOMERS_SERVICE_URL = "http://" + CUSTOMERS_HOST + ":" + CUSTOMERS_PORT;
const ORDERS_SERVICE_URL = "http://" + ORDERS_HOST + ":" + ORDERS_PORT;

const CELLERY_USER_HEADER = "x-cellery-auth-subject";

const forwardedHeaders = [
    "Authorization",
    "x-request-id",
    "x-b3-traceid",
    "x-b3-spanid",
    "x-b3-parentspanid",
    "x-b3-sampled",
    "x-b3-flags",
    "x-ot-span-context"
];

service.use(express.json());

/**
 * Handle a success response from the API invocation.
 *
 * @param res The express response object
 * @param data The returned data from the API invocation
 */
const handleSuccess = (res, data) => {
    const response = {
        status: "SUCCESS"
    };
    if (data) {
        response.data = data;
    }
    res.send(response);
};

/**
 * Handle an error which occurred during the API invocation.
 *
 * @param res The express response object
 * @param message The error message
 */
const handleError = (res, message) => {
    console.log("[ERROR] " + message);
    res.status(500).send({
        status: "ERROR",
        message: message
    });
};

/**
 * Call an API.
 *
 * @param config Axios configuration to be used
 * @param req The received request object
 * @return {Promise<any>} The promise for the data fetch request
 */
const callAPI = (config, req) => new Promise((resolve, reject) => {
    if (!config.headers) {
        config.headers = {};
    }
    forwardedHeaders.forEach((header) => {
        const headerValue = req.get(header);
        if (headerValue) {
            config.headers[header] = headerValue;
        }
    });

    axios(config)
        .then((response) => {
            const responseBody = response.data;
            if (responseBody.status === "SUCCESS") {
                resolve(responseBody.data);
            } else {
                console.log("[ERROR] Failed to call API " + config.url + " using method " + config.method + " due to " +
                    responseBody.message);
                reject(new Error("Failed to fetch data"));
            }
        })
        .catch((error) => {
            console.log("[ERROR] Failed to call API " + config.url + " using method " + config.method + " due to " +
                error);
            reject(error);
        });
});

/**
 * Call an API in the Catalog Service.
 *
 * @param config The axios configuration with the endpoint as the URL
 * @param req The received request object
 * @return {Promise<any>} The promise for the data fetch request
 */
const callCatalogService = (config, req) => callAPI({
        ...config,
        url: CATALOG_SERVICE_URL + config.url
    }, req);

/**
 * Call an API in the Customers Service.
 *
 * @param config The axios configuration with the endpoint as the URL
 * @param req The received request object
 * @return {Promise<any>} The promise for the data fetch request
 */
const callCustomersService = (config, req) => callAPI({
        ...config,
        url: CUSTOMERS_SERVICE_URL + config.url
    }, req);

/**
 * Call an API in the Orders Service.
 *
 * @param config The axios configuration with the endpoint as the URL
 * @param req The received request object
 * @return {Promise<any>} The promise for the data fetch request
 */
const callOrdersService = (config, req) => callAPI({
        ...config,
        url: ORDERS_SERVICE_URL + config.url
    }, req);

/*
 * API endpoint for getting a list of accessories available in the catalog.
 */
service.get("/catalog", (req, res) => {
    const config = {
        url: "/accessories",
        method: "GET"
    };
    callCatalogService(config, req)
        .then((data) => {
            handleSuccess(res, {
                accessories: data
            });
        })
        .catch((error) => {
            handleError(res, "Failed to fetch catalog data due to " + error);
        });
});

/*
 * API endpoint for getting a list of accessories available in the catalog.
 */
service.get("/orders", (req, res) => {
    const ordersCallConfig = {
        url: "/orders",
        method: "GET"
    };
    const catalogCallConfig = {
        url: "/accessories",
        method: "GET"
    };
    Promise.all([
        callOrdersService(ordersCallConfig, req),
        callCatalogService(catalogCallConfig, req)
    ])
        .then((data) => {
            const orders = data[0];
            const accessories = data[1];
            orders.forEach((orderDatum) => {
                orderDatum.order = orderDatum.order.map((datum) => ({
                    item: accessories.find((accessory) => datum.id === accessory.id),
                    amount: datum.amount
                }));
            });
            handleSuccess(res, {
                orders: orders
            });
        })
        .catch((error) => {
            handleError(res, "Failed to fetch orders data due to " + error);
        });
});

/*
 * API endpoint for getting a list of accessories available in the catalog.
 */
service.post("/orders", (req, res) => {
    const config = {
        url: "/orders",
        method: "POST",
        data: req.body
    };
    callOrdersService(config, req)
        .then((data) => {
            handleSuccess(res, data);
        })
        .catch((error) => {
            handleError(res, "Failed to place order due to " + error);
        });
});

/*
 * API endpoint for getting the user profile.
 */
service.get("/profile", (req, res) => {
    const username = req.get(CELLERY_USER_HEADER);
    const config = {
        url: `/customers/${username}`,
        method: "GET"
    };
    callCustomersService(config, req)
        .then((data) => {
            handleSuccess(res, {
                profile: data
            });
        })
        .catch((error) => {
            handleError(res, "Failed to fetch profile due to " + error);
        });
});

/*
 * API endpoint for creating the user profile.
 */
service.post("/profile", (req, res) => {
    const username = req.get(CELLERY_USER_HEADER);
    const config = {
        url: "/customers",
        method: "POST",
        data: {
            ...req.body,
            username: username
        }
    };
    callCustomersService(config, req)
        .then(() => {
            handleSuccess(res);
        })
        .catch((error) => {
            handleError(res, "Failed to create profile due to " + error);
        });
});

/*
 * Starting the server
 */
const server = service.listen(port, () => {
    const host = server.address().address;
    const port = server.address().port;

    console.log("[INFO] Pet Store Controller Service listening at http://%s:%s", host, port);
});

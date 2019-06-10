/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const express = require("express");
const fs = require("fs");
const moment = require("moment");

const service = express();
const port = process.env.SERVICE_PORT || 3003;
const ordersDataDir = "data";
const ordersDataFile = `${ordersDataDir}/orders.json`;

const DATE_FORMAT = "DD-MM-YYYY";
const CELLERY_USER_HEADER = "x-cellery-auth-subject";

fs.mkdirSync(ordersDataDir); // eslint-disable-line no-sync
fs.writeFileSync(ordersDataFile, "[]", "utf8"); // eslint-disable-line no-sync

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
 * Handle when the requested resource was not found.
 *
 * @param res The express response object
 * @param message The error message
 */
const handleNotFound = (res, message) => {
    res.status(404).send({
        status: "NOT_FOUND",
        message: message
    });
};

/*
 * API endpoint for getting a list of orders available.
 */
service.get("/orders", (req, res) => {
    fs.readFile(ordersDataFile, "utf8", function (err, data) {
        const orders = JSON.parse(data);
        if (err) {
            handleError(res, "Failed to read data file " + ordersDataFile + " due to " + err);
        } else {
            const user = req.get(CELLERY_USER_HEADER);
            const orderList = orders.filter((order) => order.customer === user);
            handleSuccess(res, orderList);
        }
    });
});

/*
 * API endpoint for creating a new order.
 */
service.post("/orders", (req, res) => {
    fs.readFile(ordersDataFile, "utf8", function (err, data) {
        const orders = JSON.parse(data);
        if (err) {
            handleError(res, "Failed to read data file " + ordersDataFile + " due to " + err);
        } else {
            // Creating the new order data
            const user = req.get(CELLERY_USER_HEADER);
            const maxId = orders.reduce((acc, order) => order.id > acc ? order.id : acc, 0);
            orders.push({
                ...req.body,
                customer: user,
                id: maxId + 1,
                orderDate: moment().format(DATE_FORMAT)
            });

            // Creating the new order
            fs.writeFile(ordersDataFile, JSON.stringify(orders), "utf8", function (err) {
                if (err) {
                    handleError(res, "Failed to create new order due to " + err)
                } else {
                    handleSuccess(res, {
                        id: maxId
                    });
                }
            });
        }
    });
});

/*
 * API endpoint for getting a single order.
 */
service.get("/orders/:id", (req, res) => {
    fs.readFile(ordersDataFile, "utf8", function (err, data) {
        const orders = JSON.parse(data);
        if (err) {
            handleError(res, "Failed to read data file " + ordersDataFile + " due to " + err);
        } else {
            const user = req.get(CELLERY_USER_HEADER);
            let match = orders.filter((order) => order.id === req.params.id && order.customer === user);
            if (match.length === 1) {
                handleSuccess(res, match[0]);
            } else {
                handleNotFound("Order not available");
            }
        }
    });
});

/*
 * API endpoint for updating a order.
 */
service.put("/orders/:id", (req, res) => {
    fs.readFile(ordersDataFile, "utf8", function (err, data) {
        const orders = JSON.parse(data);
        if (err) {
            handleError(res, "Failed to read data file " + ordersDataFile + " due to " + err);
        } else {
            const user = req.get(CELLERY_USER_HEADER);
            const match = orders.filter((order) => order.id === req.params.id && order.customer === user);

            if (match.length === 1) {
                Object.assign(match[0], req.body);

                // Updating the order
                fs.writeFile(ordersDataFile, JSON.stringify(orders), "utf8", function (err) {
                    if (err) {
                        handleError(res, "Failed to update order " + req.params.id + " due to " + err)
                    } else {
                        handleSuccess(res);
                    }
                });
            } else {
                handleNotFound("Order not available");
            }
        }
    });
});

/*
 * API endpoint for deleting a order.
 */
service.delete("/orders/:id", (req, res) => {
    fs.readFile(ordersDataFile, "utf8", function (err, data) {
        const orders = JSON.parse(data);
        if (err) {
            handleError(res, "Failed to read data file " + ordersDataFile + " due to " + err);
        } else {
            const user = req.get(CELLERY_USER_HEADER);
            const newOrders = orders.filter((order) => order.id !== req.params.id || order.customer !== user);

            if (newOrders.length === orders.length) {
                handleNotFound("Order not available");
            } else {
                // Deleting the order
                fs.writeFile(ordersDataFile, JSON.stringify(newOrders), "utf8", function (err) {
                    if (err) {
                        handleError(res, "Failed to delete order " + req.params.id + " due to " + err)
                    } else {
                        handleSuccess(res);
                    }
                });
            }
        }
    });
});

/*
 * Starting the server
 */
const server = service.listen(port, () => {
    const host = server.address().address;
    const port = server.address().port;

    console.log("[INFO] Pet Store Orders Service listening at http://%s:%s", host, port);
});

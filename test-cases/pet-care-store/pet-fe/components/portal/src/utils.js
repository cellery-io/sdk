/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {createMuiTheme} from "@material-ui/core/styles";
import * as axios from "axios";

const generateTheme = () => createMuiTheme({
    typography: {
        useNextVariants: true
    },
    palette: {
        primary: {
            light: "#E5EAEA",
            main: "#5359e0",
            contrastText: "#FFF",
            dark: "#3c40a1"
        },
        secondary: {
            main: "#828ff5"
        }
    }
});

const renderFullPage = (css, content, initialState, basePath) => (
    `${"<!DOCTYPE html>"
    + "<html lang='en'>"
    + "<head>"
    + "<meta charset='utf-8'>"
    + "<link rel='shortcut icon' href='./app/assets/favicon.ico'/>"
    + "<title>Pet Store</title>"
    + "<script>window.__INITIAL_STATE__="}${JSON.stringify(initialState)}</script>${
        basePath ? `<script>window.__BASE_PATH__=${JSON.stringify(basePath)}</script>` : ""
    }<style id='jss-server-side'>${css}</style>`
    + "</head>"
    + "<body>"
    + `<div id='app'>${content}</div>`
    + "<script src='./app/index.bundle.js'></script>"
    + "<script src='./app/vendors~index.bundle.js'></script>"
    + "</body>"
    + "</html>"
);

/**
 * Call a API in the portal component.
 *
 * @param {Object} axiosConfig The axios config to use
 * @returns {Promise} The promise which will be resolved or rejected after the API request.
 */
const callApi = (axiosConfig) => new Promise((resolve, reject) => {
    if (!axiosConfig.headers) {
        axiosConfig.headers = {};
    }
    if (!axiosConfig.headers.Accept) {
        axiosConfig.headers.Accept = "application/json";
    }
    if (!axiosConfig.headers["Content-Type"]) {
        axiosConfig.headers["Content-Type"] = "application/json";
    }
    if (!axiosConfig.data && (axiosConfig.method === "POST" || axiosConfig.method === "PUT"
        || axiosConfig.method === "PATCH")) {
        axiosConfig.data = {};
    }
    axiosConfig.url = `${window.__BASE_PATH__}/api${axiosConfig.url}`;

    axios(axiosConfig)
        .then((response) => {
            if (response.status >= 200 && response.status < 400) {
                resolve(response.data);
            } else {
                reject(response.data);
            }
        })
        .catch((error) => {
            if (error.response) {
                const errorResponse = error.response;
                if (errorResponse.status === 401) {
                    window.location.href = `${window.__BASE_PATH__}/sign-in`;
                }
                reject(new Error(errorResponse.data));
            } else {
                reject(error);
            }
        });
});

export {
    generateTheme,
    renderFullPage,
    callApi
};

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

/* eslint-env node */
/* eslint "import/no-extraneous-dependencies": ["error", {"devDependencies": true}] */
/* eslint "no-sync": "off"  */

const swaggerGen = require("swagger-es6");
const petStoreSwagger = require("../resources/pet-store.swagger");
const fs = require("fs");
const path = require("path");

const codeResult = swaggerGen({
    swagger: petStoreSwagger,
    moduleName: "api",
    className: "api"
});

fs.writeFileSync(path.join(__dirname, "./gen/petStoreApi.js"), codeResult);

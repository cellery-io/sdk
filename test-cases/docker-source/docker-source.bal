// Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    // Hello Component
    cellery:Component helloComponent = {
        name: "hello",
        source: {
            dockerDir: "./docker",
            tag: "sampleapp-hello"
        },
        ingresses: {
            helloAPI: <cellery:HttpApiIngress>{
                port: 9090,
                context: "/hello",
                definition: {
                    resources: [
                        {
                            path: "/sayHello",
                            method: "GET"
                        }
                    ]
                },
                expose: "global"
            }
        }
    };

    // Hello Component
    cellery:Component helloComponentx = {
        name: "hellox",
        source: {
            dockerDir: "./docker",
            tag: "sampleapp-hellox"
        },
        ingresses: {
            helloAPI: <cellery:HttpApiIngress>{
                port: 9090,
                context: "/hellox",
                definition: {
                    resources: [
                        {
                            path: "/sayHellox",
                            method: "GET"
                        }
                    ]
                },
                expose: "global"
            }
        }
    };

    cellery:CellImage helloCell = {
        components: {
            hello: helloComponent,
            hellox: helloComponentx
        }
    };
    return cellery:createImage(helloCell, untaint iName);
}


public function run(cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
    cellery:CellImage helloCell = check cellery:constructCellImage(untaint iName);
    return cellery:createInstance(helloCell, iName, instances);
}

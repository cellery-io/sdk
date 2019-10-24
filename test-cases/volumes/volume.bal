//
//  Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
//
//  WSO2 Inc. licenses this file to you under the Apache License,
//  Version 2.0 (the "License"); you may not use this file except
//  in compliance with the License.
//  You may obtain a copy of the License at
//
//  http:www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an
//  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
//  KIND, either express or implied.  See the License for the
//  specific language governing permissions and limitations
//  under the License.
//

import ballerina/io;
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //Hello World Component
    cellery:Component helloComponent = {
        name: "hello-api",
        src: {
            image: "docker.io/wso2cellery/samples-hello-world-api-hello-service"
        },
        ingresses: {
            helloApi: <cellery:HttpApiIngress>{ port: 9090,
                context: "hello",
                authenticate: false,
                apiVersion: "v1.0.0",
                definition: {
                    resources: [
                        {
                            path: "/",
                            method: "GET"
                        }
                    ]
                },
                expose: "global"
            }
        },
        volumes: {
            secret: {
                path: "/tmp/secret/",
                readOnly: false,
                volume:<cellery:NonSharedSecret>{
                    name:cellery:generateVolumeName("my-secret"),
                    data:{
                        username:"admin",
                        password:"admin"
                    }
                }
            },
            secertShared: {
                path: "/tmp/shared/secret",
                readOnly: false,
                volume:<cellery:SharedSecret>{
                    name:"my-secret-shared"
                }
            },
            config: {
                path: "/tmp/config",
                readOnly: false,
                volume:<cellery:NonSharedConfiguration>{
                    name:"my-config",
                    data:{
                        debug:"enable",
                        logs:"enable"
                    }
                }
            },
            configShared: {
                path: "/tmp/shared/config",
                readOnly: false,
                volume:<cellery:SharedConfiguration>{
                    name:"my-config-shared"
                }
            },
            volumeClaim: {
                path: "/tmp/pvc/",
                readOnly: false,
                volume:<cellery:K8sNonSharedPersistence>{
                    name:"pv1",
                    mode:"Filesystem",
                    storageClass:"slow",
                    accessMode: ["ReadWriteMany"],
                    request:"2G",
                     lookup: {
                        labels: {
                            release: "stable"
                        },
                        expressions: [{ key: "environment", operator: "In", values: ["dev", "staging"]}]
                     }
                }
            },
            volumeClaimShared: {
                path: "/tmp/pvc/shared",
                readOnly: true,
                volume:<cellery:K8sSharedPersistence>{
                    name:"pv2"
                }
            }
        }
    };


    cellery:CellImage helloCell = {
        components: {
            helloComp: helloComponent
        }
    };
    io:println("Building Hello World Cell ...");

    //Build Hello Cell
    return <@untainted> cellery:createImage(helloCell, <@untainted> iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[] | error?) {
    cellery:CellImage helloCell = check cellery:constructCellImage(<@untainted> iName);
    return <@untainted> cellery:createInstance(helloCell, iName, instances, startDependencies, shareDependencies);
}

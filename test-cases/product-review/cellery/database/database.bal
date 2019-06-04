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

public function build(cellery:ImageName iName) returns error? {
    //Build MySQL Cell
    //MySQL Component
    cellery:Component mysqlComponent = {
        name: "mysql",
        source: {
            image: "mirage20/samples-productreview-mysql"
        },
        ingresses: {
            mysqlIngress: <cellery:TCPIngress>{
                backendPort: 3306,
                gatewayPort: 31406
            }
        },
        envVars: {
            MYSQL_ROOT_PASSWORD: { value: "root" }
        }
    };

    cellery:CellImage mysqlCell = {
        components: {
            mysqlComp: mysqlComponent
        }
    };
    return cellery:createImage(mysqlCell, untaint iName);
}

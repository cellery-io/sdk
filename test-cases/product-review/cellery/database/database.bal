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

//MySQL Component
cellery:Component mysqlComponent = {
    name: "mysql",
    source: {
        image: "mirage20/samples-productreview-mysql"
    },
    ingresses: {
        mysqlIngress: new cellery:TCPIngress(3306, 31406)
    },
    parameters: {
        MYSQL_ROOT_PASSWORD: new cellery:Env(default = "root")
    }
};

cellery:CellImage mysqlCell = new();

public function build(string orgName, string imageName, string imageVersion) {
    //Build MySQL Cell
    io:println("Building MySQL Cell ...");
    mysqlCell.addComponent(mysqlComponent);
    //Expose API from Cell Gateway
    mysqlCell.exposeLocal(mysqlComponent);
    _ = cellery:createImage(mysqlCell, orgName, imageName, imageVersion);
}

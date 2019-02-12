#
# Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
#
# WSO2 Inc. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#
#!/usr/bin/env bash

BALLERINA_VERSION=$(ballerina version | awk '{print $2}')

mvn clean install

sudo cp target/cellery-0.1.0-SNAPSHOT.jar /Library/Ballerina/ballerina-${BALLERINA_VERSION}/bre/lib
sudo cp -r target/generated-balo/repo/celleryio /Library/Ballerina/ballerina-${BALLERINA_VERSION}/lib/repo
cp -r target/generated-balo/repo/celleryio ~/.ballerina/repo

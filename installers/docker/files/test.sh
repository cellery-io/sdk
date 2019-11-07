#!/bin/bash

# ----------------------------------------------------------------------------------
# Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
#
# WSO2 Inc. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
# ----------------------------------------------------------------------------------
#
# Generate required artifacts for ballerina-runtime docker image creation
set -e

USER_ID=${1}
TEMP_DIR=tmp
B7a_TOML_LOCATION=$TEMP_DIR/Ballerina.toml
BAL_PROJECT=target
B7a_EXECUTABLE=/usr/lib/ballerina/ballerina-1.0.3/bin/ballerina

usermod -u $USER_ID $USER
$B7a_EXECUTABLE new $BAL_PROJECT
cp -r $TEMP_DIR/* $BAL_PROJECT/
cd $BAL_PROJECT/
$B7a_EXECUTABLE test --all
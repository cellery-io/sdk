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
NEW_USER=${2}
OS=${3}
PROJECT_NAME=${4}
MODULE_NAME=${5}

if [[ "$USER_ID" != 1000 ]] && [[ "$OS" == "linux" ]]; then
    useradd -m -d /home/cellery --uid $USER_ID $NEW_USER
fi

cd /home/cellery/tmp
ballerina new $PROJECT_NAME
cd $PROJECT_NAME
ballerina add $MODULE_NAME
chown -R $USER_ID /home/cellery/tmp/$PROJECT_NAME/

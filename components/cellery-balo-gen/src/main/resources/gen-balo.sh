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
#!/bin/bash
echo "Generating Cellery .balo..."
BALLERINA_DISTRIBUTION_PATH=${1}
MODULE_CELLERY_PATH="${2}/../module-cellery/"
CELLERY_VERSION="0.5.0"
BALO_CACHE_PATH="~/.ballerina/balo_cache/celleryio/cellery/${CELLERY_VERSION}/"
BIR_CACHE_PATH="~/.ballerina/bir_cache"

EXECUTABLE="${BALLERINA_DISTRIBUTION_PATH}/bin/ballerina"
cd ${MODULE_CELLERY_PATH}
${EXECUTABLE} build -a -c
mkdir -p ${BALO_CACHE_PATH}
cp -r "${MODULE_CELLERY_PATH}/target/balo/*" ${BALO_CACHE_PATH}
rm -rf ${BIR_CACHE_PATH}


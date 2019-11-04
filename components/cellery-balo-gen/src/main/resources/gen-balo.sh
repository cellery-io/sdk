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
set -e

echo "Generating Cellery .balo"
BALLERINA_DISTRIBUTION_PATH=${1}
echo "BALLERINA DISTRIBUTION PATH: ${BALLERINA_DISTRIBUTION_PATH}"
CELLERY_MODULE_PATH="${2}/../module-cellery"
echo "CELLERY MODULE PATH: ${CELLERY_MODULE_PATH}"
CELLERY_VERSION="0.5.0"
BALO_CACHE_PATH="${HOME}/.ballerina/balo_cache/celleryio/cellery/${CELLERY_VERSION}/"
BIR_CACHE_PATH="${HOME}/.ballerina/bir_cache-${CELLERY_VERSION}"
JAR_CACHE_PATH="${HOME}/.ballerina/jar_cache-${CELLERY_VERSION}"

EXECUTABLE="${BALLERINA_DISTRIBUTION_PATH}/bin/ballerina"
cd ${CELLERY_MODULE_PATH}
${EXECUTABLE} build -a -c
mkdir -p ${BALO_CACHE_PATH}
echo ${BALO_CACHE_PATH}
rm -f "${BALO_CACHE_PATH}/cellery-2019r3-java8-0.5.0.balo"
cp -rf "${CELLERY_MODULE_PATH}/target/balo/cellery-2019r3-java8-0.5.0.balo" ${BALO_CACHE_PATH}
rm -rf ${BIR_CACHE_PATH}
rm -rf ${JAR_CACHE_PATH}


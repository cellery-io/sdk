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
#Generate WSO2 Cellery Installers for macOS.

#Configuration Variables and Parameters

#Argument validation
function printUsage() {
  echo -e "\033[1mUsage:\033[0m"
  echo "$0 [VERSION]"
  echo
  echo -e "\033[1mOptions:\033[0m"
  echo "  -h (--help)"
  echo
  echo -e "\033[1mExample::\033[0m"
  echo "$0 1.0.0"

}

if [[ "$1" == "-h" ||  "$1" == "--help" ]]; then
    printUsage
    exit 1
elif [[ -z "$1" ]]; then
    echo "Please enter the version of the Cellery distribution."
    printUsage
    exit 1
else
    echo "Cellery Version : $1"
fi

#Parameters
TARGET_DIRECTORY="target"
INST_VERSION=${1}
CELLERY_VERSION=${2}
CELLERY_VERSION_NUM="${CELLERY_VERSION/-SNAPSHOT/}"
PRODUCT="cellery"
INSTALLATION_DIRECTORY="cellery-ubuntu-x64-"${INST_VERSION}
DATE=`date +%Y-%m-%d`
TIME=`date +%H:%M:%S`
LOG_PREFIX="[$DATE $TIME]"
BINARY_SIZE="0 MB"
#k8s artifacts folder
K8S_DIRECTORY="k8s-artefacts"
RESOURCE_LOCATION=files
SUPPORTED_B7A_VERSION=${3}
BALLERINA_RUNTIME="ballerina-${SUPPORTED_B7A_VERSION}"

#Functions
go_to_dir() {
    pushd $1 >/dev/null 2>&1
}

log_info() {
    echo "${LOG_PREFIX}[INFO]" $1
}

log_warn() {
    echo "${LOG_PREFIX}[WARN]" $1
}

log_error() {
    echo "${LOG_PREFIX}[ERROR]" $1
}

deleteInstallationDirectory() {
    log_info "Cleaning $TARGET_DIRECTORY directory."
    rm -rf $TARGET_DIRECTORY

    if [[ $? != 0 ]]; then
        log_error "Failed to clean $TARGET_DIRECTORY directory" $?
        exit 1
    fi
}

createInstallationDirectory() {
    if [ -d ${TARGET_DIRECTORY} ]; then
        deleteInstallationDirectory
    fi
    mkdir $TARGET_DIRECTORY

    if [[ $? != 0 ]]; then
        log_error "Failed to create $TARGET_DIRECTORY directory" $?
        exit 1
    fi
}

# ballerina scripts
getBallerinaHome() {
    if [ -z "${HOME_BALLERINA}" ]; then
        BALLERINA_VERSION=$(ballerina version | awk '{print $2}')
        BALLERINA_DEFAULT_HOME_PREFIX="/Library/Ballerina/"
        HOME_BALLERINA=${BALLERINA_DEFAULT_HOME_PREFIX}/ballerina-${BALLERINA_VERSION}
        if [ ! -d $HOME_BALLERINA ]; then
            log_error "BALLERINA_HOME cannot be found."
            exit 1
        fi
    fi
}

buildBallerinaNatives() {
    go_to_dir ../../components/lang/
    mvn clean install
    popd >/dev/null 2>&1
}

buildCelleryCLI() {
    go_to_dir ../../
    make build-cli

    if [ $? != 0 ]; then
        log_error "Failed to build cellery CLI." $?
        exit 1
    fi
    popd >/dev/null 2>&1
}

buildDocsView() {
    go_to_dir ../../
    make build-docs-view
    popd >/dev/null 2>&1
}

copyDarwinDirectory(){
  createInstallationDirectory
  rm -rf ${TARGET_DIRECTORY}/darwin
  cp -r darwin ${TARGET_DIRECTORY}/
  chmod -R 755 ${TARGET_DIRECTORY}/darwin/scripts
  chmod -R 755 ${TARGET_DIRECTORY}/darwin/Resources
  chmod 755 ${TARGET_DIRECTORY}/darwin/Distribution
}

copyBuildDirectory() {
    chmod -R 755 target/darwin/scripts/postinstall

    sed -i '' -e 's/__VERSION__/'${INST_VERSION}'/g' ${TARGET_DIRECTORY}/darwin/Distribution
    sed -i '' -e 's/__PRODUCT__/cellery/g' ${TARGET_DIRECTORY}/darwin/Distribution
    chmod -R 755 ${TARGET_DIRECTORY}/darwin/Distribution

    sed -i '' -e 's/__VERSION__/'${INST_VERSION}'/g' ${TARGET_DIRECTORY}/darwin/Resources/*.html
    chmod -R 755 ${TARGET_DIRECTORY}/darwin/Resources/

    rm -rf ${TARGET_DIRECTORY}/darwinpkg
    mkdir -p ${TARGET_DIRECTORY}/darwinpkg

    #Copy cellery product to /Library/Cellery
    mkdir -p ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery
    cp ../../components/build/cellery ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery
    chmod -R 755 ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery

    #Copy ballerina to /Library/Ballerina
    #getBallerinaHome
    #mkdir -p ${TARGET_DIRECTORY}/darwinpkg/${HOME_BALLERINA}/bre/lib/
    mkdir -p ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/repo

    cp -R $RESOURCE_LOCATION/k8s-* ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/
    cp -R $RESOURCE_LOCATION/telepresence-* ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/

    #mkdir -p ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/bre-libs/${BALLERINA_RUNTIME}/bre/lib
    #cp ../../components/lang/target/cellery-*.jar ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/bre-libs/${BALLERINA_RUNTIME}/bre/lib/
    #cp -R ../../components/lang/target/generated-balo/repo/celleryio ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/repo

    mkdir -p ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/b7a-libs/balo_cache/celleryio/cellery/${CELLERY_VERSION_NUM}
    cp ../../components/module-cellery/target/balo/cellery-2019r3-java8-0.5.0.balo ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/b7a-libs/balo_cache/celleryio/cellery/${CELLERY_VERSION_NUM}

    mkdir -p ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/docs-view
    cp -R ../../components/docs-view/build/* ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/docs-view

    rm -rf ${TARGET_DIRECTORY}/package
    mkdir -p ${TARGET_DIRECTORY}/package
    chmod -R 755 ${TARGET_DIRECTORY}/package

    rm -rf ${TARGET_DIRECTORY}/pkg
    mkdir -p ${TARGET_DIRECTORY}/pkg
    chmod -R 755 ${TARGET_DIRECTORY}/pkg
}

function buildPackage() {
    log_info "Cellery product installer package building started.(1/3)"
    pkgbuild --identifier org.${PRODUCT}.${INST_VERSION} \
    --version ${INST_VERSION} \
    --scripts ${TARGET_DIRECTORY}/darwin/scripts \
    --root ${TARGET_DIRECTORY}/darwinpkg \
    ${TARGET_DIRECTORY}/package/cellery.pkg > /dev/null 2>&1
}

function buildProduct() {
    log_info "Cellery product installer product building started.(2/3)"
    productbuild --distribution ${TARGET_DIRECTORY}/darwin/Distribution \
    --resources ${TARGET_DIRECTORY}/darwin/Resources \
    --package-path ${TARGET_DIRECTORY}/package \
    ${TARGET_DIRECTORY}/pkg/$1 > /dev/null 2>&1
}

function signProduct() {
    log_info "Cellery product installer signing process started.(3/3)"
    mkdir -p ${TARGET_DIRECTORY}/pkg-signed
    chmod -R 755 ${TARGET_DIRECTORY}/pkg-signed

    productsign --sign "Developer ID Installer: <ADD ID HERE>" \
    ${TARGET_DIRECTORY}/pkg/$1 \
    ${TARGET_DIRECTORY}/pkg-signed/$1

    pkgutil --check-signature ${TARGET_DIRECTORY}/pkg-signed/$1
}

function createInstaller() {
    log_info "Cellery product installer generation process started.(3 Steps)"
    buildPackage
    buildProduct cellery-macos-installer-x64-${INST_VERSION}.pkg
    signProduct cellery-macos-installer-x64-${INST_VERSION}.pkg
    log_info "Cellery product installer generation process finished."
}

function createUninstaller(){
    cp uninstall.sh ${TARGET_DIRECTORY}/darwinpkg/Library/Cellery
    sed -i '' -e "s/__VERSION__/${INST_VERSION}/g" "${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/uninstall.sh"
    sed -i '' -e "s/__PRODUCT__/${PRODUCT}/g" "${TARGET_DIRECTORY}/darwinpkg/Library/Cellery/uninstall.sh"
}

setCelleryVersion() {
    sed -i '' -e "s/__CELLERY_VERSION__/${CELLERY_VERSION}/g" "darwin/scripts/postinstall"
}

#Pre-requisites
command -v mvn -v >/dev/null 2>&1 || {
    log_warn "Apache Maven was not found. Please install Maven first."
    exit 1
}
command -v ballerina >/dev/null 2>&1 || {
    log_warn "Ballerina was not found. Please install ballerina first."
    exit 1
}

#Main script
log_info "Installer generating process started."

setCelleryVersion
buildBallerinaNatives
buildCelleryCLI
buildDocsView

copyDarwinDirectory
copyBuildDirectory

createUninstaller
createInstaller

log_info "Installer generating process finished"
exit 0

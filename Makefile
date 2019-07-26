# --------------------------------------------------------------------
# Copyright (c) 2019, WSO2 Inc. (http://wso2.com) All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# -----------------------------------------------------------------------

PROJECT_ROOT := $(realpath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
PROJECT_PKG := github.com/cellery-io/sdk
GO_BUILD_DIRECTORY := $(PROJECT_ROOT)/components/build
GOFILES		= $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GIT_REVISION := $(shell git rev-parse --verify HEAD)
BALLERINA_VERSION := 0.991.0

DISTRIBUTION_VERSION ?= master
DISTRIBUTION_ARTIFACTS := https://github.com/wso2-cellery/distribution/archive/$(DISTRIBUTION_VERSION).zip
OBSERVABILITY_BUILD ?= lastSuccessfulBuild

OBSERVABILITY_LAST_BUILD := https://wso2.org/jenkins/job/cellery/job/mesh-observability/$(OBSERVABILITY_BUILD)
OBSERVABILITY_ARTIFACTS_PATH := $(OBSERVABILITY_LAST_BUILD)/artifact/components/global/*zip*
OBSERVABILITY_ARTIFACTS := global.zip

OBSERVABILITY_SIDDHI_ARTIFACTS_PATH := $(OBSERVABILITY_LAST_BUILD)/artifact/components/global/core/io.cellery.observability.siddhi.apps/target/
OBSERVABILITY_PORTAL_ARTIFACTS_PATH := $(OBSERVABILITY_LAST_BUILD)/artifact/components/global/portal/io.cellery.observability.ui/target

OBSERVABILITY_SIDDHI_ARTIFACT := global/core/io.cellery.observability.siddhi.apps/target/io.cellery.observability.siddhi.apps-*.zip
OBSERVABILITY_PORTAL_ARTIFACT := global/portal/io.cellery.observability.ui/target/io.cellery.observability.ui-*.zip

DISTRIBUTION_LAST_BUILD := https://wso2.org/jenkins/job/cellery/job/distribution/lastSuccessfulBuild
DISTRIBUTION_ARTIFACTS_PATH := $(DISTRIBUTION_LAST_BUILD)/artifact
DISTRIBUTION_K8S_ARTIFACT := k8s-artefacts.tar.gz

MAIN_PACKAGES := cli

VERSION ?= 0.4.0-SNAPSHOT
INSTALLER_VERSION ?= $(GIT_REVISION)

# Go build time flags
GO_LDFLAGS := -X $(PROJECT_PKG)/components/cli/pkg/version.buildVersion=$(VERSION)
GO_LDFLAGS += -X $(PROJECT_PKG)/components/cli/pkg/version.buildGitRevision=$(GIT_REVISION)
GO_LDFLAGS += -X $(PROJECT_PKG)/components/cli/pkg/version.buildTime=$(shell date +%Y-%m-%dT%H:%M:%S%z)

# Docker info
DOCKER_REPO ?= wso2cellery
DOCKER_IMAGE_TAG ?= $(VERSION)

all: code.format build-lang build-docs-view build-cli

.PHONY: install
install: install-lang install-cli install-docs-view

.PHONY: build-lang
build-lang:
	cd ${PROJECT_ROOT}/components; \
	mvn clean install;

.PHONY: build-cli
build-cli:
	go build -o ${GO_BUILD_DIRECTORY}/cellery -ldflags "$(GO_LDFLAGS)" -x ./components/cli/cmd/cellery

.PHONY: build-docs-view
build-docs-view:
	cd ${PROJECT_ROOT}/components/docs-view; \
	npm ci; \
	npm run build

.PHONY: install-lang
install-lang:
	cd ${PROJECT_ROOT}/components/lang; \
	bash copy-libs.sh;

.PHONY: install-cli
install-cli:
	go build -o ${GO_BUILD_DIRECTORY}/cellery -ldflags "$(GO_LDFLAGS)" -x ./components/cli/cmd/cellery; \
    sudo cp ${GO_BUILD_DIRECTORY}/cellery /usr/local/bin; \

.PHONY: copy-k8s-artefacts
copy-k8s-artefacts:
	cd ${PROJECT_ROOT}/installers; \
	mkdir -p build-artifacts && cd build-artifacts;\
	curl -LO --retry 5 $(DISTRIBUTION_ARTIFACTS); \
	unzip $(DISTRIBUTION_VERSION).zip && mv distribution-master/installer/k8s-artefacts .; \
	curl --retry 5 $(OBSERVABILITY_ARTIFACTS_PATH)/$(OBSERVABILITY_ARTIFACTS) --output $(OBSERVABILITY_ARTIFACTS); \
	unzip $(OBSERVABILITY_ARTIFACTS); \
	unzip $(OBSERVABILITY_SIDDHI_ARTIFACT) -d k8s-artefacts/observability/siddhi; \
	mkdir -p k8s-artefacts/observability/node-server/config; \
	unzip $(OBSERVABILITY_PORTAL_ARTIFACT) && cp config/* k8s-artefacts/observability/node-server/config/

.PHONY: build-ubuntu-installer
build-ubuntu-installer: cleanup-installers copy-k8s-artefacts
	cd ${PROJECT_ROOT}/installers/ubuntu-x64; \
	mkdir -p files; \
	mv ../build-artifacts/k8s-artefacts files/; \
	bash build-ubuntu-x64.sh $(INSTALLER_VERSION) $(VERSION)

.PHONY: build-mac-installer
build-mac-installer: cleanup-installers copy-k8s-artefacts
	cd ${PROJECT_ROOT}/installers/macOS-x64; \
	mkdir -p files; \
	mv ../build-artifacts/k8s-artefacts files/; \
	bash build-macos-x64.sh $(INSTALLER_VERSION) $(VERSION)

.PHONY: cleanup-installers
cleanup-installers:
	rm -rf ${PROJECT_ROOT}/installers/build-artifacts; \
	rm -rf ${PROJECT_ROOT}/installers/macOS-x64/files; \
	rm -rf ${PROJECT_ROOT}/installers/macOS-x64/target; \
	rm -rf ${PROJECT_ROOT}/installers/ubuntu-x64/files; \
	rm -rf ${PROJECT_ROOT}/installers/ubuntu-x64/target

.PHONY: docker
docker:
	cd ${PROJECT_ROOT}/installers/docker; \
	bash init.sh; \
	docker build -t $(DOCKER_REPO)/ballerina-runtime:$(DOCKER_IMAGE_TAG) .

.PHONY: docker-push
docker-push: docker
	docker push $(DOCKER_REPO)/ballerina-runtime:$(DOCKER_IMAGE_TAG)

.PHONY: install-docs-view
install-docs-view:
	cd ${PROJECT_ROOT}/components/docs-view; \
	bash install-dev.sh

.PHONY: code.format
code.format: tools.goimports
	@goimports -local $(PROJECT_PKG) -w -l $(GOFILES)
	cd ${PROJECT_ROOT}/components/docs-view; \
	npm ci; \
	npm run lint

.PHONY: tools tools.goimports

tools: tools.goimports

tools.goimports:
	@command -v goimports >/dev/null; if [ $$? -ne 0 ]; then \
		echo "goimports not found. Running 'go get golang.org/x/tools/cmd/goimports'"; \
		go get golang.org/x/tools/cmd/goimports; \
		fi;

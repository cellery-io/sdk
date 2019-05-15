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
BALLERINA_VERSION := 0.990.3
BALLERINA_DIST_LOCATION := $(PROJECT_ROOT)/ballerina-$(BALLERINA_VERSION).zip
BALLERINA_JRE_LOCATION := ballerina-$(BALLERINA_VERSION)/bre/lib
BALLERINA_BIN_LOCATION := ballerina-$(BALLERINA_VERSION)/bin

OBSERVABILITY_LAST_BUILD := https://wso2.org/jenkins/job/cellery/job/mesh-observability/lastSuccessfulBuild
OBSERVABILITY_ARTIFACTS_PATH := $(OBSERVABILITY_LAST_BUILD)/artifact/components/global/*zip*
OBSERVABILITY_ARTIFACTS := global.zip

OBSERVABILITY_SIDDHI_ARTIFACTS_PATH := $(OBSERVABILITY_LAST_BUILD)/artifact/components/global/core/io.cellery.observability.siddhi.apps/target/
OBSERVABILITY_PORTAL_ARTIFACTS_PATH := $(OBSERVABILITY_LAST_BUILD)/artifact/components/global/portal/io.cellery.observability.ui/target

OBSERVABILITY_SIDDHI_ARTIFACT := global/core/io.cellery.observability.siddhi.apps/target/io.cellery.observability.siddhi.apps-*.zip
OBSERVABILITY_PORTAL_ARTIFACT := global/portal/io.cellery.observability.ui/target/io.cellery.observability.ui-*.zip

DISTRIBUTION_LAST_BUILD := https://wso2.org/jenkins/job/cellery/job/distribution/lastSuccessfulBuild
DISTRIBUTION_ARTIFACTS_PATH := $(DISTRIBUTION_LAST_BUILD)/artifact
DISTRIBUTION_K8S_ARTIFACT := k8s-artefacts.tar.gz

JRE_PATH ?= $(PROJECT_ROOT)/jre1.8.0_202

MAIN_PACKAGES := cli

VERSION ?= $(GIT_REVISION)

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
	cd ${PROJECT_ROOT}/components/cli; \
	bash build.sh;

.PHONY: copy-k8s-artefacts
copy-k8s-artefacts:
	cd ${PROJECT_ROOT}/installers; \
	curl --retry 5 $(DISTRIBUTION_ARTIFACTS_PATH)/$(DISTRIBUTION_K8S_ARTIFACT) --output $(DISTRIBUTION_K8S_ARTIFACT); \
	curl --retry 5 $(OBSERVABILITY_ARTIFACTS_PATH)/$(OBSERVABILITY_ARTIFACTS) --output $(OBSERVABILITY_ARTIFACTS); \
	tar -xvf $(DISTRIBUTION_K8S_ARTIFACT); \
	unzip $(OBSERVABILITY_ARTIFACTS); \
	unzip $(OBSERVABILITY_SIDDHI_ARTIFACT) -d k8s-artefacts/observability/siddhi; \
	mkdir -p k8s-artefacts/observability/node-server/config; \
	unzip $(OBSERVABILITY_PORTAL_ARTIFACT) && cp config/* k8s-artefacts/observability/node-server/config/

.PHONY: copy-ballerina-runtime
copy-ballerina-runtime:
	cd ${PROJECT_ROOT}/installers; \
	[ -f $(BALLERINA_DIST_LOCATION) ] || \
	curl --retry 5 https://product-dist.ballerina.io/downloads/$(BALLERINA_VERSION)/ballerina-$(BALLERINA_VERSION).zip \
	--output $(BALLERINA_DIST_LOCATION)

.PHONY: build-ubuntu-installer
build-ubuntu-installer: copy-k8s-artefacts copy-ballerina-runtime
	cd ${PROJECT_ROOT}/installers/ubuntu-x64; \
	mkdir -p files; \
	cp -r ../k8s-artefacts files/; \
	unzip $(BALLERINA_DIST_LOCATION) -d files; \
	chmod -R a+rx files/$(BALLERINA_JRE_LOCATION); \
	cp -r $(JRE_PATH) files/$(BALLERINA_JRE_LOCATION); \
	cp resources/ballerina files/$(BALLERINA_BIN_LOCATION); \
	bash build-ubuntu-x64.sh $(VERSION)

.PHONY: build-mac-installer
build-mac-installer: copy-k8s-artefacts copy-ballerina-runtime
	cd ${PROJECT_ROOT}/installers/macOS-x64; \
	mkdir -p files; \
	cp -r ../k8s-artefacts files/; \
	unzip $(BALLERINA_DIST_LOCATION) -d files; \
	chmod -R a+rx files/$(BALLERINA_JRE_LOCATION); \
	cp -r $(JRE_PATH) files/$(BALLERINA_JRE_LOCATION); \
	cp darwin/Resources/ballerina files/$(BALLERINA_BIN_LOCATION); \
	bash build-macos-x64.sh $(VERSION)

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

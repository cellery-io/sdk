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

OBSERVABILITY_LAST_BUILD := https://wso2.org/jenkins/job/cellery/job/mesh-observability/lastSuccessfulBuild
OBSERVABILITY_ARTIFACTS_PATH := $(OBSERVABILITY_LAST_BUILD)/artifact/components/global/core/io.cellery.observability.siddhi.apps/target/
OBSERVABILITY_SIDDHI_ARTIFACT := io.cellery.observability.siddhi.apps-0.1.1-SNAPSHOT.zip

DISTRIBUTION_LAST_BUILD := https://wso2.org/jenkins/job/cellery/job/distribution/lastSuccessfulBuild
DISTRIBUTION_ARTIFACTS_PATH := $(DISTRIBUTION_LAST_BUILD)/artifact
DISTRIBUTION_K8S_ARTIFACT := k8s-artefacts.tar.gz

MAIN_PACKAGES := cli

VERSION ?= $(GIT_REVISION)

# Go build time flags
GO_LDFLAGS := -X $(PROJECT_PKG)/components/cli/pkg/version.buildVersion=$(VERSION)
GO_LDFLAGS += -X $(PROJECT_PKG)/components/cli/pkg/version.buildGitRevision=$(GIT_REVISION)
GO_LDFLAGS += -X $(PROJECT_PKG)/components/cli/pkg/version.buildTime=$(shell date +%Y-%m-%dT%H:%M:%S%z)

all: code.format build-lang build-docs-view build-cli

.PHONY: install
install: install-lang install-cli install-docs-view

.PHONY: build-lang
build-lang:
	cd ${PROJECT_ROOT}/components/lang; \
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
	curl --retry 5 $(OBSERVABILITY_ARTIFACTS_PATH)/$(OBSERVABILITY_SIDDHI_ARTIFACT) --output $(OBSERVABILITY_SIDDHI_ARTIFACT); \
	tar -xvf $(DISTRIBUTION_K8S_ARTIFACT); \
	unzip $(OBSERVABILITY_SIDDHI_ARTIFACT) -d k8s-artefacts/observability/siddhi

.PHONY: copy-ballerina-runtime
copy-ballerina-runtime:
	cd ${PROJECT_ROOT}/installers; \
	curl --retry 5 https://product-dist.ballerina.io/downloads/$(BALLERINA_VERSION)/ballerina-$(BALLERINA_VERSION).zip \
	--output ballerina-$(BALLERINA_VERSION).zip

.PHONY: build-ubuntu-installer
build-ubuntu-installer: copy-k8s-artefacts copy-ballerina-runtime
	cd ${PROJECT_ROOT}/installers/ubuntu-x64; \
	mkdir -p files; \
	cp -r ../k8s-artefacts files/; \
	unzip ../ballerina-$(BALLERINA_VERSION).zip -d files; \
	bash build-ubuntu-x64.sh $(VERSION)

.PHONY: build-mac-installer
build-mac-installer: copy-k8s-artefacts copy-ballerina-runtime
	cd ${PROJECT_ROOT}/installers/macOS-x64; \
	mkdir -p files; \
	cp -r ../k8s-artefacts files/; \
	unzip ../ballerina-$(BALLERINA_VERSION).zip -d files; \
	bash build-macos-x64.sh $(VERSION)

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

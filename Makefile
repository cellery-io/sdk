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

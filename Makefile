#	Copyright 2022 Richard Kosegi
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

REGISTRY 		?= ghcr.io/rkosegi
TAG_PREFIX 		= v
DOCKER 			?= docker
VERSION 		= $(shell cat VERSION)
VER_PARTS   	= $(subst ., ,$(VERSION))
VER_MAJOR		= $(word 1,$(VER_PARTS))
VER_MINOR   	= $(word 2,$(VER_PARTS))
VER_PATCH   	= $(word 3,$(VER_PARTS))
VER_NEXT_PATCH  = $(VER_MAJOR).$(VER_MINOR).$(shell echo $$(($(VER_PATCH)+1)))
TAG 			= $(TAG_PREFIX)$(VERSION)
BRANCH 			= $(strip $(shell git rev-parse --abbrev-ref HEAD))
ARCH 			= $(shell go env GOARCH)
BUILD_DATE 		= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT 		= $(shell git rev-parse --short HEAD)
PKG 			= github.com/prometheus/common
OS 				= $(shell uname -s | tr A-Z a-z)
IMAGE 			= $(REGISTRY)/tuya-smartplug-exporter
LDFLAGS			= -s -w
LDFLAGS			+= -X ${PKG}/version.Version=v${VERSION}
LDFLAGS			+= -X ${PKG}/version.Revision=${GIT_COMMIT}
LDFLAGS			+= -X ${PKG}/version.Branch=${BRANCH}
LDFLAGS			+= -X ${PKG}/version.BuildUser=$(shell id -u -n)@$(shell hostname)
LDFLAGS			+= -X ${PKG}/version.BuildDate=${BUILD_DATE}

.DEFAULT_GOAL := build-local

bump-patch-version:
	@echo Current: $(VERSION)
	@echo Next: $(VER_NEXT_PATCH)
	@echo "$(VER_NEXT_PATCH)" > VERSION
	sed -i 's/^appVersion: .*/appVersion: v$(VER_NEXT_PATCH)/g' chart/Chart.yaml
	sed -i 's/^version: .*/version: $(VER_NEXT_PATCH)/g' chart/Chart.yaml
	git add -- VERSION Makefile chart/Chart.yaml
	git commit -sm "Bump version to v$(VER_NEXT_PATCH)"

git-tag:
	git tag -am "Release v$(VERSION)" v$(VERSION)

new-release: bump-patch-version git-tag

update-go-deps:
	@for m in $$(go list -mod=readonly -m -f '{{ if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' all); do \
		go get -d $$m; \
	done
	go mod tidy

gen-docs:
	cd chart && frigate gen . > README.md

lint:
	pre-commit run --all-files

test:
	go test -v ./...

build-docker:
	docker build -t "$(IMAGE):v$(VERSION)" \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		.

build-local:
	go fmt ./pkg/...
	go mod tidy
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o exporter ./pkg/cmd/main.go


.PHONY: build-local build-docker clean lint test gen-docs update-go-deps bump-patch-version

include func.mk

# Image URL to use all building/pushing image targets
IMG ?= quay.io/opendatahub/odh-project-controller
TAG ?= $(shell git describe --tags --always)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.26

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail

.PHONY: all
all: tools test build

##@ General
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: generate
generate: tools ## Generates required resources for the controller to work properly (see config/ folder)
	controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(call fetch-external-crds,github.com/kuadrant/authorino,api/v1beta1)

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: generate fmt vet ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(LOCALBIN)/setup-envtest use $(ENVTEST_K8S_VERSION) -p path)" ginkgo -r -vet=off --junit-report=ginkgo-test-results.xml ${args}

##@ Build
GOOS?=$(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH?=$(shell uname -m | tr '[:upper:]' '[:lower:]' | sed 's/x86_64/amd64/')
GOBUILD:=GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0

.PHONY: deps
deps:
	go mod download && go mod tidy

.PHONY: build
build: deps generate fmt vet ## Build manager binary.
	${GOBUILD} go build -o bin/manager main.go

.PHONY: run
run: generate fmt vet ## Run a controller from your host.
	go run ./main.go

##@ Container images

CONTAINER_ENGINE ?= podman

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	${CONTAINER_ENGINE} build . -t ${IMG}:${TAG}

.PHONY: docker-push
docker-push: docker-build ## Push docker image with the manager.
	${CONTAINER_ENGINE} push ${IMG}:${TAG}

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: deploy
deploy: generate ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && kustomize edit set image controller=${IMG}
	kubectl apply -k config/base

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	kubectl delete --ignore-not-found=$(ignore-not-found) -k config/base

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(shell	mkdir -p $(LOCALBIN))
PATH:=$(LOCALBIN):$(PATH)

.PHONY: tools
tools: deps
tools: $(LOCALBIN)/controller-gen $(LOCALBIN)/kustomize ## Installs required tools in local ./bin folder
tools: $(LOCALBIN)/setup-envtest $(LOCALBIN)/ginkgo

KUSTOMIZE_VERSION ?= v5.0.1
$(LOCALBIN)/kustomize:
	$(call header,"Installing $(notdir $@)")
	wget -q -c https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F$(KUSTOMIZE_VERSION)/kustomize_$(KUSTOMIZE_VERSION)_$(GOOS)_$(GOARCH).tar.gz -O /tmp/kustomize.tar.gz
	tar xzvf /tmp/kustomize.tar.gz -C $(LOCALBIN)
	chmod +x $(LOCALBIN)/kustomize

CONTROLLER_TOOLS_VERSION?=$(call go-mod-version,'controller-tools')
$(LOCALBIN)/controller-gen:
	$(call header,"Installing $(notdir $@)")
	$(call go-get-tool,controller-gen,sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION))

$(LOCALBIN)/setup-envtest:
	$(call header,"Installing $(notdir $@)")
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

$(LOCALBIN)/ginkgo:
	$(call header,"Installing $(notdir $@)")
	GOBIN=$(LOCALBIN) go install -mod=readonly github.com/onsi/ginkgo/v2/ginkgo

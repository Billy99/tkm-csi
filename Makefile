VERSION ?= 0.0.1

QUAY_USER ?= billy99
TAG_CSI ?= latest
CSI_IMG ?= quay.io/$(QUAY_USER)/tkm-csi-plugin:$(TAG_CSI)

# Image building tool (docker / podman) - docker is preferred in CI
CONTAINER_BIN_PATH := $(shell which docker 2>/dev/null || which podman)
CONTAINER_TOOL ?= $(shell basename ${CONTAINER_BIN_PATH})
# CONTAINER_FLAGS

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build
.PHONY: build-csi
build-csi: ## Build the tkm-csi-plugin binary.
	go build -o bin/tkm-csi-plugin .

.PHONY: build-tkm-agent-stub
build-tkm-agent-stub: ## Build the tkm-agent-stub binary.
	go build -C test/tkm-agent-stub -o "../../bin/tkm-agent-stub" .

.PHONY: build
build:  build-csi build-tkm-agent-stub ## Build all the tkm-csi binaries.

.PHONY: build-images
build-images: ## Build tkm-csi-plugin container image.
	$(CONTAINER_TOOL) build $(CONTAINER_FLAGS)  --progress=plain -f Containerfile -t ${CSI_IMG} .

.PHONY: push-images
push-images: ## Push tkm-csi-plugin container image.
	$(CONTAINER_TOOL) push ${CSI_IMG}

.PHONY: proto
proto: ## Build the gRPC protobuf. This generates tkm-csi_grpc.pb.go and tkm-csi.pb.go.
	cd proto && \
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    tkm-csi.proto

##@ Kind Cluster Management

TKM_TAG ?= latest
OPERATOR_IMG ?= quay.io/tkm/operator:$(TKM_TAG)
AGENT_IMG ?=quay.io/tkm/agent:$(TKM_TAG)

KIND_GPU_SIM_SCRIPT := https://raw.githubusercontent.com/maryamtahhan/kind-gpu-sim/refs/heads/main/kind-gpu-sim.sh
KIND_CLUSTER_NAME ?= kind-gpu-sim
# GPU Type (either "rocm" or "nvidia")
GPU_TYPE ?= rocm

IGNORE_NOT_FOUND ?= true

#.PHONY: setup-kind
#setup-kind: ## Create a Kind GPU cluster
#	@echo "Creating Kind GPU cluster with GPU type: $(GPU_TYPE) and cluster name: $(KIND_CLUSTER_NAME)"
#	wget -qO- $(KIND_GPU_SIM_SCRIPT) | bash -s create $(GPU_TYPE) --cluster-name $(KIND_CLUSTER_NAME)
#	@echo "Kind GPU cluster $(KIND_CLUSTER_NAME) created successfully."

#.PHONY: destroy-kind
#destroy-kind: ## Delete the Kind GPU cluster
#	@echo "Deleting Kind GPU cluster: $(KIND_CLUSTER_NAME)"
#	wget -qO- $(KIND_GPU_SIM_SCRIPT) | bash -s delete --cluster-name $(KIND_CLUSTER_NAME)
#	@echo "Kind GPU cluster $(KIND_CLUSTER_NAME) deleted successfully."

#.PHONY: kind-load-images
#kind-load-images: ## Load images into the Kind cluster
#	@echo "Loading operator image into Kind cluster: $(KIND_CLUSTER_NAME)"
#	kind load docker-image ${OPERATOR_IMG} --name $(KIND_CLUSTER_NAME)
#	@echo "Loading agent image into Kind cluster: $(KIND_CLUSTER_NAME)"
#	kind load docker-image ${AGENT_IMG} --name $(KIND_CLUSTER_NAME)
#	@echo "Loading CSI image into Kind cluster: $(KIND_CLUSTER_NAME)"
#	kind load docker-image ${CSI_IMG} --name $(KIND_CLUSTER_NAME)
#	@echo "Images loaded successfully into Kind cluster: $(KIND_CLUSTER_NAME)"

#.PHONY: deploy-on-kind
#deploy-on-kind: manifests kustomize ## Deploy operator and agent to the Kind GPU cluster.
#	cd config/manager && $(KUSTOMIZE) edit set image quay.io/tkm/operator=${OPERATOR_IMG}
#	cd config/agent && $(KUSTOMIZE) edit set image quay.io/tkm/agent=${AGENT_IMG}
#	$(KUSTOMIZE) build config/kind-gpu | kubectl apply -f -
#	@echo "Deployment on Kind GPU cluster completed."


#.PHONY: undeploy-on-kind
#undeploy-on-kind: ## Undeploy operator and agent from the Kind GPU cluster.
#	@echo "Undeploying operator and agent from Kind GPU cluster: $(KIND_CLUSTER_NAME)"
#	$(KUSTOMIZE) build config/kind-gpu | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -
#	@echo "Undeployment from Kind GPU cluster $(KIND_CLUSTER_NAME) completed."

#.PHONY: run-on-kind
#run-on-kind: setup-kind kind-load-images deploy-on-kind ## Setup Kind cluster, load images, and deploy
#	@echo "Cluster created, images loaded, and agent deployed on Kind GPU cluster."

.PHONY: deploy
deploy: ## Deploy CSI Controller and Node to K8s cluster specified in ~/.kube/config.
	@echo "Deploying CSI-Driver Object."
	kubectl apply -f manifests/csi-driver.yaml
	@echo "Deploying CSI Controller RBAC."
	kubectl apply -f manifests/rbac-controller.yaml
	@echo "Deploying CSI Controller Deployment."
	kubectl apply -f manifests/controller-plugin.yaml
	@echo "Deploying CSI Node RBAC."
	kubectl apply -f manifests/rbac-node.yaml
	@echo "Deploying CSI Node Daemonset."
	kubectl apply -f manifests/node-plugin.yaml
	@echo "Add label tkm-test-node= to node kind-gpu-sim-worker."
	kubectl label node kind-gpu-sim-worker tkm-test-node=true
	@echo "Deployment of CSI to cluster completed."

.PHONY: undeploy
undeploy: undeploy-test-pod ## Undeploy CSI Controller and Node from K8s cluster specified in ~/.kube/config.
	@echo "Undeploying CSI Node Daemonset."
	kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f manifests/node-plugin.yaml
	@echo "Undeploying CSI Node RBAC."
	kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f manifests/rbac-node.yaml
	@echo "Undeploying CSI Controller Deployment."
	kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f manifests/controller-plugin.yaml
	@echo "Undeploying CSI Controller RBAC."
	kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f manifests/rbac-controller.yaml
	@echo "Undeploying CSI-Driver Object."
	kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f manifests/csi-driver.yaml
	@echo "Remove label tkm-test-node= from node kind-gpu-sim-worker."
	kubectl label node kind-gpu-sim-worker tkm-test-node-
	@echo "Deployment of CSI removed cluster completed."

.PHONY: deploy-test-pod
deploy-test-pod: ## Deploying TritonKernelCacheCluster Instance and a Test Pod.
	@echo "Deploying TritonKernelCacheCluster Instance and a Test Pod."
	kubectl apply -f examples/testpod.yaml

.PHONY: undeploy-test-pod
undeploy-test-pod: ## Deploy CSI Controller and Node to K8s cluster specified in ~/.kube/config.
	@echo "Undeploying TritonKernelCacheCluster Instance and Test Pod."
	kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f examples/testpod.yaml

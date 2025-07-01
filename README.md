# tkm-csi

CSI Driver for the Triton Kernel Manager (TKM).
TKM deploys, manages and monitors Triton Kernels, which are packaged in an OCI Image.
The TKM CSI Driver is responsible for calling Triton Cache Vault (TCV), which is a
sub-component of TKM, to unpackaged a Triton Kernel on a Node, then mount the
Kernel in a Pod.

## Baremetal Development

To test on barematal, run the steps in this section.
Current State:

* PARTIAL: TKM Agent Stub calls TKM CSI Driver, so the OCI Images are expanded on Node.
  However, this does not follow the current architecture, which is still under discussion.
* BROKEN: There is no Kubelet stub, so nothing to simulate calls to TKM CSI Driver.
  Launch with `--test` so TKM CSI Driver does not try to listen (and fail) on Kubelet socket.

### TKM CSI Driver Local Build

To build TKM CSI Driver and the TKM Agent Stub.

```bash
make build
```

### Install Triton Cache Vault (TCV)

TKM CSI Driver assumes TCV is in it's `PATH`.
Easiest way is to run `make install` from TCV.

```bash
cd $SRC_DIR/TKM/tcv/
make install
```

### Run Unit Tests

The TKM CSI Driver has a set of Unit Tests that are run in github CI.
To run locally, build the binary then run the Unit Tests.
TKM CSI Driver runs as root so it can mount files into a pod, so the Unit Tests
also need to run as privileged.

```bash
make build
sudo env "PATH=$PATH" make test
```

### Test TKM CSI Driver

First, launch the TKM CSI Driver.

* The `--test` flag is used to skip the Kubelet socket read, because it's not there.
* Based on lack hardware available, the `--nogpu` flag allows TCV to expand out Triton
  Kernel Caches from OCI Images, even if hardware is not available.

```bash
sudo ./bin/tkm-csi-plugin --test --nogpu
{"level":"info","ts":"2025-06-03T10:15:16-04:00","logger":"tkm-csi","msg":"Created a new driver:","driver":{"Client":null,"SocketFilename":"unix:///var/lib/kubelet/plugins/csi-tkm/csi.sock","NodeName":"local","Namespace":"default","TestMode":true}}
{"level":"info","ts":"2025-06-03T10:15:16-04:00","logger":"tkm-csi","msg":"Created a new Image Server:","image":{"NodeName":"local","Namespace":"default","ImagePort":":50051","TestMode":false}}
:
```

In another window, use the `tkm-agent-stub` to simulate TritonKernelCache CRs being created or deleted:

```bash
./bin/tkm-agent-stub -load -image quay.io/tkm/vector-add-cache:rocm -crdName kernel-y -namespace my-app-ns
2025/06/03 10:15:42 Response from gRPC server's LoadKernelImage function: Load Image Request Succeeded

./bin/tkm-agent-stub -unload -crdName kernel-y -namespace my-app-ns
2025/06/03 10:25:01 Response from gRPC server's UnloadKernelImage function: Unload Image Request Received
```

## Kubernetes Development

To deploy on Kubernetes, run the steps in this section.
Current State:

* PARTIAL: TKM Agent does not call TKM CSI Driver, but the `tkm-agent-stub` binary
  is included in the CSI Driver pod, so user can exec into the pod and load images.
* PARTIAL: Kubelet calls TKM CSI Driver when Test Pod is created, TKM CSI Driver
  handles the NodePublishVolume and NodeUnpublishVolume requests, bind mounts and
  unmounts the Kernel Cache directories.
  The rest of the Kubelet calls are stubbed out.

### Deploy TKM Operator

This is a temporary step.
Once TKM starts posting images to quay.io regularly, the Makefile will just pull from latest.
See the [TKM README](https://github.com/redhat-et/TKM), but here is a brief summary:

```bash
git clone https://github.com/redhat-et/TKM.git
cd $SRC_DIR/TKM/tcv/

make build
make docker-build
```

Start the GPU-Simulated Kind Cluster:

```bash
make run-on-kind
```

### Deploy TKM CSI Driver

Build the TKM CSI Driver images.
By default they build and push to `quay.io/billy99/tkm-csi-plugin:latest`.
This can be overwritten using the `QUAY_USER` environment variable.

```bash
QUAY_USER=$USER make build-images
QUAY_USER=$USER make push-images

# Edit manifests/controller-plugin.yaml and manifests/node-plugin.yaml to use QUAY_USER
# instead of billy99. This will be automated using Kustomize at a future time.
make deploy

kubectl get pods -n tkm-system -o wide
NAME                                               READY   STATUS    RESTARTS   AGE     IP            NODE
tkm-csi-controller-5cb79656b8-cdsc4                3/3     Running   0          29m     10.244.0.13   kind-gpu-sim-control-plane
tkm-csi-node-5kr8p                                 2/2     Running   0          29m     10.89.0.23    kind-gpu-sim-control-plane
tkm-csi-node-9tv97                                 2/2     Running   0          29m     10.89.0.24    kind-gpu-sim-worker
tkm-csi-node-pqtpl                                 2/2     Running   0          29m     10.89.0.25    kind-gpu-sim-worker2
tkm-operator-controller-manager-55774ff55b-8tv2z   1/1     Running   0          12d     10.244.0.5    kind-gpu-sim-control-plane
tkm-operator-tkm-agent-g72bs                       1/1     Running   0          12d     10.244.1.3    kind-gpu-sim-worker
tkm-operator-tkm-agent-gb995                       1/1     Running   0          12d     10.244.2.3    kind-gpu-sim-worker2
```

### Test TKM CSI Driver

The `make deploy` adds a label to Node `kind-gpu-sim-worker` so it's easier to debug.
So start logs on the `tkm-csi-node-xxxxx` on that node:

```bash
kubectl logs -n tkm-system -f tkm-csi-node-9tv97
```

The TKM Agent has not been updated to call CSI Driver yet (message flow is still being
designed), but the `tkm-agent-stub` used with baremetal testing has temporarily been
added to the CSI Driver Node pod.
To prepopulate the cache on a given node, exec into the node and run the `tkm-agent-stub`
CLI.
Use the same `tkm-csi-node-xxxxx` as the previous step:

```bash
kubectl exec -it -n tkm-system -c tkm-csi-node-plugin tkm-csi-node-9tv97 -- sh
sh-5.2#
sh-5.2# tkm-agent-stub -load -image quay.io/tkm/vector-add-cache:rocm -crdName flash-attention-rocm
2025/06/04 18:53:47 Response from gRPC server's LoadKernelImage function: Load Image Request Succeeded

# Command to unload, if needed
sh-5.2# tkm-agent-stub -unload -crdName flash-attention-rocm
2025/06/04 18:55:23 Response from gRPC server's UnloadKernelImage function: Unload Image Request Received
```

**NOTE:** The test pod in the next step is creating a Cluster Scoped CR.
If Namespaced Scoped CRs are desired, include `-namespace` in the commands above and
modify the Pod Spec below accordingly.

Next deploy a TritonKernelCacheCluster instance along with a test pod referencing it.

```bash
make deploy-test-pod

kubectl get pods -n tkm-system -o wide
NAME                                               READY   STATUS    RESTARTS   AGE     IP            NODE
tkm-csi-controller-5cb79656b8-cdsc4                3/3     Running   0          29m     10.244.0.13   kind-gpu-sim-control-plane
tkm-csi-node-5kr8p                                 2/2     Running   0          29m     10.89.0.23    kind-gpu-sim-control-plane
tkm-csi-node-9tv97                                 2/2     Running   0          29m     10.89.0.24    kind-gpu-sim-worker
tkm-csi-node-pqtpl                                 2/2     Running   0          29m     10.89.0.25    kind-gpu-sim-worker2
tkm-csi-test                                       1/1     Running   0          4m58s   10.244.1.6    kind-gpu-sim-worker
tkm-operator-controller-manager-55774ff55b-8tv2z   1/1     Running   0          12d     10.244.0.5    kind-gpu-sim-control-plane
tkm-operator-tkm-agent-g72bs                       1/1     Running   0          12d     10.244.1.3    kind-gpu-sim-worker
tkm-operator-tkm-agent-gb995                       1/1     Running   0          12d     10.244.2.3    kind-gpu-sim-worker2

kubectl get tritonkernelcacheclusters
NAME                   AGE
flash-attention-rocm   69s

kubectl get tritonkernelcacheclusters flash-attention-rocm -o yaml
apiVersion: tkm.io/v1alpha1
kind: TritonKernelCacheCluster
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"tkm.io/v1alpha1","kind":"TritonKernelCacheCluster","metadata":{"annotations":{},"name":"flash-attention-rocm","namespace":"default"},"spec":{"cacheImage":"quay.io/tkm/vector-add-cache:rocm","validateSignature":false}}
  creationTimestamp: "2025-06-05T13:51:44Z"
  generation: 1
  name: flash-attention-rocm
  namespace: default
  resourceVersion: "128957"
  uid: ffd5ec70-9543-4372-af0a-fd86f8901034
spec:
  cacheImage: quay.io/tkm/vector-add-cache:rocm
  validateSignature: false
status:
  conditions:
  - lastTransitionTime: "2025-06-05T13:51:44Z"
    message: 'failed to fetch image: failed to fetch image: failed to fetch image:
      Get "https://quay.io/v2/": tls: failed to verify certificate: x509: certificate
      signed by unknown authority'
    reason: ImageFetchFailed
    status: "False"
    type: Verified

kubectl exec -it -n tkm-system tkm-csi-test -- ls -al /cache/
total 0
drwxr-xr-x  6 root root 120 Jun  5 13:51 .
drwxr-xr-x. 1 root root  93 Jun  5 13:51 ..
drwxr-xr-x  2 root root  60 Jun  5 13:51 CETLGDE7YAKGU4FRJ26IM6S47TFSIUU7KWBWDR3H2K3QRNRABUCA
drwxr-xr-x  2 root root  60 Jun  5 13:51 CHN6BLIJ7AJJRKY2IETERW2O7JXTFBUD3PH2WE3USNVKZEKXG64Q
drwxr-xr-x  2 root root 180 Jun  5 13:51 MCELTMXFCSPAMZYLZ3C3WPPYYVTVR4QOYNE52X3X6FIH7Z6N6X5A
drwxr-xr-x  2 root root  60 Jun  5 13:51 c4d45c651d6ac181a78d8d2f3ead424b8b8f07dd23dc3de0a99f425d8a633fc6
```

### Cleanup

To cleanup, undeploy the test pod, the undeploy the TKM CSI Driver.

```bash
cd $SRC_DIR/tkm-csi
make undeploy-test-pod
make undeploy
```

To tear down the KIND cluster:

```bash
cd $SRC_DIR/TKM
make destroy-kind
```

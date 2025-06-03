# tkm-csi
CSI Driver for the Triton Kernel Manager (TKM).
TKM deploys, manages and monitors Triton Kernels, which are packaged in an OCI Image.
The TKM CSI Driver is responsible for calling Triton Cache Vault (TCV), which is a
sub-component of TKM, to unpackaged a Triton Kernel on a Node, then mount the
Kernel in a Pod.

## Local Development

To test locally, run the steps in this section.
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

* BROKEN: TKM Agent does not call TKM CSI Driver, so the OCI Image is not on Node.
* PARTIAL: Kubelet calls TKM CSI Driver when Test Pod is created, TKM CSI Driver
  responds with OK (Dummy Data).
  Volume is not mounted, but pod comes up.

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

The `make deploy` adds a label to Node `kind-gpu-sim-worker` so it's easier to debug.
So start logs on the `tkm-csi-node-xxxxx` on that node:

```bash
$ kubectl logs -n tkm-system -f tkm-csi-node-9tv97
```

Then deploy a TritonKernelCacheCluster instance along with a test pod referencing it:

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
```

### Cleanup

To cleanup, undeploy the test pod, the undeploy the TKM CSI Driver.

```bash
make undeploy-test-pod
make undeploy
```

To tear down the KIND cluster:

```bash
cd TKM
make destroy-kind
```

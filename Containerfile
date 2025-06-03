FROM golang:1.24 AS csi-builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY pkgs/ pkgs/
COPY proto/ proto/
COPY vendor/ vendor/
COPY Makefile Makefile

# Build the TKM CSI Driver binary.
RUN make build-csi


FROM golang:1.24 AS tcv-builder
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=1
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgpgme-dev \
    libbtrfs-dev \
    build-essential \
    git \
    libc-dev \
    libffi-dev \
    linux-headers-amd64 \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
RUN git clone https://github.com/redhat-et/TKDK.git  --depth 1

WORKDIR /workspace/TKDK/tcv
RUN pwd && ls -al

# Build TCV
RUN make build

# Use the fedora minimal image to reduce the size of the final image but still
# be able to easily install extra packages.
FROM quay.io/fedora/fedora-minimal

# Copy the binary from the builder
COPY --from=csi-builder /workspace/bin/tkm-csi-plugin /usr/sbin/.
COPY --from=tcv-builder /workspace/TKDK/tcv/_output/bin/linux_amd64/tcv /usr/sbin/.

# Run as non-root user
USER 65532:65532

ENTRYPOINT ["tkm-csi-plugin"]

FROM golang:1.24 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY pkgs/ pkgs/
COPY vendor/ vendor/
COPY Makefile Makefile

# Build the binary with CGO enabled
RUN make build

# Use the fedora minimal image to reduce the size of the final image but still
# be able to easily install extra packages.
FROM quay.io/fedora/fedora-minimal

# Copy the binary from the builder
COPY --from=builder /workspace/bin/tkm-csi-plugin .

# Run as non-root user
USER 65532:65532

ENTRYPOINT ["/tkm-csi-plugin"]

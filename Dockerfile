ARG GOLANG_VERSION=1.20
FROM golang:${GOLANG_VERSION} as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.* /workspace/odh-project-controller/

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN cd /workspace/odh-project-controller && go mod download

COPY . /workspace/odh-project-controller

WORKDIR /workspace/odh-project-controller

# Allows to pass other targets, such as go-build.
# go-build simply compiles the binary assuming all the prerequisites are provided.
# You can e.g. call `make image -e DOCKER_ARGS="--build-arg BUILD_TARGET=go-build"`
ARG BUILD_TARGET=build
## LDFLAGS are passed from Makefile to contain metadata extracted from git during the build
ARG LDFLAGS
RUN make $BUILD_TARGET -e LDFLAGS="$LDFLAGS"

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.9
WORKDIR /
COPY --from=builder /workspace/odh-project-controller/bin/manager .
ENTRYPOINT ["/manager"]

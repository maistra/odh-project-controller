ARG GOLANG_VERSION=1.20
FROM golang:${GOLANG_VERSION} as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.* /workspace/

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Ensure we have tools installed in the separate layer,
# so we don't have to re-run every time when there is just a code change
COPY func.mk .
COPY Makefile .
RUN make tools

# Copy the rest of the project
COPY . /workspace/

# Allows to pass other targets, such as go-build.
# go-build simply compiles the binary assuming all the prerequisites are provided.
# You can e.g. call `make image -e DOCKER_ARGS="--build-arg BUILD_TARGET=go-build"`
ARG BUILD_TARGET=build
## LDFLAGS are passed from Makefile to contain metadata extracted from git during the build
ARG LDFLAGS
RUN make $BUILD_TARGET -e LDFLAGS="$LDFLAGS"

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.9
WORKDIR /
COPY --from=builder /workspace/bin/manager .
ENTRYPOINT ["/manager"]

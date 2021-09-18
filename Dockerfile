FROM golang:1.15 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
ADD . .
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o topology-updater main.go

FROM regel/cardano-cli-slim:0.0.1 as cli

# libsodium
RUN ls -la /lib/

FROM debian:stable-slim
WORKDIR /
COPY --from=cli /opt/cardano-cli /bin/cardano-cli
# libsodium
COPY --from=cli /lib/ /lib/
COPY --from=builder /workspace/topology-updater .
COPY scripts/topologyUpdater.sh .
COPY scripts/poolVet.sh .

RUN apt-get update && apt-get install -y \
    curl \
    jq \
    redis \
    parallel \
 && rm -rf /var/lib/apt/lists/*

USER 65532:65532

ENTRYPOINT ["/topology-updater"]

FROM golang:1.15 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
ADD . .
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o cardano-p2p main.go

FROM regel/cardano-cli-slim:0.0.3 as cli

# libsodium
RUN ls -la /lib/

# build curl 7.79.1 to use curl flag --fail-with-body
FROM debian:stable-slim as curl

RUN apt-get update -y && \
  apt-get install -y curl libssl-dev build-essential && \
  curl -sL https://github.com/curl/curl/releases/download/curl-7_79_1/curl-7.79.1.tar.gz | tar -xz && \
  cd curl-7.79.1 && ( ./configure --with-openssl --disable-shared --prefix=/opt ; make ; make install )

FROM debian:stable-slim
WORKDIR /
COPY --from=cli /opt/cardano-cli /bin/cardano-cli
# libsodium
COPY --from=cli /lib/ /lib/
COPY --from=curl /opt/bin/curl /usr/bin/curl
COPY --from=builder /workspace/cardano-p2p .
COPY scripts/topologyUpdater.sh .
COPY scripts/poolVet.sh .

RUN apt-get update && apt-get install -y --no-install-recommends \
    jq \
    redis \
    parallel \
    ca-certificates \
 && update-ca-certificates \
 && rm -rf /var/lib/apt/lists/*

USER 65532:65532

ENTRYPOINT ["/cardano-p2p"]

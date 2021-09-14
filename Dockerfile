FROM golang:1.15 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
ADD . .
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o topology-updater main.go

# kube-dns issues in alpine images above this version
FROM alpine:3.10.9
WORKDIR /
COPY --from=builder /workspace/topology-updater .
COPY scripts/topologyUpdater.sh .
COPY scripts/poolVet.sh .

RUN apk update \
  && apk add bash curl jq redis parallel coreutils

USER 65532:65532

ENTRYPOINT ["/topology-updater"]

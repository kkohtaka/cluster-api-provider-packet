# Build the manager binary
FROM golang:1.12.4 as builder

# Copy in the go src
WORKDIR /go/src/github.com/kkohtaka/cluster-api-provider-packet
COPY cmd/    cmd/
COPY vendor/ vendor/
COPY pkg/    pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/kkohtaka/cluster-api-provider-packet/cmd/manager

# Get CA certificates
FROM alpine:3.9.3 as certs-installer
RUN apk add --update ca-certificates

# Copy the controller-manager into a thin image
FROM scratch
WORKDIR /
COPY --from=builder /go/src/github.com/kkohtaka/cluster-api-provider-packet/manager .
COPY --from=certs-installer /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT ["/manager"]

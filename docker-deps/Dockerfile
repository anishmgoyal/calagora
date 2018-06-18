FROM golang

WORKDIR /cl/src/go
ENV GOPATH /cl/src/go

# Calagora dependencies
RUN go get github.com/aws/aws-sdk-go && \
    go get github.com/lib/pq && \
    go get golang.org/x/crypto/scrypt && \
    go get golang.org/x/net/websocket

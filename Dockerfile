FROM golang:1.16.9-alpine3.14 AS build_go
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io,direct

WORKDIR /go/src/app
COPY . .
RUN \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add alpine-sdk

RUN go mod tidy && \
    go fmt ./... && \
    go vet ./... && \
    make build

FROM alpine:3.14
EXPOSE 80
COPY --from=build_go /go/src/app/bin/linux/httpServer /bin/httpServer
WORKDIR /bin
ENTRYPOINT ["./httpServer", "-alsologtostderr"]

FROM golang:1.11-alpine as builder
ARG REPO=autoreloader-go
ARG DIR=${GOPATH}/src/github.com/deliveroo/${REPO}
ADD . $DIR
WORKDIR $DIR

RUN apk add --no-cache alpine-sdk musl-dev
RUN echo "package main" > version.go \
    && echo "const version = \"$(cat VERSION)\"" >> version.go \
    && CC=$(which gcc) go build --ldflags '-linkmode external -extldflags "-static -s"' -o /${REPO}

FROM scratch
COPY --from=builder /${REPO} /

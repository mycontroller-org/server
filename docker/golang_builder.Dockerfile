# this image used in github actions
# build locally and use it as builder image
FROM quay.io/mycontroller-org/golang:1.16.0-alpine3.13

RUN mkdir /app
ADD . /app
WORKDIR /app

ARG GOPROXY
# download deps before gobuild
RUN go mod download -x
ARG TARGETOS
ARG TARGETARCH
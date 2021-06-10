FROM --platform=${BUILDPLATFORM} golang:1.16-alpine3.13 AS builder
RUN mkdir /app
ADD . /app
WORKDIR /app

ARG GOPROXY
# download deps before gobuild
RUN go mod download -x
ARG TARGETOS
ARG TARGETARCH
ENV TARGET_BUILD="gateway"
RUN scripts/build_binary.sh

FROM alpine:3.13

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ENV APP_HOME="/app" \
    DATA_HOME="/mc_home"

EXPOSE 8080

# install timzone utils
RUN apk --no-cache add tzdata

# create a user and give permission for the locations
RUN mkdir -p ${APP_HOME} && mkdir -p ${DATA_HOME}

# copy application bin file
COPY --from=builder /app/mycontroller-gateway ${APP_HOME}/mycontroller-gateway

RUN chmod +x ${APP_HOME}/mycontroller-gateway

# copy default files
COPY ./resources/sample-docker-gateway.yaml ${APP_HOME}/gateway.yaml

WORKDIR ${APP_HOME}

CMD ["/app/mycontroller-gateway", "-config", "/app/gateway.yaml"]

FROM --platform=${BUILDPLATFORM} golang:1.16-alpine3.13 AS builder
RUN mkdir /app
ADD . /app
WORKDIR /app

ARG GOPROXY
# download deps before gobuild
RUN go mod download -x
ARG TARGETOS
ARG TARGETARCH
ENV TARGET_BUILD="core"
RUN scripts/generate_binary.sh

FROM alpine:3.12.4

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ENV APP_HOME="/app" \
    DATA_HOME="/mc_home"

EXPOSE 8080

# install timzone utils
RUN apk --no-cache add tzdata

# create a user and give permission for the locations
RUN mkdir -p ${APP_HOME} && mkdir -p ${DATA_HOME}

# copy Web console files
COPY ./console-web/build /ui

# copy application bin file
COPY --from=builder /app/mycontroller-core ${APP_HOME}/mycontroller-core

RUN chmod +x ${APP_HOME}/mycontroller-core

# copy default files
COPY ./resources/sample-docker-core.yaml ${APP_HOME}/mycontroller.yaml

WORKDIR ${APP_HOME}

CMD ["/app/mycontroller-core", "-config", "/app/mycontroller.yaml"]

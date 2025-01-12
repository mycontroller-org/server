FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23-alpine3.21 AS builder
RUN mkdir /app
ADD . /app
WORKDIR /app

# "tzdata" to print the build date with timezone
RUN apk add --no-cache tzdata git

ARG GOPROXY
# download deps before gobuild
RUN go mod download -x
ARG TARGETOS
ARG TARGETARCH
ENV TARGET_BUILD="server"
RUN scripts/container_binary.sh

FROM alpine:3.21

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ENV APP_HOME="/app" \
    DATA_HOME="/mc_home"

EXPOSE 8080

# install timzone utils
RUN apk --no-cache add tzdata

# create a user and give permission for the locations
RUN mkdir -p ${APP_HOME} && mkdir -p ${DATA_HOME}

# copy Web console files
COPY ./web-console/build /ui

# copy application bin file
COPY --from=builder /app/mycontroller-server ${APP_HOME}/mycontroller-server

RUN chmod +x ${APP_HOME}/mycontroller-server

# copy default files
COPY ./resources/sample-docker-server.yaml ${APP_HOME}/mycontroller.yaml

WORKDIR ${APP_HOME}

ENTRYPOINT [ "/app/mycontroller-server" ]
CMD [ "--config", "/app/mycontroller.yaml" ]

FROM --platform=${BUILDPLATFORM} golang:1.18-alpine3.15 AS builder
RUN mkdir /app
ADD . /app
WORKDIR /app

ARG GOPROXY
# download deps before gobuild
RUN go mod download -x
ARG TARGETOS
ARG TARGETARCH
ENV TARGET_BUILD="handler"
RUN scripts/container_binary.sh

FROM alpine:3.15

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ENV APP_HOME="/app" \
    DATA_HOME="/mc_home"

EXPOSE 8080

# install timzone utils
RUN apk --no-cache add tzdata

# create a user and give permission for the locations
RUN mkdir -p ${APP_HOME} && mkdir -p ${DATA_HOME}

# copy application bin file
COPY --from=builder /app/mycontroller-handler ${APP_HOME}/mycontroller-handler

RUN chmod +x ${APP_HOME}/mycontroller-handler

# copy default files
COPY ./resources/sample-docker-handler.yaml ${APP_HOME}/handler.yaml

WORKDIR ${APP_HOME}

ENTRYPOINT [ "/app/mycontroller-handler" ]
CMD [ "-config", "/app/handler.yaml" ]

FROM --platform=${BUILDPLATFORM:-linux/amd64} quay.io/mycontroller-org/golang:1.16.0-alpine3.13 AS builder
RUN mkdir /app
ADD . /app
WORKDIR /app

ARG GOPROXY
# download deps before gobuild
RUN go mod download -x
ARG TARGETOS
ARG TARGETARCH
ENV TARGET_BUILD="all-in-one"
RUN scripts/generate_bin.sh

FROM alpine:3.13.2

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
COPY --from=builder /app/mycontroller-all-in-one ${APP_HOME}/mycontroller-all-in-one

RUN chmod +x ${APP_HOME}/mycontroller-all-in-one

# copy default files
COPY ./resources/default-all-in-one.yaml ${APP_HOME}/mycontroller.yaml

WORKDIR ${APP_HOME}

CMD ["/app/mycontroller-all-in-one", "-config", "/app/mycontroller.yaml"]

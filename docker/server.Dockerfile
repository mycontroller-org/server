FROM alpine:3.24

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV APP_HOME="/app" \
    DATA_HOME="/mc_home"

# sample-docker-server.yaml: http 8080, https_ssl 8443, https_acme 9443
EXPOSE 8080 8443 9443

RUN apk --no-cache add ca-certificates tzdata \
    && mkdir -p ${APP_HOME} ${DATA_HOME}

COPY builds/binary/${TARGETOS}-${TARGETARCH}${TARGETVARIANT}/mycontroller-server ${APP_HOME}/mycontroller-server
COPY ./resources/sample-docker-server.yaml ${APP_HOME}/mycontroller.yaml

RUN chmod +x ${APP_HOME}/mycontroller-server

WORKDIR ${APP_HOME}

ENTRYPOINT [ "/app/mycontroller-server" ]
CMD [ "--config", "/app/mycontroller.yaml" ]

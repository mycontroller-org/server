FROM alpine:3.24

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ARG BINARY_PATH=builds/binary/linux-amd64/mycontroller-gateway

ENV APP_HOME="/app" \
    DATA_HOME="/mc_home"

RUN apk --no-cache add ca-certificates tzdata \
    && mkdir -p ${APP_HOME} ${DATA_HOME}

COPY ${BINARY_PATH} ${APP_HOME}/mycontroller-gateway
COPY ./resources/sample-docker-gateway.yaml ${APP_HOME}/gateway.yaml

RUN chmod +x ${APP_HOME}/mycontroller-gateway

WORKDIR ${APP_HOME}

ENTRYPOINT [ "/app/mycontroller-gateway" ]
CMD [ "--config", "/app/gateway.yaml" ]

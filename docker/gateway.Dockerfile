FROM alpine:3.24

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV APP_HOME="/app" \
    DATA_HOME="/mc_home"

RUN apk --no-cache add ca-certificates tzdata \
    && mkdir -p ${APP_HOME} ${DATA_HOME}

COPY builds/binary/${TARGETOS}-${TARGETARCH}${TARGETVARIANT}/mycontroller-gateway ${APP_HOME}/mycontroller-gateway
COPY ./resources/sample-docker-gateway.yaml ${APP_HOME}/gateway.yaml

RUN chmod +x ${APP_HOME}/mycontroller-gateway

WORKDIR ${APP_HOME}

ENTRYPOINT [ "/app/mycontroller-gateway" ]
CMD [ "--config", "/app/gateway.yaml" ]

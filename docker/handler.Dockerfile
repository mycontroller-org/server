FROM alpine:3.24

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV APP_HOME="/app" \
    DATA_HOME="/mc_home"

RUN apk --no-cache add ca-certificates tzdata \
    && mkdir -p ${APP_HOME} ${DATA_HOME}

COPY builds/binary/${TARGETOS}-${TARGETARCH}${TARGETVARIANT}/mycontroller-handler ${APP_HOME}/mycontroller-handler
COPY ./resources/sample-docker-handler.yaml ${APP_HOME}/handler.yaml

RUN chmod +x ${APP_HOME}/mycontroller-handler

WORKDIR ${APP_HOME}

ENTRYPOINT [ "/app/mycontroller-handler" ]
CMD [ "--config", "/app/handler.yaml" ]

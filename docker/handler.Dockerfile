FROM alpine:3.24

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ARG BINARY_PATH=builds/binary/linux-amd64/mycontroller-handler

ENV APP_HOME="/app" \
    DATA_HOME="/mc_home"

RUN apk --no-cache add ca-certificates tzdata \
    && mkdir -p ${APP_HOME} ${DATA_HOME}

COPY ${BINARY_PATH} ${APP_HOME}/mycontroller-handler
COPY ./resources/sample-docker-handler.yaml ${APP_HOME}/handler.yaml

RUN chmod +x ${APP_HOME}/mycontroller-handler

WORKDIR ${APP_HOME}

ENTRYPOINT [ "/app/mycontroller-handler" ]
CMD [ "--config", "/app/handler.yaml" ]

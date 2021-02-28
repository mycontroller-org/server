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
COPY ./mycontroller-gateway ${APP_HOME}/mycontroller-gateway

RUN chmod +x ${APP_HOME}/mycontroller-gateway

# copy default files
COPY ./resources/default-gateway.yaml ${APP_HOME}/mycontroller.yaml

WORKDIR ${APP_HOME}

CMD ["/app/mycontroller-gateway", "-config", "/app/mycontroller.yaml"]

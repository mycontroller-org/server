FROM alpine:3.12

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ENV APP_HOME="/app" \
    DATA_HOME="/mc_data"

EXPOSE 8080

# create a user and give permission for the locations
RUN mkdir -p ${APP_HOME} && mkdir -p ${DATA_HOME}

# copy Web console files
COPY ./ui /ui

# copy application bin file
COPY ./mycontroller ${APP_HOME}/mycontroller

# copy default files
COPY ./default.yaml ${DATA_HOME}/mycontroller.yaml

WORKDIR ${APP_HOME}

CMD ["/app/mycontroller", "-config", "/mc_data/mycontroller.yaml"]

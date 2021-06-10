#!/bin/sh

# user command
# command 'start' starts the service, if no input default command is 'start'
# command 'stop' stops the service
USER_COMMAND=${1}

if [ -z "${USER_COMMAND}" ];then
  echo "there is no command specified, supported commands are [start, stop]"
  exit 1
fi

# Get user current location
USER_LOCATION=${PWD}
ACTUAL_LOCATION=`dirname $0`

# change the location to where exactly script is located
cd ${ACTUAL_LOCATION}

BINARY_FILE=./mycontroller-handler
CONFIG_FILE=./handler.yaml

START_COMMAND="${BINARY_FILE} -config ${CONFIG_FILE}"

MYC_PID=`ps -ef | grep "${START_COMMAND}" | grep -v grep | awk '{ print $2 }'`

if [ ${USER_COMMAND} = "start" ]; then
  if [ ! -z "$MYC_PID" ];then
    echo "there is a running instance of the MyController handler on the pid: ${MYC_PID}"
  else
    mkdir -p logs
    exec $START_COMMAND >> logs/handler.log 2>&1 &
    echo "start command issued to the MyController handler"
  fi
elif [ ${USER_COMMAND} = "stop" ]; then
  if [ ${MYC_PID} ]; then
    kill -15 ${MYC_PID}
    echo "stop command issued to the MyController handler"
  else
    echo "MyController handler is not running"
  fi
else
  echo "invalid command [${USER_COMMAND}], supported commands are [start, stop]"
fi

# back to user location
cd ${USER_LOCATION}
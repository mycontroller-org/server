#!/bin/sh

# user command
# command 'start' starts the service
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

BINARY_FILE=./mycontroller-server
CONFIG_FILE=./mycontroller.yaml

START_COMMAND="${BINARY_FILE} -config ${CONFIG_FILE}"

MYC_PID=`ps -ef | grep "${START_COMMAND}" | grep -v grep | awk '{ print $2 }'`

if [ ${USER_COMMAND} = "start" ]; then
  if [ ! -z "$MYC_PID" ];then
    echo "there is a running instance of the MyController server on the pid: ${MYC_PID}"
  else
    mkdir -p logs
    exec $START_COMMAND >> logs/mycontroller.log 2>&1 &
    echo "start command issued to the MyController server"
  fi
elif [ ${USER_COMMAND} = "stop" ]; then
  if [ ${MYC_PID} ]; then
    kill -15 ${MYC_PID}
    echo "stop command issued to the MyController server"
  else
    echo "MyController server is not running"
  fi
else
  echo "invalid command [${USER_COMMAND}], supported commands are [start, stop]"
fi

# back to user location
cd ${USER_LOCATION}
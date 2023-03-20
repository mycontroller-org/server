#!/bin/bash

# get web-console branch/tag from versions.txt
CONSOLE_WEB_VERSION=`grep web-console= versions.txt | awk -F= '{print $2}'`


echo "web-console branch/tag: ${CONSOLE_WEB_VERSION}"

# build web console
git submodule update --init --recursive
git submodule update --remote
cd web-console
git fetch --all --tags
git checkout $CONSOLE_WEB_VERSION
cd ../

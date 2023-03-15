#!/bin/bash

# get console-web branch/tag from versions.txt
CONSOLE_WEB_VERSION=`grep console-web= versions.txt | awk -F= '{print $2}'`


echo "console-web branch/tag: ${CONSOLE_WEB_VERSION}"

# build web console
git submodule update --init --recursive
git submodule update --remote
cd console-web
git fetch --all --tags
git checkout $CONSOLE_WEB_VERSION
cd ../

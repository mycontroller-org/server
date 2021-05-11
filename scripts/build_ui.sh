#!/bin/bash

BUILDS_DIR=builds

# sync the web console repository
./scripts/setup_ui.sh

# build web console
cd console-web
yarn install
CI=false yarn build
cd ../

GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
mkdir -p ${BUILDS_DIR}
cp console-web/build ${BUILDS_DIR}/web_console -r
cd ${BUILDS_DIR}
zip -r -q web_console_${GIT_BRANCH}.zip  web_console
tar czf web_console_${GIT_BRANCH}.tar.gz web_console

cd ..
# remove copied ui build files
rm ${BUILDS_DIR}/web_console -rf

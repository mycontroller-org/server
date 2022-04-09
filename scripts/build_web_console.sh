#!/bin/bash

# sync the web console repository
./scripts/setup_web_console.sh

# build web console
cd console-web
./scripts/build.sh
cd ../

# BUILDS_DIR=builds
# GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
# mkdir -p ${BUILDS_DIR}
# cp console-web/build ${BUILDS_DIR}/web_console -r
# cd ${BUILDS_DIR}
# zip -r -q web_console_${GIT_BRANCH}.zip  web_console
# tar czf web_console_${GIT_BRANCH}.tar.gz web_console
# cd ..
# remove copied ui build files
# rm ${BUILDS_DIR}/web_console -rf

# generate web console embedded assets into go file
# embed web console assets disabled, enable this if required
# go get github.com/mjibson/esc
# go install github.com/mjibson/esc
# esc -pkg assets -o cmd/server/app/web-console/actual/generated_assets.go -prefix console-web/build console-web/build

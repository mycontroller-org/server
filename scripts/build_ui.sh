#!/bin/bash

# get backend branch details
BACKEND_BRANCH=`git rev-parse --abbrev-ref HEAD`

# build web console
git submodule update --init --recursive
git submodule update --remote
cd console-web
git checkout $BACKEND_BRANCH  # sync with backend branch for webconsole
yarn install
CI=false yarn build
cd ../

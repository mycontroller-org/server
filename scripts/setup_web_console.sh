#!/bin/bash

# get backend branch details
BACKEND_BRANCH=`git symbolic-ref -q --short HEAD || git describe --tags --exact-match`

echo "Backend branch/tag: ${BACKEND_BRANCH}"

# build web console
git submodule update --init --recursive
git submodule update --remote
cd console-web
git fetch --all --tags
git checkout $BACKEND_BRANCH  # sync web console with the backend version
cd ../
